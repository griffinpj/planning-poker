package utils

import (
	"fmt"
	"html"
	"regexp"
	"strings"
)

var (
	// Username validation: 1-50 characters, alphanumeric plus spaces, hyphens, underscores
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9\s\-_]{1,50}$`)
	
	// Session name validation: 1-100 characters
	sessionNameRegex = regexp.MustCompile(`^.{1,100}$`)
	
	// Ticket title validation: 1-200 characters
	ticketTitleRegex = regexp.MustCompile(`^.{1,200}$`)
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	
	messages := make([]string, len(e))
	for i, err := range e {
		messages[i] = err.Error()
	}
	
	return strings.Join(messages, "; ")
}

func (e ValidationErrors) HasErrors() bool {
	return len(e) > 0
}

func ValidateUsername(username string) ValidationErrors {
	var errors ValidationErrors
	
	username = strings.TrimSpace(username)
	
	if username == "" {
		errors = append(errors, ValidationError{
			Field:   "username",
			Message: "Username is required",
		})
		return errors
	}
	
	if !usernameRegex.MatchString(username) {
		errors = append(errors, ValidationError{
			Field:   "username",
			Message: "Username must be 1-50 characters and contain only letters, numbers, spaces, hyphens, and underscores",
		})
	}
	
	return errors
}

func ValidateSessionName(name string) ValidationErrors {
	var errors ValidationErrors
	
	name = strings.TrimSpace(name)
	
	if name == "" {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "Session name is required",
		})
		return errors
	}
	
	if !sessionNameRegex.MatchString(name) {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "Session name must be 1-100 characters",
		})
	}
	
	return errors
}

func ValidateTicketTitle(title string) ValidationErrors {
	var errors ValidationErrors
	
	title = strings.TrimSpace(title)
	
	if title == "" {
		errors = append(errors, ValidationError{
			Field:   "title",
			Message: "Ticket title is required",
		})
		return errors
	}
	
	if !ticketTitleRegex.MatchString(title) {
		errors = append(errors, ValidationError{
			Field:   "title",
			Message: "Ticket title must be 1-200 characters",
		})
	}
	
	return errors
}

func ValidateTicketDescription(description string) ValidationErrors {
	var errors ValidationErrors
	
	// Description is optional, but if provided, limit to 1000 characters
	if len(description) > 1000 {
		errors = append(errors, ValidationError{
			Field:   "description",
			Message: "Ticket description must be no more than 1000 characters",
		})
	}
	
	return errors
}

func SanitizeInput(input string) string {
	// Trim whitespace and escape HTML
	return html.EscapeString(strings.TrimSpace(input))
}

func ValidateVoteValue(voteValue string) ValidationErrors {
	var errors ValidationErrors
	
	validVotes := []string{"0", "1", "2", "3", "5", "8", "13", "21", "34", "☕", "?"}
	
	for _, valid := range validVotes {
		if voteValue == valid {
			return errors // No errors if valid
		}
	}
	
	errors = append(errors, ValidationError{
		Field:   "vote",
		Message: "Invalid vote value",
	})
	
	return errors
}

func ValidateEmoji(emoji string) ValidationErrors {
	var errors ValidationErrors
	
	if emoji == "" {
		errors = append(errors, ValidationError{
			Field:   "emoji",
			Message: "Emoji is required",
		})
		return errors
	}
	
	// Simple check for emoji length (most emojis are 1-4 bytes in UTF-8)
	if len(emoji) > 10 {
		errors = append(errors, ValidationError{
			Field:   "emoji",
			Message: "Invalid emoji",
		})
	}
	
	return errors
}