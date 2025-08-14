package services

import (
	"database/sql"
	"fmt"
	"time"

	"poker-planning/internal/models"

	"github.com/google/uuid"
)

type UserService struct {
	db *sql.DB
}

func NewUserService(db *sql.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) CreateUser(username string) (*models.User, error) {
	userID := uuid.New().String()
	now := time.Now()

	query := `INSERT INTO users (id, username, created_at, last_seen) VALUES (?, ?, ?, ?)`
	_, err := s.db.Exec(query, userID, username, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &models.User{
		ID:        userID,
		Username:  username,
		CreatedAt: now,
		LastSeen:  now,
	}, nil
}

func (s *UserService) GetUserByID(userID string) (*models.User, error) {
	var user models.User
	query := `SELECT id, username, created_at, last_seen FROM users WHERE id = ?`
	
	err := s.db.QueryRow(query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.CreatedAt,
		&user.LastSeen,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (s *UserService) UpdateLastSeen(userID string) error {
	query := `UPDATE users SET last_seen = ? WHERE id = ?`
	_, err := s.db.Exec(query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update last seen: %w", err)
	}
	return nil
}

func (s *UserService) CleanupInactiveUsers() error {
	cutoff := time.Now().Add(-6 * time.Hour)
	query := `DELETE FROM users WHERE last_seen < ?`
	
	_, err := s.db.Exec(query, cutoff)
	if err != nil {
		return fmt.Errorf("failed to cleanup inactive users: %w", err)
	}
	
	return nil
}