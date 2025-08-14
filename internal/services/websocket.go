package services

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"poker-planning/internal/models"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin
	},
}

type WSClient struct {
	ID        string
	SessionID string
	UserID    string
	Conn      *websocket.Conn
	Send      chan models.SSEMessage
}

type WSService struct {
	clients    map[string]*WSClient
	register   chan *WSClient
	unregister chan *WSClient
	broadcast  chan BroadcastMessage
	mutex      sync.RWMutex
}

type BroadcastMessage struct {
	SessionID string
	Message   models.SSEMessage
}

func NewWSService() *WSService {
	return &WSService{
		clients:    make(map[string]*WSClient),
		register:   make(chan *WSClient),
		unregister: make(chan *WSClient),
		broadcast:  make(chan BroadcastMessage),
	}
}

func (ws *WSService) Run() {
	for {
		select {
		case client := <-ws.register:
			ws.mutex.Lock()
			ws.clients[client.ID] = client
			ws.mutex.Unlock()
			log.Printf("WebSocket client connected: %s", client.ID)

		case client := <-ws.unregister:
			ws.mutex.Lock()
			if _, ok := ws.clients[client.ID]; ok {
				delete(ws.clients, client.ID)
				close(client.Send)
			}
			ws.mutex.Unlock()
			log.Printf("WebSocket client disconnected: %s", client.ID)

		case message := <-ws.broadcast:
			ws.mutex.RLock()
			clientCount := 0
			for _, client := range ws.clients {
				if client.SessionID == message.SessionID {
					clientCount++
					select {
					case client.Send <- message.Message:
					default:
						delete(ws.clients, client.ID)
						close(client.Send)
					}
				}
			}
			ws.mutex.RUnlock()
			log.Printf("WebSocket broadcast: type=%s, sessionID=%s, clients=%d", message.Message.Type, message.SessionID, clientCount)
		}
	}
}

func (ws *WSService) HandleWebSocket(w http.ResponseWriter, r *http.Request, sessionID, userID string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	clientID := sessionID + "_" + userID
	client := &WSClient{
		ID:        clientID,
		SessionID: sessionID,
		UserID:    userID,
		Conn:      conn,
		Send:      make(chan models.SSEMessage, 256),
	}

	ws.register <- client

	go ws.writePump(client)
	go ws.readPump(client)
}

func (ws *WSService) readPump(client *WSClient) {
	defer func() {
		ws.unregister <- client
		client.Conn.Close()
	}()

	client.Conn.SetReadLimit(512)
	for {
		_, _, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
	}
}

func (ws *WSService) writePump(client *WSClient) {
	defer client.Conn.Close()

	// Send connection confirmation
	connectMsg := models.SSEMessage{
		Type: "connected",
		Data: map[string]interface{}{
			"client_id": client.ID,
		},
	}

	data, _ := json.Marshal(connectMsg)
	if err := client.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return
	}

	for {
		select {
		case message, ok := <-client.Send:
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			data, err := json.Marshal(message)
			if err != nil {
				log.Printf("JSON marshal error: %v", err)
				continue
			}

			if err := client.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
		}
	}
}

func (ws *WSService) Broadcast(sessionID string, message models.SSEMessage) {
	ws.broadcast <- BroadcastMessage{
		SessionID: sessionID,
		Message:   message,
	}
}

func (ws *WSService) SendToUser(sessionID, userID string, message models.SSEMessage) {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()

	for _, client := range ws.clients {
		if client.SessionID == sessionID && client.UserID == userID {
			select {
			case client.Send <- message:
			default:
				delete(ws.clients, client.ID)
				close(client.Send)
			}
		}
	}
}

func (ws *WSService) GetClientCount(sessionID string) int {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()

	count := 0
	for _, client := range ws.clients {
		if client.SessionID == sessionID {
			count++
		}
	}
	return count
}