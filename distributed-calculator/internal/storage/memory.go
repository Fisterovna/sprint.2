func (s *MemoryTaskStorage) dependenciesMet(task *storage.Task) bool {
	for _, depID := range task.Dependencies {
		depTask, exists := s.tasks[depID]
		if !exists || depTask.Status != "done" {
			return false
		}
	}
	return true
}

func (s *MemoryTaskStorage) GetNextAvailableTask() (*storage.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, task := range s.tasks {
		if task.Status == "pending" && s.dependenciesMet(task) {
			task.Status = "processing"
			return task, nil
		}
	}
	return nil, nil
}
