package main

import (
	"distributed-calculator/orchestrator/internal/handler"
	"distributed-calculator/orchestrator/internal/parser"
	"distributed-calculator/orchestrator/internal/service"
	"distributed-calculator/orchestrator/internal/storage"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	// Инициализация хранилищ
	exprStorage := storage.NewMemoryExpressionStorage()
	taskStorage := storage.NewMemoryTaskStorage()

	// Создание парсера
	exprParser := parser.NewParser()

	// Конфигурация времени операций
	opTimes := map[string]int{
		"+": getEnvAsInt("TIME_ADDITION_MS", 1000),
		"-": getEnvAsInt("TIME_SUBTRACTION_MS", 1000),
		"*": getEnvAsInt("TIME_MULTIPLICATIONS_MS", 2000),
		"/": getEnvAsInt("TIME_DIVISIONS_MS", 2000),
	}

	// Инициализация сервисов
	taskSvc := service.NewTaskService(taskStorage, opTimes)
	exprSvc := service.NewExpressionService(exprStorage, taskStorage, exprParser, taskSvc)

	// HTTP обработчики
	h := handler.NewHandler(exprSvc, taskSvc)

	// Запуск сервера
	log.Println("Starting orchestrator on :8080")
	log.Fatal(http.ListenAndServe(":8080", h.InitRoutes()))
}

func getEnvAsInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}
