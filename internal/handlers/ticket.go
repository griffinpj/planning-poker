package handlers

import (
	"net/http"
	"strconv"

	"poker-planning/internal/models"
	"poker-planning/internal/utils"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) CreateTicket(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "Only session owner can create tickets", http.StatusForbidden)
		return
	}

	title := utils.SanitizeInput(r.FormValue("title"))
	description := utils.SanitizeInput(r.FormValue("description"))

	var allErrors utils.ValidationErrors
	allErrors = append(allErrors, utils.ValidateTicketTitle(title)...)
	allErrors = append(allErrors, utils.ValidateTicketDescription(description)...)
	
	if allErrors.HasErrors() {
		utils.WriteHTMLError(w, http.StatusBadRequest, allErrors.Error())
		return
	}

	ticket, err := h.ticketService.CreateTicket(sessionID, title, description)
	if err != nil {
		http.Error(w, "Failed to create ticket", http.StatusInternalServerError)
		return
	}

	h.wsService.Broadcast(sessionID, models.SSEMessage{
		Type: "ticket-created",
		Data: ticket,
	})

	// Return success response for HTMX, redirect for regular requests
	if r.Header.Get("HX-Request") != "" {
		// Return success status - form uses hx-swap="none" so no content swapping occurs
		w.WriteHeader(http.StatusOK)
	} else {
		http.Redirect(w, r, "/session/"+sessionID, http.StatusSeeOther)
	}
}

func (h *Handler) DeleteTicket(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sessionID := chi.URLParam(r, "sessionID")
	ticketIDStr := chi.URLParam(r, "ticketID")

	ticketID, err := strconv.Atoi(ticketIDStr)
	if err != nil {
		http.Error(w, "Invalid ticket ID", http.StatusBadRequest)
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

	if session.OwnerID != user.ID {
		http.Error(w, "Only session owner can delete tickets", http.StatusForbidden)
		return
	}

	// Get ticket before deletion for broadcast
	ticket, err := h.ticketService.GetTicketByID(ticketID)
	if err != nil {
		http.Error(w, "Failed to get ticket", http.StatusInternalServerError)
		return
	}
	if ticket == nil {
		http.Error(w, "Ticket not found", http.StatusNotFound)
		return
	}

	if ticket.SessionID != sessionID {
		http.Error(w, "Ticket does not belong to this session", http.StatusBadRequest)
		return
	}

	// If this is the current ticket, clear it from the session
	if session.CurrentTicketID != nil && *session.CurrentTicketID == ticketID {
		session.CurrentTicketID = nil
		session.IsVotingActive = false
		err = h.sessionService.UpdateSession(session)
		if err != nil {
			http.Error(w, "Failed to update session", http.StatusInternalServerError)
			return
		}
	}

	err = h.ticketService.DeleteTicket(ticketID)
	if err != nil {
		http.Error(w, "Failed to delete ticket", http.StatusInternalServerError)
		return
	}

	h.wsService.Broadcast(sessionID, models.SSEMessage{
		Type: "ticket-deleted",
		Data: map[string]interface{}{
			"ticket_id": ticketID,
		},
	})

	http.Redirect(w, r, "/session/"+sessionID, http.StatusSeeOther)
}

func (h *Handler) UpdateTicket(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sessionID := chi.URLParam(r, "sessionID")
	ticketIDStr := chi.URLParam(r, "ticketID")

	ticketID, err := strconv.Atoi(ticketIDStr)
	if err != nil {
		http.Error(w, "Invalid ticket ID", http.StatusBadRequest)
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

	if session.OwnerID != user.ID {
		http.Error(w, "Only session owner can update tickets", http.StatusForbidden)
		return
	}

	ticket, err := h.ticketService.GetTicketByID(ticketID)
	if err != nil {
		http.Error(w, "Failed to get ticket", http.StatusInternalServerError)
		return
	}
	if ticket == nil {
		http.Error(w, "Ticket not found", http.StatusNotFound)
		return
	}

	if ticket.SessionID != sessionID {
		http.Error(w, "Ticket does not belong to this session", http.StatusBadRequest)
		return
	}

	// Update ticket fields
	title := r.FormValue("title")
	description := r.FormValue("description")
	
	if title != "" {
		ticket.Title = title
	}
	ticket.Description = description

	// Handle final estimate if provided
	estimateStr := r.FormValue("final_estimate")
	if estimateStr != "" {
		estimate, err := strconv.Atoi(estimateStr)
		if err == nil {
			ticket.FinalEstimate = &estimate
		}
	}

	err = h.ticketService.UpdateTicket(ticket)
	if err != nil {
		http.Error(w, "Failed to update ticket", http.StatusInternalServerError)
		return
	}

	h.wsService.Broadcast(sessionID, models.SSEMessage{
		Type: "ticket-updated",
		Data: ticket,
	})

	http.Redirect(w, r, "/session/"+sessionID, http.StatusSeeOther)
}