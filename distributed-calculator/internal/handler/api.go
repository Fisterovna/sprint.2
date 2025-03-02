func (h *Handler) handleCreateExpression(w http.ResponseWriter, r *http.Request) {
	// ...
	if err != nil {
		switch {
		case errors.Is(err, parser.ErrInvalidExpression):
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		case errors.Is(err, parser.ErrUnsupportedOperand):
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}
	// ...
}

func (h *Handler) handleSubmitResult(w http.ResponseWriter, r *http.Request) {
	// ...
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrTaskNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}
	// ...
}
