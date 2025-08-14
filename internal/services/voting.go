package services

import (
	"database/sql"
	"fmt"
	"time"

	"poker-planning/internal/models"
)

type VotingService struct {
	db *sql.DB
}

func NewVotingService(db *sql.DB) *VotingService {
	return &VotingService{db: db}
}

func (s *VotingService) SubmitVote(ticketID int, userID, voteValue string) (*models.Vote, error) {
	now := time.Now()
	
	query := `INSERT OR REPLACE INTO votes (ticket_id, user_id, vote_value, created_at) 
			  VALUES (?, ?, ?, ?)`
	
	result, err := s.db.Exec(query, ticketID, userID, voteValue, now)
	if err != nil {
		return nil, fmt.Errorf("failed to submit vote: %w", err)
	}

	voteID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get vote ID: %w", err)
	}

	return &models.Vote{
		ID:        int(voteID),
		TicketID:  ticketID,
		UserID:    userID,
		VoteValue: voteValue,
		CreatedAt: now,
	}, nil
}

func (s *VotingService) GetVotesForTicket(ticketID int) ([]models.Vote, error) {
	query := `SELECT v.id, v.ticket_id, v.user_id, v.vote_value, v.created_at,
					 u.username
			  FROM votes v
			  JOIN users u ON v.user_id = u.id
			  WHERE v.ticket_id = ?
			  ORDER BY v.created_at`
	
	rows, err := s.db.Query(query, ticketID)
	if err != nil {
		return nil, fmt.Errorf("failed to get votes: %w", err)
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
			return nil, fmt.Errorf("failed to scan vote: %w", err)
		}
		
		user.ID = vote.UserID
		vote.User = &user
		votes = append(votes, vote)
	}

	return votes, nil
}

func (s *VotingService) ClearVotesForTicket(ticketID int) error {
	query := `DELETE FROM votes WHERE ticket_id = ?`
	
	_, err := s.db.Exec(query, ticketID)
	if err != nil {
		return fmt.Errorf("failed to clear votes: %w", err)
	}
	
	return nil
}

func (s *VotingService) GetUserVoteForTicket(ticketID int, userID string) (*models.Vote, error) {
	var vote models.Vote
	query := `SELECT id, ticket_id, user_id, vote_value, created_at 
			  FROM votes 
			  WHERE ticket_id = ? AND user_id = ?`
	
	err := s.db.QueryRow(query, ticketID, userID).Scan(
		&vote.ID,
		&vote.TicketID,
		&vote.UserID,
		&vote.VoteValue,
		&vote.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user vote: %w", err)
	}

	return &vote, nil
}