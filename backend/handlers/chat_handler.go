package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"go-to-do/auth"
	"go-to-do/models"
	"go-to-do/repository"
	"go-to-do/websocket"

	ws "github.com/gorilla/websocket"
)

type ChatHandler struct {
	repo        *repository.ChatRepository
	userRepo    *repository.UserRepository
	authService *auth.AuthService
	hub         *websocket.Hub
}

func NewChatHandler(repo *repository.ChatRepository, userRepo *repository.UserRepository, authService *auth.AuthService, hub *websocket.Hub) *ChatHandler {
	return &ChatHandler{
		repo:        repo,
		userRepo:    userRepo,
		authService: authService,
		hub:         hub,
	}
}

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

func (h *ChatHandler) requireAuth(w http.ResponseWriter, r *http.Request) int {
	user, err := h.authService.GetUserFromRequest(r)
	if err != nil || user == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Unauthorized"})
		return 0
	}
	return user.ID
}

// RequireAuthPublic is exported for use in main.go
func (h *ChatHandler) RequireAuthPublic(w http.ResponseWriter, r *http.Request) int {
	return h.requireAuth(w, r)
}

// WebSocket endpoint
func (h *ChatHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	user, err := h.authService.GetUserFromRequest(r)
	if err != nil || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := websocket.NewClient(h.hub, conn, user.ID)
	h.hub.Register(client)

	// Update last seen
	h.repo.UpdateLastSeen(user.ID)

	go client.WritePump()
	go client.ReadPump()
}

// GET /api/conversations
func (h *ChatHandler) ListConversations(w http.ResponseWriter, r *http.Request) {
	userID := h.requireAuth(w, r)
	if userID == 0 {
		return
	}

	conversations, err := h.repo.GetUserConversations(userID)
	if err != nil {
		log.Printf("Error getting conversations: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Failed to get conversations"})
		return
	}

	if conversations == nil {
		conversations = []models.Conversation{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "data": conversations})
}

type CreateDirectChatRequest struct {
	UserID int    `json:"user_id"`
	Phone  string `json:"phone"`
	Email  string `json:"email"`
}

// POST /api/conversations/direct
func (h *ChatHandler) CreateDirectChat(w http.ResponseWriter, r *http.Request) {
	userID := h.requireAuth(w, r)
	if userID == 0 {
		return
	}

	var req CreateDirectChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Invalid request"})
		return
	}

	var targetUserID int
	if req.UserID > 0 {
		targetUserID = req.UserID
	} else if req.Phone != "" {
		user, err := h.userRepo.GetByPhone(req.Phone)
		if err != nil || user == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "User not found"})
			return
		}
		targetUserID = user.ID
	} else if req.Email != "" {
		user, err := h.userRepo.GetByEmail(req.Email)
		if err != nil || user == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "User not found"})
			return
		}
		targetUserID = user.ID
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Provide user_id, phone, or email"})
		return
	}

	if targetUserID == userID {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Cannot chat with yourself"})
		return
	}

	conv, err := h.repo.CreateDirectConversation(userID, targetUserID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Failed to create conversation"})
		return
	}

	// Notify both users about new conversation
	payload, _ := json.Marshal(conv)
	h.hub.BroadcastToUsers([]int{userID, targetUserID}, &websocket.Event{
		Type:           websocket.EventTypeConversation,
		ConversationID: conv.ID,
		Payload:        payload,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "data": conv})
}

type CreateGroupRequest struct {
	Name      string `json:"name"`
	MemberIDs []int  `json:"member_ids"`
}

// POST /api/conversations/group
func (h *ChatHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	userID := h.requireAuth(w, r)
	if userID == 0 {
		return
	}

	var req CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Group name required", http.StatusBadRequest)
		return
	}

	conv, err := h.repo.CreateGroupConversation(req.Name, userID, req.MemberIDs)
	if err != nil {
		http.Error(w, "Failed to create group", http.StatusInternalServerError)
		return
	}

	// Notify members
	memberIDs, _ := h.repo.GetConversationMemberIDs(conv.ID)
	payload, _ := json.Marshal(conv)
	h.hub.BroadcastToUsers(memberIDs, &websocket.Event{
		Type:           websocket.EventTypeConversation,
		ConversationID: conv.ID,
		Payload:        payload,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(conv)
}

// GET /api/conversations/{id}/messages
func (h *ChatHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	userID := h.requireAuth(w, r)
	if userID == 0 {
		return
	}

	convIDStr := r.URL.Query().Get("id")
	convID, err := strconv.Atoi(convIDStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Invalid conversation ID"})
		return
	}

	// Verify user is member
	_, err = h.repo.GetConversationByID(convID, userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Conversation not found"})
		return
	}

	limit := 50
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		limit, _ = strconv.Atoi(l)
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		offset, _ = strconv.Atoi(o)
	}

	messages, err := h.repo.GetMessages(convID, limit, offset)
	if err != nil {
		log.Printf("Error getting messages: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Failed to get messages"})
		return
	}

	if messages == nil {
		messages = []models.Message{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "data": messages})
}

type SendMessageRequest struct {
	Content       string    `json:"content"`
	Type          string    `json:"type"`
	ReplyToID     *int      `json:"reply_to_id"`
	VoiceFilePath string    `json:"voice_file_path"`
	VoiceDuration float64   `json:"voice_duration"`
	VoiceWaveform []float64 `json:"voice_waveform"`
	VoiceFileSize int64     `json:"voice_file_size"`
}

// POST /api/conversations/{id}/messages
func (h *ChatHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	userID := h.requireAuth(w, r)
	if userID == 0 {
		return
	}

	convIDStr := r.URL.Query().Get("id")
	convID, err := strconv.Atoi(convIDStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Invalid conversation ID"})
		return
	}

	// Verify user is member
	_, err = h.repo.GetConversationByID(convID, userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Conversation not found"})
		return
	}

	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Invalid request"})
		return
	}

	if req.Content == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Message content required"})
		return
	}

	msgType := models.MessageTypeText
	if req.Type != "" {
		msgType = models.MessageType(req.Type)
	}

	msg, err := h.repo.CreateMessage(convID, userID, msgType, req.Content, req.ReplyToID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Failed to send message"})
		return
	}

	// Add voice data if this is a voice message
	if msgType == models.MessageTypeVoice && req.VoiceFilePath != "" {
		msg.VoiceFilePath = req.VoiceFilePath
		msg.VoiceDuration = req.VoiceDuration
		msg.VoiceWaveform = req.VoiceWaveform
		msg.VoiceFileSize = req.VoiceFileSize
	}

	// Get sender info
	sender, _ := h.userRepo.GetByID(userID)
	msg.Sender = sender

	// Broadcast to conversation members
	memberIDs, _ := h.repo.GetConversationMemberIDs(convID)
	payload, _ := json.Marshal(msg)
	h.hub.BroadcastToUsers(memberIDs, &websocket.Event{
		Type:           websocket.EventTypeMessage,
		ConversationID: convID,
		UserID:         userID,
		Payload:        payload,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "data": msg})
}

type EditMessageRequest struct {
	Content string `json:"content"`
}

// PUT /api/conversations/messages/edit
func (h *ChatHandler) EditMessage(w http.ResponseWriter, r *http.Request) {
	userID := h.requireAuth(w, r)
	if userID == 0 {
		return
	}

	convIDStr := r.URL.Query().Get("id")
	convID, err := strconv.Atoi(convIDStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Invalid conversation ID"})
		return
	}

	msgIDStr := r.URL.Query().Get("msg_id")
	msgID, err := strconv.Atoi(msgIDStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Invalid message ID"})
		return
	}

	// Verify user is member of conversation
	_, err = h.repo.GetConversationByID(convID, userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Conversation not found"})
		return
	}

	var req EditMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Invalid request"})
		return
	}

	if req.Content == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Message content required"})
		return
	}

	msg, err := h.repo.UpdateMessage(msgID, userID, req.Content)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Failed to update message"})
		return
	}

	// Get sender info
	sender, _ := h.userRepo.GetByID(userID)
	msg.Sender = sender

	// Broadcast edit to conversation members
	memberIDs, _ := h.repo.GetConversationMemberIDs(convID)
	if len(memberIDs) > 0 {
		payload, _ := json.Marshal(msg)
		h.hub.BroadcastToUsers(memberIDs, &websocket.Event{
			Type:           websocket.EventTypeMessageEdit,
			ConversationID: convID,
			Payload:        payload,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "data": msg})
}

// POST /api/conversations/{id}/read
func (h *ChatHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	userID := h.requireAuth(w, r)
	if userID == 0 {
		return
	}

	convIDStr := r.URL.Query().Get("id")
	convID, err := strconv.Atoi(convIDStr)
	if err != nil {
		http.Error(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}

	if err := h.repo.MarkAsRead(convID, userID); err != nil {
		http.Error(w, "Failed to mark as read", http.StatusInternalServerError)
		return
	}

	// Broadcast read receipt
	memberIDs, _ := h.repo.GetConversationMemberIDs(convID)
	h.hub.BroadcastToUsers(memberIDs, &websocket.Event{
		Type:           websocket.EventTypeRead,
		ConversationID: convID,
		UserID:         userID,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// POST /api/conversations/{id}/typing
func (h *ChatHandler) SendTyping(w http.ResponseWriter, r *http.Request) {
	userID := h.requireAuth(w, r)
	if userID == 0 {
		return
	}

	convIDStr := r.URL.Query().Get("id")
	convID, err := strconv.Atoi(convIDStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Invalid conversation ID"})
		return
	}

	memberIDs, _ := h.repo.GetConversationMemberIDs(convID)
	h.hub.BroadcastToUsers(memberIDs, &websocket.Event{
		Type:           websocket.EventTypeTyping,
		ConversationID: convID,
		UserID:         userID,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// GET /api/users/online
func (h *ChatHandler) GetOnlineUsers(w http.ResponseWriter, r *http.Request) {
	userID := h.requireAuth(w, r)
	if userID == 0 {
		return
	}

	onlineUsers := h.hub.GetOnlineUsers()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"online_users": onlineUsers,
	})
}

// DELETE /api/conversations/delete
func (h *ChatHandler) DeleteConversation(w http.ResponseWriter, r *http.Request) {
	userID := h.requireAuth(w, r)
	if userID == 0 {
		return
	}

	convIDStr := r.URL.Query().Get("id")
	convID, err := strconv.Atoi(convIDStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Invalid conversation ID"})
		return
	}

	// Verify user is member
	conv, err := h.repo.GetConversationByID(convID, userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Conversation not found"})
		return
	}

	if conv.Type == "group" {
		// For groups, just leave (remove member)
		err = h.repo.RemoveMemberFromGroup(convID, userID)
	} else {
		// For direct chats, delete the conversation entirely
		err = h.repo.DeleteConversation(convID)
	}

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Failed to delete conversation"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// POST /api/messages/react
func (h *ChatHandler) ReactToMessage(w http.ResponseWriter, r *http.Request) {
	userID := h.requireAuth(w, r)
	if userID == 0 {
		return
	}

	msgIDStr := r.URL.Query().Get("msg_id")
	msgID, err := strconv.Atoi(msgIDStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Invalid message ID"})
		return
	}

	var req struct {
		Emoji string `json:"emoji"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Emoji == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Emoji required"})
		return
	}

	if err := h.repo.AddReaction(msgID, userID, req.Emoji); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Failed to add reaction"})
		return
	}

	// Get updated reactions
	reactions, _ := h.repo.GetMessageReactions(msgID)

	// Broadcast reaction to conversation members
	// TODO: Get conversation ID from message and broadcast

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "data": reactions})
}

// DELETE /api/messages/react
func (h *ChatHandler) RemoveReaction(w http.ResponseWriter, r *http.Request) {
	userID := h.requireAuth(w, r)
	if userID == 0 {
		return
	}

	msgIDStr := r.URL.Query().Get("msg_id")
	msgID, err := strconv.Atoi(msgIDStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Invalid message ID"})
		return
	}

	emoji := r.URL.Query().Get("emoji")
	if emoji == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Emoji required"})
		return
	}

	if err := h.repo.RemoveReaction(msgID, userID, emoji); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Failed to remove reaction"})
		return
	}

	// Get updated reactions
	reactions, _ := h.repo.GetMessageReactions(msgID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "data": reactions})
}

// POST /api/messages/forward
func (h *ChatHandler) ForwardMessage(w http.ResponseWriter, r *http.Request) {
	userID := h.requireAuth(w, r)
	if userID == 0 {
		return
	}

	var req struct {
		MessageID     int   `json:"message_id"`
		TargetConvIDs []int `json:"target_conversation_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Invalid request"})
		return
	}

	if len(req.TargetConvIDs) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Target conversations required"})
		return
	}

	var forwardedMessages []*models.Message
	for _, targetConvID := range req.TargetConvIDs {
		// Verify user is member of target conversation
		_, err := h.repo.GetConversationByID(targetConvID, userID)
		if err != nil {
			continue // Skip if not a member
		}

		msg, err := h.repo.ForwardMessage(req.MessageID, targetConvID, userID)
		if err != nil {
			continue
		}

		// Get sender info
		sender, _ := h.userRepo.GetByID(userID)
		msg.Sender = sender

		forwardedMessages = append(forwardedMessages, msg)

		// Broadcast to conversation members
		memberIDs, _ := h.repo.GetConversationMemberIDs(targetConvID)
		payload, _ := json.Marshal(msg)
		h.hub.BroadcastToUsers(memberIDs, &websocket.Event{
			Type:           websocket.EventTypeMessage,
			ConversationID: targetConvID,
			UserID:         userID,
			Payload:        payload,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    forwardedMessages,
		"count":   len(forwardedMessages),
	})
}

// DELETE /api/messages/delete
func (h *ChatHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	userID := h.requireAuth(w, r)
	if userID == 0 {
		return
	}

	msgIDStr := r.URL.Query().Get("msg_id")
	msgID, err := strconv.Atoi(msgIDStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Invalid message ID"})
		return
	}

	// Get conversation ID before deleting (for broadcasting)
	convID, err := h.repo.GetMessageConversationID(msgID, userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Message not found"})
		return
	}

	if err := h.repo.DeleteMessage(msgID, userID); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Failed to delete message"})
		return
	}

	// Broadcast deletion to conversation members
	memberIDs, _ := h.repo.GetConversationMemberIDs(convID)
	if len(memberIDs) > 0 {
		payload, _ := json.Marshal(map[string]interface{}{
			"id":              msgID,
			"conversation_id": convID,
			"deleted_at":      time.Now(),
			"content":         "Message deleted",
		})
		h.hub.BroadcastToUsers(memberIDs, &websocket.Event{
			Type:           websocket.EventTypeMessageDelete,
			ConversationID: convID,
			Payload:        payload,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}
