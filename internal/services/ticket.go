package services

import (
	"database/sql"
	"fmt"
	"time"

	"poker-planning/internal/models"
)

type TicketService struct {
	db *sql.DB
}

func NewTicketService(db *sql.DB) *TicketService {
	return &TicketService{db: db}
}

func (s *TicketService) CreateTicket(sessionID, title, description string) (*models.Ticket, error) {
	now := time.Now()
	
	// Get next position
	var maxPosition int
	posQuery := `SELECT COALESCE(MAX(position), 0) FROM tickets WHERE session_id = ?`
	err := s.db.QueryRow(posQuery, sessionID).Scan(&maxPosition)
	if err != nil {
		return nil, fmt.Errorf("failed to get max position: %w", err)
	}

	query := `INSERT INTO tickets (session_id, title, description, position, created_at) 
			  VALUES (?, ?, ?, ?, ?)`
	
	result, err := s.db.Exec(query, sessionID, title, description, maxPosition+1, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create ticket: %w", err)
	}

	ticketID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket ID: %w", err)
	}

	return &models.Ticket{
		ID:          int(ticketID),
		SessionID:   sessionID,
		Title:       title,
		Description: description,
		Position:    maxPosition + 1,
		CreatedAt:   now,
	}, nil
}

func (s *TicketService) GetTicketByID(ticketID int) (*models.Ticket, error) {
	var ticket models.Ticket
	query := `SELECT id, session_id, title, description, final_estimate, position, created_at 
			  FROM tickets WHERE id = ?`
	
	err := s.db.QueryRow(query, ticketID).Scan(
		&ticket.ID,
		&ticket.SessionID,
		&ticket.Title,
		&ticket.Description,
		&ticket.FinalEstimate,
		&ticket.Position,
		&ticket.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get ticket: %w", err)
	}

	return &ticket, nil
}

func (s *TicketService) UpdateTicket(ticket *models.Ticket) error {
	query := `UPDATE tickets SET 
			  title = ?, 
			  description = ?, 
			  final_estimate = ?, 
			  position = ? 
			  WHERE id = ?`
	
	_, err := s.db.Exec(query,
		ticket.Title,
		ticket.Description,
		ticket.FinalEstimate,
		ticket.Position,
		ticket.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update ticket: %w", err)
	}
	
	return nil
}

func (s *TicketService) DeleteTicket(ticketID int) error {
	// Get the ticket to find its position and session
	ticket, err := s.GetTicketByID(ticketID)
	if err != nil {
		return fmt.Errorf("failed to get ticket: %w", err)
	}
	if ticket == nil {
		return fmt.Errorf("ticket not found")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete the ticket
	deleteQuery := `DELETE FROM tickets WHERE id = ?`
	_, err = tx.Exec(deleteQuery, ticketID)
	if err != nil {
		return fmt.Errorf("failed to delete ticket: %w", err)
	}

	// Update positions of subsequent tickets
	updateQuery := `UPDATE tickets SET position = position - 1 
					WHERE session_id = ? AND position > ?`
	_, err = tx.Exec(updateQuery, ticket.SessionID, ticket.Position)
	if err != nil {
		return fmt.Errorf("failed to update positions: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *TicketService) GetTicketsForSession(sessionID string) ([]models.Ticket, error) {
	query := `SELECT id, session_id, title, description, final_estimate, position, created_at 
			  FROM tickets 
			  WHERE session_id = ? 
			  ORDER BY position`
	
	rows, err := s.db.Query(query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tickets: %w", err)
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
			return nil, fmt.Errorf("failed to scan ticket: %w", err)
		}
		tickets = append(tickets, ticket)
	}

	return tickets, nil
}

func (s *TicketService) SetFinalEstimate(ticketID int, estimate int) error {
	query := `UPDATE tickets SET final_estimate = ? WHERE id = ?`
	_, err := s.db.Exec(query, estimate, ticketID)
	if err != nil {
		return fmt.Errorf("failed to set final estimate: %w", err)
	}
	return nil
}

func (s *TicketService) ReorderTickets(sessionID string, ticketIDs []int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `UPDATE tickets SET position = ? WHERE id = ? AND session_id = ?`
	
	for i, ticketID := range ticketIDs {
		_, err = tx.Exec(query, i+1, ticketID, sessionID)
		if err != nil {
			return fmt.Errorf("failed to update ticket position: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}