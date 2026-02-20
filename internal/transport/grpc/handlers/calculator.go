package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/AlekSi/pointer"
	"github.com/tainj/distributed_calculator2/internal/auth"
	"github.com/tainj/distributed_calculator2/internal/models"
	client "github.com/tainj/distributed_calculator2/pkg/api"
)

// Service - business logic interface
// to be able to mock in tests
type Service interface {
	Calculate(ctx context.Context, example *models.Example) (*models.Example, error)
	GetResult(ctx context.Context, exampleID string) (float64, error)
	Register(ctx context.Context, user *models.UserCredentials) (*models.User, error)
	Login(ctx context.Context, user *models.UserCredentials) (*models.LoginResponse, error)
	GetExamplesByUserID(ctx context.Context, userID string) ([]models.Example, error)
}

// CalculatorService — gRPC сервер
type CalculatorService struct {
	client.UnimplementedCalculatorServer
	service Service
}

// NewCalculatorService создаёт новый хендлер
func NewCalculatorService(srv Service) *CalculatorService {
	return &CalculatorService{service: srv}
}

// Calculate — обрабатывает запрос на вычисление
func (s *CalculatorService) Calculate(ctx context.Context, req *client.CalculateRequest) (*client.CalculateResponse, error) {
	// вызываем бизнес-логику
	resp, err := s.service.Calculate(ctx, &models.Example{
		Expression: req.GetExpression(),
		UserID:     auth.UserIDFromCtx(ctx), // берём user_id из контекста
	})
	if err != nil {
		return nil, fmt.Errorf("Calculate: %w", err)
	}

	// извлекаем id
	r := pointer.Get(resp)
	return &client.CalculateResponse{
		TaskId: r.ID,
	}, nil
}

// GetResult — возвращает результат по id
func (s *CalculatorService) GetResult(ctx context.Context, req *client.GetResultRequest) (*client.GetResultResponse, error) {
	taskID := req.GetTaskId()

	// получаем результат
	result, err := s.service.GetResult(ctx, taskID)
	if err != nil {
		// ошибка — кладём в oneof
		return &client.GetResultResponse{
			Result: &client.GetResultResponse_Error{
				Error: err.Error(),
			},
		}, nil
	}

	// успех — возвращаем значение
	return &client.GetResultResponse{
		Result: &client.GetResultResponse_Value{
			Value: result,
		},
	}, nil
}

// Register — регистрирует нового пользователя
func (s *CalculatorService) Register(ctx context.Context, req *client.RegisterRequest) (*client.RegisterResponse, error) {
	// передаём креды в сервис
	_, err := s.service.Register(ctx, &models.UserCredentials{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	})
	if err != nil {
		// возвращаем ошибку в теле, не как gRPC error
		return &client.RegisterResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// успех
	return &client.RegisterResponse{
		Success: true,
		Error:   "",
	}, nil
}

// Login — вход пользователя
func (s *CalculatorService) Login(ctx context.Context, req *client.LoginRequest) (*client.LoginResponse, error) {
	// передаём креды в сервис
	loginResponse, err := s.service.Login(ctx, &models.UserCredentials{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	})
	if err != nil {
		// возвращаем ошибку в теле, не как gRPC error
		return &client.LoginResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// успех
	return &client.LoginResponse{
		Success: true,
		Token:   loginResponse.Token,
		UserId:  loginResponse.UserID,
		Error:   "",
	}, nil
}

// GetAllExamples — возвращает все вычисления пользователя
func (s *CalculatorService) GetAllExamples(ctx context.Context, req *client.GetAllExamplesRequest) (*client.GetAllExamplesResponse, error) {
	// берём user_id из контекста
	userId := auth.UserIDFromCtx(ctx)

	// получаем примеры из сервиса
	resp, err := s.service.GetExamplesByUserID(ctx, userId)
	if err != nil {
		return nil, err
	}

	// преобразуем внутренние модели в gRPC
	examples := make([]*client.Example, 0)
	for _, example := range resp {
		examples = append(examples, &client.Example{
			Id:         example.ID,
			Expression: example.Expression,
			Calculated: example.Calculated,
			Result:     example.Result, // может быть nil
			Error:      example.Error,
			CreatedAt:  example.CreatedAt.Format(time.RFC3339), // нормальный формат времени
		})
	}

	// возвращаем список
	return &client.GetAllExamplesResponse{
		Examples: examples,
	}, nil
}
