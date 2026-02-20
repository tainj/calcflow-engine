// internal/worker/worker.go

package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/tainj/distributed_calculator2/internal/models"
	repo "github.com/tainj/distributed_calculator2/internal/repository"
	"github.com/tainj/distributed_calculator2/internal/valueprovider"
	"github.com/tainj/distributed_calculator2/pkg/calculator"
	"github.com/tainj/distributed_calculator2/pkg/logger"
	"github.com/tainj/distributed_calculator2/pkg/messaging/kafka"
)

type Worker struct {
	exampleRepo   repo.ExampleRepository
	cacheRepo     repo.VariableRepository
	kafkaQueue    kafka.TaskQueue
	valueProvider valueprovider.Provider
	logger        logger.Logger

	// for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// for HTTP server
	httpServer *http.Server
	port       string
}

func NewWorker(
	exampleRepo repo.ExampleRepository,
	cacheRepo repo.VariableRepository,
	kafkaQueue kafka.TaskQueue,
	valueProvider valueprovider.Provider,
	logger logger.Logger,
	port string,
) *Worker {
	ctx, cancel := context.WithCancel(context.Background())
	return &Worker{
		exampleRepo:   exampleRepo,
		cacheRepo:     cacheRepo,
		kafkaQueue:    kafkaQueue,
		valueProvider: valueProvider,
		logger:        logger,
		ctx:           ctx,
		cancel:        cancel,
		port:          port,
	}
}

func (w *Worker) Start() {
	// Start HTTP server
	w.startHTTPServer()

	// Start Kafka consumer
	w.wg.Add(1)
	go w.consumeLoop()
}

func (w *Worker) consumeLoop() {
	defer w.wg.Done()
	w.logger.Info(w.ctx, "worker started, reading tasks from kafka...")

	for {
		select {
		case <-w.ctx.Done():
			w.logger.Info(w.ctx, "consume loop stopped due to shutdown")
			return
		default:
			// read message
			jsonData, message, err := w.kafkaQueue.ReadTask()
			if err != nil {
				w.logger.Error(w.ctx, "failed to read task", "error", err)
				continue
			}

			var task models.Task
			if err := json.Unmarshal(jsonData, &task); err != nil {
				w.logger.Error(w.ctx, "failed to unmarshal task", "error", err)
				continue
			}
			w.logger.Debug(w.ctx, "received task", "raw_json", string(jsonData))
			w.logger.Debug(w.ctx, "unmarshaled task", "task", fmt.Sprintf("%+v", task))

			// process task
			result, err := w.ProcessTask(w.ctx, task)

			// business errors: division by zero, syntax
			if err != nil && (err == calculator.ErrDivisionByZero || err == calculator.ErrCovertExample) {
				w.logger.Debug(w.ctx, "business error in task", "task Variable", task.Variable, "error", err)
				if errDB := w.exampleRepo.UpdateExampleWithError(w.ctx, task.ExampleID, err.Error()); errDB != nil {
					w.logger.Error(w.ctx, "failed to save error to db", "error", errDB)
					continue
				}
				if errCommit := w.kafkaQueue.Commit(message); errCommit != nil {
					w.logger.Error(w.ctx, "failed to commit after error", "error", errCommit)
					continue
				}
				continue
			}

			// infra errors: redis, network - don't commit, Kafka will retry
			if err != nil {
				w.logger.Error(w.ctx, "infra error, will retry", "error", err)
				continue
			}

			// all ok - commit
			if err := w.kafkaQueue.Commit(message); err != nil {
				w.logger.Error(w.ctx, "failed to commit message", "error", err)
				continue
			}

			// if final - save result
			if task.IsFinal {
				if err := w.handleFinalTask(w.ctx, task, result); err != nil {
					w.logger.Error(w.ctx, "failed to save final result", "error", err)
				}
			}
		}
	}
}

func (w *Worker) ProcessTask(ctx context.Context, task models.Task) (float64, error) {
	w.logger.Info(ctx, "processing task", "task", fmt.Sprintf("%+v", task))
	val1, err := w.valueProvider.Resolve(ctx, task.Num1)
	if err != nil {
		return 0, fmt.Errorf("resolve num1 (%s): %w", task.Num1, err)
	}
	val2, err := w.valueProvider.Resolve(ctx, task.Num2)
	if err != nil {
		return 0, fmt.Errorf("resolve num2 (%s): %w", task.Num2, err)
	}

	calc := calculator.NewNode(val1, val2, task.Sign)
	result, err := calc.Calculate()
	if err != nil {
		return 0, err
	}

	if err := w.cacheRepo.SetResult(ctx, task.Variable, result); err != nil {
		return 0, fmt.Errorf("save result to Redis: %w", err)
	}

	w.logger.Info(ctx, "task processed",
		"num1", val1,
		"sign", task.Sign,
		"num2", val2,
		"result", result,
		"response", task.Variable,
	)
	return result, nil
}

func (w *Worker) handleFinalTask(ctx context.Context, task models.Task, result float64) error {
	w.logger.Info(ctx, "trying to save final result", "example_id", task.ExampleID, "result", result)
	if err := w.exampleRepo.UpdateExample(ctx, task.ExampleID, result); err != nil {
		return fmt.Errorf("update example in DB: %w", err)
	}
	w.logger.Info(ctx, "final result saved", "example", task.ExampleID, "result", result)
	return nil
}

// startHTTPServer - starts /health endpoint
func (w *Worker) startHTTPServer() {
	mux := http.NewServeMux()

	// Handler for /health
	mux.HandleFunc("/health", func(rw http.ResponseWriter, r *http.Request) {
		// Add CORS headers
		rw.Header().Set("Access-Control-Allow-Origin", "*")                   // Allow all
		rw.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS") // Allowed methods
		rw.Header().Set("Access-Control-Allow-Headers", "Content-Type")       // Allowed headers

		// Handle preflight OPTIONS request
		if r.Method == "OPTIONS" {
			rw.WriteHeader(http.StatusOK)
			return
		}

		// Response to POST/GET
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(`{"status":"alive", "port":"` + w.port + `"}`))
	})

	w.httpServer = &http.Server{
		Addr:    ":" + w.port,
		Handler: mux,
	}

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		w.logger.Info(w.ctx, "health server started", "port", w.port)
		if err := w.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			w.logger.Error(w.ctx, "health server failed", "error", err)
		}
	}()
}

// Stop — graceful shutdown
func (w *Worker) Stop() {
	w.logger.Info(w.ctx, "shutting down worker...")
	w.cancel() // stop consumer

	// stop HTTP server
	if w.httpServer != nil {
		if err := w.httpServer.Shutdown(context.Background()); err != nil {
			w.logger.Error(w.ctx, "failed to shutdown http server", "error", err)
		}
	}

	w.wg.Wait()
	w.logger.Info(w.ctx, "worker stopped gracefully")
}
