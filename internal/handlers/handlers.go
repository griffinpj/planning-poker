package handlers

import (
	"html/template"
	"net/http"

	"poker-planning/internal/models"
	"poker-planning/internal/services"
	"poker-planning/internal/utils"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	userService    *services.UserService
	sessionService *services.SessionService
	votingService  *services.VotingService
	ticketService  *services.TicketService
	wsService      *services.WSService
	templates      *template.Template
}

func NewHandler(userService *services.UserService, sessionService *services.SessionService, votingService *services.VotingService, ticketService *services.TicketService, wsService *services.WSService) *Handler {
	templates := template.Must(template.ParseGlob("templates/*.html"))
	
	return &Handler{
		userService:    userService,
		sessionService: sessionService,
		votingService:  votingService,
		ticketService:  ticketService,
		wsService:      wsService,
		templates:      templates,
	}
}

type PageData struct {
	Title           string
	Template        string
	User            *models.User
	Session         *models.Session
	SessionName     string
	VotingCards     []string
	UserVote        *models.Vote
	VoteHistogram   []VoteCount
	CurrentTicketIndex int
	TicketAverages  map[int]float64 // ticket ID -> average
	// Summary page data
	TotalVotes       int
	EstimatedTickets int
	OverallAverage   float64
	TicketVoteGroups map[int][]VoteCount // ticket ID -> vote groups
	ParticipantStats map[string]*ParticipantStat // user ID -> stats
}

type ParticipantStat struct {
	VoteCount   int
	AverageVote float64
}

type VoteCount struct {
	Value      string
	Count      int
	Percentage int
}

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	
	data := PageData{
		Title:    "Home",
		Template: "home",
		User:     user,
	}
	
	h.executeTemplate(w, "base.html", data)
}

func (h *Handler) SetUsername(w http.ResponseWriter, r *http.Request) {
	username := utils.SanitizeInput(r.FormValue("username"))
	
	if validationErrors := utils.ValidateUsername(username); validationErrors.HasErrors() {
		utils.WriteHTMLError(w, http.StatusBadRequest, validationErrors.Error())
		return
	}

	user, err := h.userService.CreateUser(username)
	if err != nil {
		utils.LogError("SetUsername", err)
		utils.WriteHTMLError(w, http.StatusInternalServerError, "Failed to create user account")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    user.ID,
		MaxAge:   6 * 3600, // 6 hours
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	// Check if there's a redirect_to parameter or referer
	redirectTo := r.FormValue("redirect_to")
	if redirectTo == "" {
		referer := r.Header.Get("Referer")
		if referer != "" && referer != r.Header.Get("Host") {
			redirectTo = referer
		}
	}
	
	if redirectTo != "" && redirectTo != "/" {
		w.Header().Set("HX-Redirect", redirectTo)
	} else {
		w.Header().Set("HX-Refresh", "true")
	}
}

func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	name := utils.SanitizeInput(r.FormValue("name"))
	
	if validationErrors := utils.ValidateSessionName(name); validationErrors.HasErrors() {
		utils.WriteHTMLError(w, http.StatusBadRequest, validationErrors.Error())
		return
	}

	session, err := h.sessionService.CreateSession(name, user.ID)
	if err != nil {
		utils.LogError("CreateSession", err)
		utils.WriteHTMLError(w, http.StatusInternalServerError, "Failed to create planning session")
		return
	}

	w.Header().Set("HX-Redirect", "/session/"+session.ID)
}

func (h *Handler) GetSessionPartial(w http.ResponseWriter, r *http.Request) {
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

	var userVote *models.Vote
	var voteHistogram []VoteCount
	var currentTicketIndex int
	
	// Calculate averages for all tickets
	ticketAverages := make(map[int]float64)
	for _, ticket := range session.Tickets {
		if len(ticket.Votes) > 0 {
			if avg := h.calculateVoteAverage(ticket.Votes); avg != nil {
				ticketAverages[ticket.ID] = *avg
			}
		}
	}

	if session.CurrentTicket != nil {
		for i, ticket := range session.Tickets {
			if ticket.ID == session.CurrentTicket.ID {
				currentTicketIndex = i + 1
				break
			}
		}

		for _, vote := range session.CurrentTicket.Votes {
			if vote.UserID == user.ID {
				userVote = &vote
				break
			}
		}

		if !session.IsVotingActive {
			voteHistogram = h.calculateVoteHistogram(session.CurrentTicket.Votes)
		}
	}

	data := PageData{
		Title:              session.Name,
		Template:           "session",
		User:               user,
		Session:            session,
		SessionName:        session.Name,
		VotingCards:        models.AllVotingCards(),
		UserVote:           userVote,
		VoteHistogram:      voteHistogram,
		CurrentTicketIndex: currentTicketIndex,
		TicketAverages:     ticketAverages,
	}

	// Return only the session content, not the full page
	h.executeTemplate(w, "session-content", data)
}

func (h *Handler) GetSession(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		// Redirect to home page with redirect_to parameter
		redirectURL := "/?redirect_to=" + r.URL.Path
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
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

	userJoined, err := h.sessionService.JoinSession(sessionID, user.ID)
	if err != nil {
		http.Error(w, "Failed to join session", http.StatusInternalServerError)
		return
	}

	// Only broadcast if user actually joined (wasn't already a participant)
	if userJoined {
		h.wsService.Broadcast(sessionID, models.SSEMessage{
			Type: "user-joined",
			Data: user,
		})
	}

	session, err = h.sessionService.GetSessionByID(sessionID)
	if err != nil {
		http.Error(w, "Failed to refresh session", http.StatusInternalServerError)
		return
	}

	var userVote *models.Vote
	var voteHistogram []VoteCount
	var currentTicketIndex int
	
	// Calculate averages for all tickets
	ticketAverages := make(map[int]float64)
	for _, ticket := range session.Tickets {
		if len(ticket.Votes) > 0 {
			if avg := h.calculateVoteAverage(ticket.Votes); avg != nil {
				ticketAverages[ticket.ID] = *avg
			}
		}
	}

	if session.CurrentTicket != nil {
		for i, ticket := range session.Tickets {
			if ticket.ID == session.CurrentTicket.ID {
				currentTicketIndex = i + 1
				break
			}
		}

		for _, vote := range session.CurrentTicket.Votes {
			if vote.UserID == user.ID {
				userVote = &vote
				break
			}
		}

		if !session.IsVotingActive {
			voteHistogram = h.calculateVoteHistogram(session.CurrentTicket.Votes)
		}
	}

	data := PageData{
		Title:              session.Name,
		Template:           "session",
		User:               user,
		Session:            session,
		SessionName:        session.Name,
		VotingCards:        models.AllVotingCards(),
		UserVote:           userVote,
		VoteHistogram:      voteHistogram,
		CurrentTicketIndex: currentTicketIndex,
		TicketAverages:     ticketAverages,
	}

	h.executeTemplate(w, "base.html", data)
}

func (h *Handler) JoinSession(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	sessionID := chi.URLParam(r, "sessionID")
	
	userJoined, err := h.sessionService.JoinSession(sessionID, user.ID)
	if err != nil {
		http.Error(w, "Failed to join session", http.StatusInternalServerError)
		return
	}

	// Only broadcast if user actually joined (wasn't already a participant)
	if userJoined {
		h.wsService.Broadcast(sessionID, models.SSEMessage{
			Type: "user-joined",
			Data: user,
		})
	}

	http.Redirect(w, r, "/session/"+sessionID, http.StatusSeeOther)
}

func (h *Handler) LeaveSession(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sessionID := chi.URLParam(r, "sessionID")
	
	err := h.sessionService.LeaveSession(sessionID, user.ID)
	if err != nil {
		http.Error(w, "Failed to leave session", http.StatusInternalServerError)
		return
	}

	h.wsService.Broadcast(sessionID, models.SSEMessage{
		Type: "user-left",
		Data: user,
	})

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteSession(w http.ResponseWriter, r *http.Request) {
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

	// Only the session owner can delete the session
	if session.OwnerID != user.ID {
		http.Error(w, "Only session owner can delete the session", http.StatusForbidden)
		return
	}

	// Broadcast session end to all participants before deletion
	h.wsService.Broadcast(sessionID, models.SSEMessage{
		Type: "session-ended",
		Data: map[string]interface{}{
			"message": "Session has been ended by the owner",
		},
	})

	err = h.sessionService.DeleteSession(sessionID)
	if err != nil {
		http.Error(w, "Failed to delete session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) calculateVoteHistogram(votes []models.Vote) []VoteCount {
	voteCounts := make(map[string]int)
	total := len(votes)

	for _, vote := range votes {
		voteCounts[vote.VoteValue]++
	}

	var histogram []VoteCount
	// Only include vote values that actually received votes
	for voteValue, count := range voteCounts {
		if count > 0 {
			percentage := 0
			if total > 0 {
				percentage = (count * 100) / total
			}
			
			histogram = append(histogram, VoteCount{
				Value:      voteValue,
				Count:      count,
				Percentage: percentage,
			})
		}
	}

	return histogram
}

func (h *Handler) calculateVoteAverage(votes []models.Vote) *float64 {
	if len(votes) == 0 {
		return nil
	}
	
	var sum float64
	var count int
	
	for _, vote := range votes {
		// Only include numeric votes in average calculation
		// Skip special cards like ☕ and ?
		switch vote.VoteValue {
		case "0", "1", "2", "3", "5", "8", "13", "21", "34", "55", "89", "144":
			if val := parseVoteValue(vote.VoteValue); val >= 0 {
				sum += float64(val)
				count++
			}
		}
	}
	
	if count == 0 {
		return nil
	}
	
	average := sum / float64(count)
	return &average
}

func parseVoteValue(voteValue string) int {
	switch voteValue {
	case "0":
		return 0
	case "1":
		return 1
	case "2":
		return 2
	case "3":
		return 3
	case "5":
		return 5
	case "8":
		return 8
	case "13":
		return 13
	case "21":
		return 21
	case "34":
		return 34
	case "55":
		return 55
	case "89":
		return 89
	case "144":
		return 144
	default:
		return -1 // Invalid/special vote
	}
}

func (h *Handler) ReviewSession(w http.ResponseWriter, r *http.Request) {
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

	// Only the session owner can start a review
	if session.OwnerID != user.ID {
		http.Error(w, "Only session owner can start review", http.StatusForbidden)
		return
	}

	// End the session by broadcasting session-ended and marking it for review
	h.wsService.Broadcast(sessionID, models.SSEMessage{
		Type: "session-ended",
		Data: map[string]interface{}{
			"message": "Session review started by the owner",
			"redirect": "/session/" + sessionID + "/summary",
		},
	})

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetSessionSummary(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		// Redirect to home page
		http.Redirect(w, r, "/", http.StatusSeeOther)
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

	// Check if user was a participant
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

	// Calculate summary statistics
	totalVotes := 0
	estimatedTickets := 0
	var allVotes []models.Vote
	ticketAverages := make(map[int]float64)
	ticketVoteGroups := make(map[int][]VoteCount)

	for _, ticket := range session.Tickets {
		if len(ticket.Votes) > 0 {
			totalVotes += len(ticket.Votes)
			allVotes = append(allVotes, ticket.Votes...)
			
			if avg := h.calculateVoteAverage(ticket.Votes); avg != nil {
				ticketAverages[ticket.ID] = *avg
				estimatedTickets++
			}
			
			ticketVoteGroups[ticket.ID] = h.calculateVoteHistogram(ticket.Votes)
		}
	}

	// Calculate overall average
	var overallAverage float64
	if overallAvg := h.calculateVoteAverage(allVotes); overallAvg != nil {
		overallAverage = *overallAvg
	}

	// Calculate participant statistics
	participantStats := make(map[string]*ParticipantStat)
	for _, participant := range session.Participants {
		var participantVotes []models.Vote
		for _, ticket := range session.Tickets {
			for _, vote := range ticket.Votes {
				if vote.UserID == participant.ID {
					participantVotes = append(participantVotes, vote)
				}
			}
		}
		
		stat := &ParticipantStat{
			VoteCount: len(participantVotes),
		}
		
		if avg := h.calculateVoteAverage(participantVotes); avg != nil {
			stat.AverageVote = *avg
		}
		
		participantStats[participant.ID] = stat
	}

	data := PageData{
		Title:            session.Name + " - Summary",
		Template:         "summary",
		User:             user,
		Session:          session,
		SessionName:      session.Name,
		TicketAverages:   ticketAverages,
		TotalVotes:       totalVotes,
		EstimatedTickets: estimatedTickets,
		OverallAverage:   overallAverage,
		TicketVoteGroups: ticketVoteGroups,
		ParticipantStats: participantStats,
	}

	h.executeTemplate(w, "base.html", data)
}

func (h *Handler) executeTemplate(w http.ResponseWriter, tmplName string, data interface{}) {
	err := h.templates.ExecuteTemplate(w, tmplName, data)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
	}
}