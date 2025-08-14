package handlers

import (
	"net/http"
	"strconv"

	"poker-planning/internal/models"
	"poker-planning/internal/utils"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) SubmitVote(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sessionID := chi.URLParam(r, "sessionID")
	voteValue := utils.SanitizeInput(r.FormValue("vote"))

	if validationErrors := utils.ValidateVoteValue(voteValue); validationErrors.HasErrors() {
		utils.WriteHTMLError(w, http.StatusBadRequest, validationErrors.Error())
		return
	}

	session, err := h.sessionService.GetSessionByID(sessionID)
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}
	if session == nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Allow voting during active voting OR after voting has ended (for vote changes)
	// Only prevent voting if no current ticket is selected
	if session.CurrentTicket == nil {
		http.Error(w, "No active ticket", http.StatusBadRequest)
		return
	}

	// Validate vote value
	validVote := false
	for _, card := range models.AllVotingCards() {
		if card == voteValue {
			validVote = true
			break
		}
	}
	if !validVote {
		http.Error(w, "Invalid vote value", http.StatusBadRequest)
		return
	}

	vote, err := h.votingService.SubmitVote(session.CurrentTicket.ID, user.ID, voteValue)
	if err != nil {
		http.Error(w, "Failed to submit vote", http.StatusInternalServerError)
		return
	}

	h.wsService.Broadcast(sessionID, models.SSEMessage{
		Type: "vote-cast",
		Data: map[string]interface{}{
			"user_id": user.ID,
			"vote":    vote,
		},
	})

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) StartVoting(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sessionID := chi.URLParam(r, "sessionID")
	
	session, err := h.sessionService.GetSessionByID(sessionID)
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}
	if session == nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	if session.OwnerID != user.ID {
		http.Error(w, "Only session owner can start voting", http.StatusForbidden)
		return
	}

	if session.CurrentTicket == nil {
		http.Error(w, "No active ticket", http.StatusBadRequest)
		return
	}

	session.IsVotingActive = true
	err = h.sessionService.UpdateSession(session)
	if err != nil {
		http.Error(w, "Failed to start voting", http.StatusInternalServerError)
		return
	}

	// Clear existing votes for this ticket
	err = h.votingService.ClearVotesForTicket(session.CurrentTicket.ID)
	if err != nil {
		http.Error(w, "Failed to clear votes", http.StatusInternalServerError)
		return
	}

	h.wsService.Broadcast(sessionID, models.SSEMessage{
		Type: "voting-started",
		Data: session.CurrentTicket,
	})

	http.Redirect(w, r, "/session/"+sessionID, http.StatusSeeOther)
}

func (h *Handler) EndVoting(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sessionID := chi.URLParam(r, "sessionID")
	
	session, err := h.sessionService.GetSessionByID(sessionID)
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}
	if session == nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	if session.OwnerID != user.ID {
		http.Error(w, "Only session owner can end voting", http.StatusForbidden)
		return
	}

	session.IsVotingActive = false
	err = h.sessionService.UpdateSession(session)
	if err != nil {
		http.Error(w, "Failed to end voting", http.StatusInternalServerError)
		return
	}

	// Get updated votes for the ticket
	votes, err := h.votingService.GetVotesForTicket(session.CurrentTicket.ID)
	if err != nil {
		http.Error(w, "Failed to get votes", http.StatusInternalServerError)
		return
	}

	h.wsService.Broadcast(sessionID, models.SSEMessage{
		Type: "voting-ended",
		Data: map[string]interface{}{
			"ticket": session.CurrentTicket,
			"votes":  votes,
		},
	})

	http.Redirect(w, r, "/session/"+sessionID, http.StatusSeeOther)
}

func (h *Handler) NextTicket(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sessionID := chi.URLParam(r, "sessionID")
	
	session, err := h.sessionService.GetSessionByID(sessionID)
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}
	if session == nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	if session.OwnerID != user.ID {
		http.Error(w, "Only session owner can advance tickets", http.StatusForbidden)
		return
	}

	// Find next ticket
	var nextTicket *models.Ticket
	if session.CurrentTicket == nil && len(session.Tickets) > 0 {
		// Start with first ticket
		nextTicket = &session.Tickets[0]
	} else if session.CurrentTicket != nil {
		// Find current ticket and get next one
		for i, ticket := range session.Tickets {
			if ticket.ID == session.CurrentTicket.ID && i+1 < len(session.Tickets) {
				nextTicket = &session.Tickets[i+1]
				break
			}
		}
	}

	if nextTicket != nil {
		session.CurrentTicketID = &nextTicket.ID
	} else {
		session.CurrentTicketID = nil
	}
	
	session.IsVotingActive = false
	err = h.sessionService.UpdateSession(session)
	if err != nil {
		http.Error(w, "Failed to advance ticket", http.StatusInternalServerError)
		return
	}

	h.wsService.Broadcast(sessionID, models.SSEMessage{
		Type: "ticket-changed",
		Data: nextTicket,
	})

	http.Redirect(w, r, "/session/"+sessionID, http.StatusSeeOther)
}

func (h *Handler) SelectTicket(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sessionID := chi.URLParam(r, "sessionID")
	ticketIDStr := chi.URLParam(r, "ticketID")
	
	session, err := h.sessionService.GetSessionByID(sessionID)
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}
	if session == nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	if session.OwnerID != user.ID {
		http.Error(w, "Only session owner can select tickets", http.StatusForbidden)
		return
	}

	// Convert ticket ID string to int
	ticketID, err := strconv.Atoi(ticketIDStr)
	if err != nil {
		http.Error(w, "Invalid ticket ID", http.StatusBadRequest)
		return
	}
	
	// Find the ticket by ID
	var selectedTicket *models.Ticket
	for _, ticket := range session.Tickets {
		if ticket.ID == ticketID {
			selectedTicket = &ticket
			break
		}
	}

	if selectedTicket == nil {
		http.Error(w, "Ticket not found", http.StatusNotFound)
		return
	}

	// Update session with selected ticket
	session.CurrentTicketID = &ticketID
	session.IsVotingActive = false
	err = h.sessionService.UpdateSession(session)
	if err != nil {
		http.Error(w, "Failed to select ticket", http.StatusInternalServerError)
		return
	}

	h.wsService.Broadcast(sessionID, models.SSEMessage{
		Type: "ticket-changed",
		Data: selectedTicket,
	})

	http.Redirect(w, r, "/session/"+sessionID, http.StatusSeeOther)
}