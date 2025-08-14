package services

import (
	"database/sql"
	"fmt"
	"time"

	"poker-planning/internal/models"

	"github.com/google/uuid"
)

type SessionService struct {
	db *sql.DB
}

func NewSessionService(db *sql.DB) *SessionService {
	return &SessionService{db: db}
}

func (s *SessionService) CreateSession(name, ownerID string) (*models.Session, error) {
	sessionID := uuid.New().String()
	now := time.Now()

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `INSERT INTO sessions (id, name, owner_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`
	_, err = tx.Exec(query, sessionID, name, ownerID, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	participantQuery := `INSERT INTO participants (session_id, user_id, joined_at) VALUES (?, ?, ?)`
	_, err = tx.Exec(participantQuery, sessionID, ownerID, now)
	if err != nil {
		return nil, fmt.Errorf("failed to add owner as participant: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &models.Session{
		ID:        sessionID,
		Name:      name,
		OwnerID:   ownerID,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (s *SessionService) GetSessionByID(sessionID string) (*models.Session, error) {
	var session models.Session
	query := `SELECT id, name, owner_id, current_ticket_id, is_voting_active, created_at, updated_at 
			  FROM sessions WHERE id = ?`
	
	err := s.db.QueryRow(query, sessionID).Scan(
		&session.ID,
		&session.Name,
		&session.OwnerID,
		&session.CurrentTicketID,
		&session.IsVotingActive,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	participants, err := s.getSessionParticipants(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participants: %w", err)
	}
	session.Participants = participants

	tickets, err := s.getSessionTickets(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tickets: %w", err)
	}
	session.Tickets = tickets

	if session.CurrentTicketID != nil {
		for i, ticket := range tickets {
			if ticket.ID == *session.CurrentTicketID {
				session.CurrentTicket = &tickets[i]
				
				// Get votes for the current ticket
				votes, err := s.getTicketVotes(*session.CurrentTicketID)
				if err != nil {
					return nil, fmt.Errorf("failed to get ticket votes: %w", err)
				}
				session.CurrentTicket.Votes = votes
				break
			}
		}
	}

	return &session, nil
}

func (s *SessionService) JoinSession(sessionID, userID string) (bool, error) {
	// Check if user is already a participant
	checkQuery := `SELECT COUNT(*) FROM participants WHERE session_id = ? AND user_id = ?`
	var count int
	err := s.db.QueryRow(checkQuery, sessionID, userID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check participant status: %w", err)
	}
	
	if count > 0 {
		// User is already a participant
		return false, nil
	}
	
	// Add user as participant
	insertQuery := `INSERT INTO participants (session_id, user_id, joined_at) VALUES (?, ?, ?)`
	_, err = s.db.Exec(insertQuery, sessionID, userID, time.Now())
	if err != nil {
		return false, fmt.Errorf("failed to join session: %w", err)
	}
	
	// User was actually added
	return true, nil
}

func (s *SessionService) LeaveSession(sessionID, userID string) error {
	query := `DELETE FROM participants WHERE session_id = ? AND user_id = ?`
	_, err := s.db.Exec(query, sessionID, userID)
	if err != nil {
		return fmt.Errorf("failed to leave session: %w", err)
	}
	return nil
}

func (s *SessionService) getSessionParticipants(sessionID string) ([]models.User, error) {
	query := `SELECT u.id, u.username, u.created_at, u.last_seen 
			  FROM users u 
			  JOIN participants p ON u.id = p.user_id 
			  WHERE p.session_id = ? 
			  ORDER BY p.joined_at`
	
	rows, err := s.db.Query(query, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participants []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Username, &user.CreatedAt, &user.LastSeen)
		if err != nil {
			return nil, err
		}
		participants = append(participants, user)
	}

	return participants, nil
}

func (s *SessionService) getSessionTickets(sessionID string) ([]models.Ticket, error) {
	query := `SELECT id, session_id, title, description, final_estimate, position, created_at 
			  FROM tickets 
			  WHERE session_id = ? 
			  ORDER BY position`
	
	rows, err := s.db.Query(query, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []models.Ticket
	for rows.Next() {
		var ticket models.Ticket
		err := rows.Scan(
			&ticket.ID,
			&ticket.SessionID,
			&ticket.Title,
			&ticket.Description,
			&ticket.FinalEstimate,
			&ticket.Position,
			&ticket.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		// Load votes for each ticket
		votes, err := s.getTicketVotes(ticket.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get votes for ticket %d: %w", ticket.ID, err)
		}
		ticket.Votes = votes
		
		tickets = append(tickets, ticket)
	}

	return tickets, nil
}

func (s *SessionService) getTicketVotes(ticketID int) ([]models.Vote, error) {
	query := `SELECT v.id, v.ticket_id, v.user_id, v.vote_value, v.created_at,
					 u.username
			  FROM votes v
			  JOIN users u ON v.user_id = u.id
			  WHERE v.ticket_id = ?
			  ORDER BY v.created_at`
	
	rows, err := s.db.Query(query, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var votes []models.Vote
	for rows.Next() {
		var vote models.Vote
		var user models.User
		
		err := rows.Scan(
			&vote.ID,
			&vote.TicketID,
			&vote.UserID,
			&vote.VoteValue,
			&vote.CreatedAt,
			&user.Username,
		)
		if err != nil {
			return nil, err
		}
		
		user.ID = vote.UserID
		vote.User = &user
		votes = append(votes, vote)
	}

	return votes, nil
}

func (s *SessionService) UpdateSession(session *models.Session) error {
	query := `UPDATE sessions SET 
			  name = ?, 
			  current_ticket_id = ?, 
			  is_voting_active = ?, 
			  updated_at = ? 
			  WHERE id = ?`
	
	_, err := s.db.Exec(query,
		session.Name,
		session.CurrentTicketID,
		session.IsVotingActive,
		time.Now(),
		session.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}
	
	return nil
}

func (s *SessionService) DeleteSession(sessionID string) error {
	// Note: SQLite with ON DELETE CASCADE will automatically handle deletion of:
	// - participants
	// - tickets (and their votes due to ticket FK constraint)
	query := `DELETE FROM sessions WHERE id = ?`
	_, err := s.db.Exec(query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}