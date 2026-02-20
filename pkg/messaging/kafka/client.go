package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/segmentio/kafka-go"
	"github.com/tainj/distributed_calculator2/pkg/logger"
)

type TaskQueue interface {
	SendTask(task interface{}) error
	ReadTask() ([]byte, kafka.Message, error)
	Commit(message kafka.Message) error
}

type KafkaQueue struct {
	writer *kafka.Writer
	reader *kafka.Reader
	logger logger.Logger
}

type Config struct {
	BootstrapServers  string `env:"KAFKA_BOOTSTRAP_SERVERS" env-default:"localhost:9092,localhost:9094,localhost:9096"`
	TopicCalculations string `env:"KAFKA_TOPIC_CALCULATIONS" env-default:"calculator_tasks"`
	TopicResults      string `env:"KAFKA_TOPIC_RESULTS" env-default:"calculator_results"`
}

func NewKafkaQueue(cfg Config, l logger.Logger) (*KafkaQueue, error) {
	brokers := strings.Split(cfg.BootstrapServers, ",")
	l.Info(context.Background(), "initializing KafkaQueue", "brokers", brokers, "topic", cfg.TopicCalculations)
	if len(brokers) == 0 {
		return nil, fmt.Errorf("no Kafka brokers specified")
	}
	return &KafkaQueue{
		writer: &kafka.Writer{
			Addr:  kafka.TCP(brokers...),
			Topic: cfg.TopicCalculations,
		},
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        brokers,
			Topic:          cfg.TopicCalculations,
			GroupID:        "calculator_group",
			MinBytes:       1,
			MaxBytes:       1e6,
			CommitInterval: 0, // commit manually
		}),
		logger: l,
	}, nil
}

func (k *KafkaQueue) SendTask(task interface{}) error {
	jsonData, err := json.Marshal(task)
	if err != nil {
		k.logger.Error(context.Background(), "failed ot marshal task", "error", err)
		return err
	}
	k.logger.Debug(context.Background(), "send task to Kafka", "body", string(jsonData))
	err = k.writer.WriteMessages(context.Background(), kafka.Message{
		Value: jsonData,
	})
	if err != nil {
		k.logger.Error(context.Background(), "failed to send task to Kafka", "error", err)
		return err
	}
	k.logger.Debug(context.Background(), "task sent successfully")
	return nil
}

// ReadTask returns message data and message itself for commit
func (k *KafkaQueue) ReadTask() ([]byte, kafka.Message, error) {
	k.logger.Debug(context.Background(), "read task from Kafka")
	message, err := k.reader.ReadMessage(context.Background())
	if err != nil {
		k.logger.Error(context.Background(), "failed to read message from Kafka", "error", err)
		return nil, kafka.Message{}, err
	}

	k.logger.Debug(context.Background(), "received message", "message", string(message.Value), "offset", message.Offset)
	return message.Value, message, nil
}

// Commit confirms message processing
func (k *KafkaQueue) Commit(message kafka.Message) error {
	k.logger.Debug(context.Background(), "committing message", "offset", message.Offset)

	err := k.reader.CommitMessages(context.Background(), message)
	if err != nil {
		k.logger.Error(context.Background(), "failed to commit message", "error", err)
		return err
	}

	k.logger.Debug(context.Background(), "message committed successfully")
	return nil
}
