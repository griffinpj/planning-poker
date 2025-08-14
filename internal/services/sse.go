package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"poker-planning/internal/models"
)

type SSEClient struct {
	ID        string
	SessionID string
	UserID    string
	Channel   chan models.SSEMessage
	Request   *http.Request
}

type SSEService struct {
	clients map[string]*SSEClient
	mutex   sync.RWMutex
}

func NewSSEService() *SSEService {
	return &SSEService{
		clients: make(map[string]*SSEClient),
	}
}

func (s *SSEService) AddClient(sessionID, userID string, r *http.Request) *SSEClient {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	clientID := fmt.Sprintf("%s_%s_%d", sessionID, userID, time.Now().UnixNano())
	
	client := &SSEClient{
		ID:        clientID,
		SessionID: sessionID,
		UserID:    userID,
		Channel:   make(chan models.SSEMessage, 10),
		Request:   r,
	}

	s.clients[clientID] = client
	return client
}

func (s *SSEService) RemoveClient(clientID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if client, exists := s.clients[clientID]; exists {
		close(client.Channel)
		delete(s.clients, clientID)
	}
}

func (s *SSEService) Broadcast(sessionID string, message models.SSEMessage) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	clientCount := 0
	for _, client := range s.clients {
		if client.SessionID == sessionID {
			clientCount++
			select {
			case client.Channel <- message:
			case <-time.After(100 * time.Millisecond):
				// Client channel is full or client is slow, skip
			}
		}
	}
	
	// Debug logging
	fmt.Printf("SSE Broadcast: type=%s, sessionID=%s, clients=%d\n", message.Type, sessionID, clientCount)
}

func (s *SSEService) SendToUser(sessionID, userID string, message models.SSEMessage) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, client := range s.clients {
		if client.SessionID == sessionID && client.UserID == userID {
			select {
			case client.Channel <- message:
			case <-time.After(100 * time.Millisecond):
				// Client channel is full or client is slow, skip
			}
		}
	}
}

func (s *SSEService) GetClientCount(sessionID string) int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	count := 0
	for _, client := range s.clients {
		if client.SessionID == sessionID {
			count++
		}
	}
	return count
}

func (s *SSEService) HandleSSE(w http.ResponseWriter, client *SSEClient) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Send initial connection message
	s.sendSSEMessage(w, flusher, "connected", map[string]interface{}{
		"client_id": client.ID,
		"timestamp": time.Now().Unix(),
	})

	// Clean up when connection closes
	defer func() {
		s.RemoveClient(client.ID)
	}()

	// Send heartbeat every 30 seconds
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	ctx := client.Request.Context()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.sendSSEMessage(w, flusher, "heartbeat", map[string]interface{}{
				"timestamp": time.Now().Unix(),
			})
		case message := <-client.Channel:
			data, err := json.Marshal(message.Data)
			if err != nil {
				continue
			}
			s.sendSSEMessage(w, flusher, message.Type, string(data))
		}
	}
}

func (s *SSEService) sendSSEMessage(w http.ResponseWriter, flusher http.Flusher, eventType, data interface{}) {
	dataStr, ok := data.(string)
	if !ok {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return
		}
		dataStr = string(jsonData)
	}

	fmt.Fprintf(w, "event: %s\n", eventType)
	fmt.Fprintf(w, "data: %s\n\n", dataStr)
	flusher.Flush()
}