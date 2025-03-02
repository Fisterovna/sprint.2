package handler

import (
	"encoding/json"
	"net/http"

	"distributed-calculator/orchestrator/internal/service"
)

type TaskHandler struct {
	taskService *service.TaskService
}

func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	task, err := h.taskService.GetNextTask()
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"task": task})
}

func (h *TaskHandler) SubmitResult(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID     string  `json:"id"`
		Result float64 `json:"result"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusUnprocessableEntity)
		return
	}

	if err := h.taskService.CompleteTask(req.ID, req.Result); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}
