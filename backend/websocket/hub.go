package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

// Hub maintains active connections and broadcasts messages
type Hub struct {
	// Registered clients by user ID
	clients map[int]map[*Client]bool

	// Register requests
	register chan *Client

	// Unregister requests
	unregister chan *Client

	// Inbound messages to broadcast
	broadcast chan *Event

	mu sync.RWMutex
}

// Event types for WebSocket communication
type EventType string

const (
	EventTypeMessage       EventType = "message"
	EventTypeTyping        EventType = "typing"
	EventTypeStopTyping    EventType = "stop_typing"
	EventTypeOnline        EventType = "online"
	EventTypeOffline       EventType = "offline"
	EventTypeRead          EventType = "read"
	EventTypeConversation  EventType = "conversation"
	EventTypeUserJoined    EventType = "user_joined"
	EventTypeUserLeft      EventType = "user_left"
	EventTypeMessageDelete EventType = "message_delete"
	EventTypeMessageEdit   EventType = "message_edit"
)

// Event is the WebSocket message format
type Event struct {
	Type           EventType       `json:"type"`
	ConversationID int             `json:"conversation_id,omitempty"`
	UserID         int             `json:"user_id,omitempty"`
	Payload        json.RawMessage `json:"payload,omitempty"`
	Timestamp      time.Time       `json:"timestamp"`

	// Internal: target users (not sent to client)
	TargetUsers []int `json:"-"`
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[int]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Event, 256),
	}
}

// Register adds a client to the hub
func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.UserID] == nil {
				h.clients[client.UserID] = make(map[*Client]bool)
			}
			h.clients[client.UserID][client] = true
			h.mu.Unlock()

			log.Printf("Client connected: user %d", client.UserID)

			// Broadcast online status
			h.BroadcastToAll(&Event{
				Type:      EventTypeOnline,
				UserID:    client.UserID,
				Timestamp: time.Now(),
			})

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.UserID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.send)
					if len(clients) == 0 {
						delete(h.clients, client.UserID)
						// Broadcast offline status
						h.mu.Unlock()
						h.BroadcastToAll(&Event{
							Type:      EventTypeOffline,
							UserID:    client.UserID,
							Timestamp: time.Now(),
						})
						h.mu.Lock()
					}
				}
			}
			h.mu.Unlock()
			log.Printf("Client disconnected: user %d", client.UserID)

		case event := <-h.broadcast:
			h.mu.RLock()
			if len(event.TargetUsers) > 0 {
				// Send to specific users
				for _, userID := range event.TargetUsers {
					if clients, ok := h.clients[userID]; ok {
						for client := range clients {
							select {
							case client.send <- event:
							default:
								close(client.send)
								delete(clients, client)
							}
						}
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastToUsers sends event to specific users
func (h *Hub) BroadcastToUsers(userIDs []int, event *Event) {
	event.TargetUsers = userIDs
	event.Timestamp = time.Now()
	h.broadcast <- event
}

// BroadcastToAll sends event to all connected users
func (h *Hub) BroadcastToAll(event *Event) {
	h.mu.RLock()
	allUsers := make([]int, 0, len(h.clients))
	for userID := range h.clients {
		allUsers = append(allUsers, userID)
	}
	h.mu.RUnlock()

	h.BroadcastToUsers(allUsers, event)
}

// IsUserOnline checks if user has active connections
func (h *Hub) IsUserOnline(userID int) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	clients, ok := h.clients[userID]
	return ok && len(clients) > 0
}

// GetOnlineUsers returns list of online user IDs
func (h *Hub) GetOnlineUsers() []int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	users := make([]int, 0, len(h.clients))
	for userID := range h.clients {
		users = append(users, userID)
	}
	return users
}
