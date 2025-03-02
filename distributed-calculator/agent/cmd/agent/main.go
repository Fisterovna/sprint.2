package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	workers, _ := strconv.Atoi(os.Getenv("COMPUTING_POWER"))
	for i := 0; i < workers; i++ {
		go worker(i+1)
	}
	select {}
}

func worker(id int) {
	for {
		task := fetchTask()
		if task != nil {
			result := compute(task)
			sendResult(task.ID, result)
		}
		time.Sleep(1 * time.Second)
	}
}

func fetchTask() *Task {
	resp, err := http.Get("http://orchestrator:8080/internal/task")
	if err != nil || resp.StatusCode != 200 {
		return nil
	}
	defer resp.Body.Close()

	var response struct{ Task *Task }
	json.NewDecoder(resp.Body).Decode(&response)
	return response.Task
}

func compute(task *Task) float64 {
	time.Sleep(time.Duration(task.OperationTime) * time.Millisecond)
	// Реализация вычислений
}

func sendResult(taskID string, result float64) {
	payload := map[string]interface{}{"id": taskID, "result": result}
	jsonData, _ := json.Marshal(payload)
	http.Post("http://orchestrator:8080/internal/task", "application/json", bytes.NewBuffer(jsonData))
}
