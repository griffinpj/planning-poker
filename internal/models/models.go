package models

import (
	"time"
)

type User struct {
	ID       string    `json:"id"`
	Username string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	LastSeen  time.Time `json:"last_seen"`
}

type Session struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	OwnerID         string     `json:"owner_id"`
	CurrentTicketID *int       `json:"current_ticket_id"`
	IsVotingActive  bool       `json:"is_voting_active"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	Participants    []User     `json:"participants,omitempty"`
	Tickets         []Ticket   `json:"tickets,omitempty"`
	CurrentTicket   *Ticket    `json:"current_ticket,omitempty"`
}

type Ticket struct {
	ID            int     `json:"id"`
	SessionID     string  `json:"session_id"`
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	FinalEstimate *int    `json:"final_estimate"`
	Position      int     `json:"position"`
	CreatedAt     time.Time `json:"created_at"`
	Votes         []Vote  `json:"votes,omitempty"`
}

type Vote struct {
	ID        int       `json:"id"`
	TicketID  int       `json:"ticket_id"`
	UserID    string    `json:"user_id"`
	VoteValue string    `json:"vote_value"`
	CreatedAt time.Time `json:"created_at"`
	User      *User     `json:"user,omitempty"`
}

type Participant struct {
	SessionID string    `json:"session_id"`
	UserID    string    `json:"user_id"`
	JoinedAt  time.Time `json:"joined_at"`
	User      *User     `json:"user,omitempty"`
}

type RecentEmoji struct {
	UserID string    `json:"user_id"`
	Emoji  string    `json:"emoji"`
	UsedAt time.Time `json:"used_at"`
}

type SSEMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type EmojiReaction struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Emoji  string `json:"emoji"`
	FromUser *User `json:"from_user,omitempty"`
	ToUser   *User `json:"to_user,omitempty"`
}

const (
	VotingCards = "0,1,2,3,5,8,13,21,34,☕,?"
)

var FibonacciCards = []string{"0", "1", "2", "3", "5", "8", "13", "21", "34", "55", "89", "144"}
var SpecialCards = []string{"☕", "?"}

func AllVotingCards() []string {
	cards := make([]string, len(FibonacciCards)+len(SpecialCards))
	copy(cards, FibonacciCards)
	copy(cards[len(FibonacciCards):], SpecialCards)
	return cards
}