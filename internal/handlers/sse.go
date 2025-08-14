package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) SSEHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sessionID := chi.URLParam(r, "sessionID")
	
	// Verify session exists and user is a participant
	session, err := h.sessionService.GetSessionByID(sessionID)
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}
	if session == nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Check if user is a participant
	isParticipant := false
	for _, participant := range session.Participants {
		if participant.ID == user.ID {
			isParticipant = true
			break
		}
	}

	if !isParticipant {
		http.Error(w, "Not a session participant", http.StatusForbidden)
		return
	}

	// Create SSE client and handle connection
	client := h.sseService.AddClient(sessionID, user.ID, r)
	h.sseService.HandleSSE(w, client)
}