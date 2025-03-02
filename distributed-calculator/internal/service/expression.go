package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
	
	"distributed-calculator/orchestrator/internal/parser"
	"distributed-calculator/orchestrator/internal/storage"
)

var (
	ErrInvalidExpression = errors.New("invalid expression")
	ErrNotFound          = errors.New("expression not found")
)

type ExpressionService struct {
	exprStorage  storage.ExpressionStorage
	taskStorage  storage.TaskStorage
	parser       *parser.Parser
	taskService  *TaskService
	mu           sync.RWMutex
	activeTasks  map[string]context.CancelFunc
}

func NewExpressionService(
	exprStorage storage.ExpressionStorage,
	taskStorage storage.TaskStorage,
	parser *parser.Parser,
	taskService *TaskService,
) *ExpressionService {
	return &ExpressionService{
		exprStorage: exprStorage,
		taskStorage: taskStorage,
		parser:      parser,
		taskService: taskService,
		activeTasks: make(map[string]context.CancelFunc),
	}
}

func (s *ExpressionService) Create(expression string) (string, error) {
	if err := validateExpression(expression); err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidExpression, err)
	}

	tasks, err := s.parser.Parse(expression)
	if err != nil {
		return "", err
	}

	expr := &storage.Expression{
		ID:        generateID(),
		Raw:       expression,
		Status:    storage.StatusPending,
		CreatedAt: time.Now(),
	}

	if err := s.exprStorage.Save(expr); err != nil {
		return "", fmt.Errorf("storage error: %w", err)
	}

	for _, task := range tasks {
		task.ExpressionID = expr.ID
		if err := s.taskStorage.Add(task); err != nil {
			return "", fmt.Errorf("task storage error: %w", err)
		}
	}

	go s.processExpression(expr.ID)

	return expr.ID, nil
}

func (s *ExpressionService) GetByID(id string) (*storage.Expression, error) {
	expr, err := s.exprStorage.Get(id)
	if err != nil {
		return nil, ErrNotFound
	}
	return expr, nil
}

func (s *ExpressionService) GetAll() ([]*storage.Expression, error) {
	return s.exprStorage.GetAll(), nil
}

func (s *ExpressionService) processExpression(exprID string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.mu.Lock()
	s.activeTasks[exprID] = cancel
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.activeTasks, exprID)
		s.mu.Unlock()
	}()

	s.updateStatus(exprID, storage.StatusProcessing)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			tasks, err := s.taskStorage.GetByExpression(exprID)
			if err != nil {
				s.updateStatus(exprID, storage.StatusError)
				return
			}

			allDone := true
			for _, task := range tasks {
				if task.Status != storage.StatusDone {
					allDone = false
					break
				}
			}

			if allDone {
				result, err := s.calculateFinalResult(tasks)
				if err != nil {
					s.updateStatus(exprID, storage.StatusError)
					return
				}
				s.updateResult(exprID, result)
				return
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func validateExpression(expr string) error {
	if len(expr) == 0 {
		return errors.New("empty expression")
	}
	return nil
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func (s *ExpressionService) updateStatus(id string, status storage.ExpressionStatus) {
	expr, _ := s.exprStorage.Get(id)
	expr.Status = status
	expr.UpdatedAt = time.Now()
	s.exprStorage.Save(expr)
}

func (s *ExpressionService) updateResult(id string, result float64) {
	expr, _ := s.exprStorage.Get(id)
	expr.Result = result
	expr.Status = storage.StatusDone
	expr.UpdatedAt = time.Now()
	s.exprStorage.Save(expr)
}

func (s *ExpressionService) calculateFinalResult(tasks []*storage.Task) (float64, error) {
	results := make(map[string]float64)
	
	for _, task := range tasks {
		if task.Status != storage.StatusDone {
			return 0, errors.New("unfinished tasks")
		}
		results[task.ID] = task.Result
	}

	return results["final-task-id"], nil
}
