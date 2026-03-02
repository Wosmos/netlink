package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"netlink/middleware"
	"netlink/models"
	"netlink/repository"
	"netlink/websocket"

	ws "github.com/gorilla/websocket"
)

type ChatHandler struct {
	repo           ChatRepoInterface
	userRepo       UserRepoInterface
	authService    AuthServiceInterface
	hub            HubInterface
	allowedOrigins map[string]bool
}

func NewChatHandler(repo ChatRepoInterface, userRepo UserRepoInterface, authService AuthServiceInterface, hub HubInterface, allowedOrigins ...map[string]bool) *ChatHandler {
	origins := map[string]bool{}
	if len(allowedOrigins) > 0 {
		origins = allowedOrigins[0]
	}
	return &ChatHandler{
		repo:           repo,
		userRepo:       userRepo,
		authService:    authService,
		hub:            hub,
		allowedOrigins: origins,
	}
}

func (h *ChatHandler) requireAuth(w http.ResponseWriter, r *http.Request) int {
	user, err := h.authService.GetUserFromRequest(r)
	if err != nil || user == nil {
		middleware.JSONError(w, "Unauthorized", http.StatusUnauthorized)
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
		middleware.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	upgrader := ws.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			if len(h.allowedOrigins) == 0 {
				return true // dev mode fallback
			}
			return h.allowedOrigins[origin]
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	concreteHub, ok := h.hub.(*websocket.Hub)
	if !ok {
		return
	}
	client := websocket.NewClient(concreteHub, conn, user.ID)
	h.hub.Register(client)

	// Update last seen (best-effort, non-critical)
	if err := h.repo.UpdateLastSeen(user.ID); err != nil {
		log.Printf("Failed to update last seen for user %d: %v", user.ID, err)
	}

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
		middleware.JSONError(w, "Failed to get conversations", http.StatusInternalServerError)
		return
	}

	if conversations == nil {
		conversations = []models.Conversation{}
	}

	middleware.JSONSuccess(w, conversations)
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
		middleware.JSONError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var targetUserID int
	if req.UserID > 0 {
		targetUserID = req.UserID
	} else if req.Phone != "" {
		user, err := h.userRepo.GetByPhone(req.Phone)
		if err != nil || user == nil {
			middleware.JSONError(w, "User not found", http.StatusNotFound)
			return
		}
		targetUserID = user.ID
	} else if req.Email != "" {
		user, err := h.userRepo.GetByEmail(req.Email)
		if err != nil || user == nil {
			middleware.JSONError(w, "User not found", http.StatusNotFound)
			return
		}
		targetUserID = user.ID
	} else {
		middleware.JSONError(w, "Provide user_id, phone, or email", http.StatusBadRequest)
		return
	}

	if targetUserID == userID {
		middleware.JSONError(w, "Cannot chat with yourself", http.StatusBadRequest)
		return
	}

	conv, err := h.repo.CreateDirectConversation(userID, targetUserID)
	if err != nil {
		middleware.JSONError(w, "Failed to create conversation", http.StatusInternalServerError)
		return
	}

	// Notify both users about new conversation
	payload, _ := json.Marshal(conv)
	h.hub.BroadcastToUsers([]int{userID, targetUserID}, &websocket.Event{
		Type:           websocket.EventTypeConversation,
		ConversationID: conv.ID,
		Payload:        payload,
	})

	middleware.JSONCreated(w, conv)
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
		middleware.JSONError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Name == "" || len(req.Name) > 100 {
		middleware.JSONError(w, "Group name required (max 100 chars)", http.StatusBadRequest)
		return
	}

	conv, err := h.repo.CreateGroupConversation(req.Name, userID, req.MemberIDs)
	if err != nil {
		middleware.JSONError(w, "Failed to create group", http.StatusInternalServerError)
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

	middleware.JSONCreated(w, conv)
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
		middleware.JSONError(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}

	// Verify user is member
	_, err = h.repo.GetConversationByID(convID, userID)
	if err != nil {
		middleware.JSONError(w, "Conversation not found", http.StatusNotFound)
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
		middleware.JSONError(w, "Failed to get messages", http.StatusInternalServerError)
		return
	}

	if messages == nil {
		messages = []models.Message{}
	}

	middleware.JSONSuccess(w, messages)
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
		middleware.JSONError(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}

	// Verify user is member
	_, err = h.repo.GetConversationByID(convID, userID)
	if err != nil {
		middleware.JSONError(w, "Conversation not found", http.StatusNotFound)
		return
	}

	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSONError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Content == "" || len(req.Content) > 10000 {
		middleware.JSONError(w, "Message content required (max 10000 chars)", http.StatusBadRequest)
		return
	}

	msgType := models.MessageTypeText
	if req.Type != "" {
		msgType = models.MessageType(req.Type)
	}

	var voiceParams []repository.CreateMessageParams
	if msgType == models.MessageTypeVoice && req.VoiceFilePath != "" {
		voiceParams = append(voiceParams, repository.CreateMessageParams{
			VoiceFilePath: req.VoiceFilePath,
			VoiceDuration: req.VoiceDuration,
			VoiceWaveform: req.VoiceWaveform,
			VoiceFileSize: req.VoiceFileSize,
		})
	}

	msg, err := h.repo.CreateMessage(convID, userID, msgType, req.Content, req.ReplyToID, voiceParams...)
	if err != nil {
		middleware.JSONError(w, "Failed to send message", http.StatusInternalServerError)
		return
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

	middleware.JSONCreated(w, msg)
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
		middleware.JSONError(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}

	msgIDStr := r.URL.Query().Get("msg_id")
	msgID, err := strconv.Atoi(msgIDStr)
	if err != nil {
		middleware.JSONError(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	// Verify user is member of conversation
	_, err = h.repo.GetConversationByID(convID, userID)
	if err != nil {
		middleware.JSONError(w, "Conversation not found", http.StatusNotFound)
		return
	}

	var req EditMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSONError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Content == "" || len(req.Content) > 10000 {
		middleware.JSONError(w, "Message content required (max 10000 chars)", http.StatusBadRequest)
		return
	}

	msg, err := h.repo.UpdateMessage(msgID, userID, req.Content)
	if err != nil {
		middleware.JSONError(w, "Failed to update message", http.StatusInternalServerError)
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

	middleware.JSONSuccess(w, msg)
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
		middleware.JSONError(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}

	if err := h.repo.MarkAsRead(convID, userID); err != nil {
		middleware.JSONError(w, "Failed to mark as read", http.StatusInternalServerError)
		return
	}

	// Broadcast read receipt
	memberIDs, _ := h.repo.GetConversationMemberIDs(convID)
	h.hub.BroadcastToUsers(memberIDs, &websocket.Event{
		Type:           websocket.EventTypeRead,
		ConversationID: convID,
		UserID:         userID,
	})

	middleware.JSONOk(w)
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
		middleware.JSONError(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}

	memberIDs, _ := h.repo.GetConversationMemberIDs(convID)
	h.hub.BroadcastToUsers(memberIDs, &websocket.Event{
		Type:           websocket.EventTypeTyping,
		ConversationID: convID,
		UserID:         userID,
	})

	middleware.JSONOk(w)
}

// GET /api/users/online
func (h *ChatHandler) GetOnlineUsers(w http.ResponseWriter, r *http.Request) {
	userID := h.requireAuth(w, r)
	if userID == 0 {
		return
	}

	onlineUsers := h.hub.GetOnlineUsers()

	middleware.JSON(w, http.StatusOK, map[string]interface{}{
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
		middleware.JSONError(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}

	// Verify user is member
	conv, err := h.repo.GetConversationByID(convID, userID)
	if err != nil {
		middleware.JSONError(w, "Conversation not found", http.StatusNotFound)
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
		middleware.JSONError(w, "Failed to delete conversation", http.StatusInternalServerError)
		return
	}

	middleware.JSONOk(w)
}

// POST /api/messages/react - Add or toggle reaction (optimized for 20-30ms)
func (h *ChatHandler) ReactToMessage(w http.ResponseWriter, r *http.Request) {
	userID := h.requireAuth(w, r)
	if userID == 0 {
		return
	}

	msgIDStr := r.URL.Query().Get("msg_id")
	msgID, err := strconv.Atoi(msgIDStr)
	if err != nil {
		middleware.JSONError(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Emoji     string `json:"emoji"`
		IsCustom  bool   `json:"is_custom"`
		CustomURL string `json:"custom_url"`
		Toggle    bool   `json:"toggle"` // If true, toggle reaction on/off
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Emoji == "" {
		middleware.JSONError(w, "Emoji required", http.StatusBadRequest)
		return
	}

	// Get conversation ID for broadcasting
	convID, err := h.repo.GetMessageConversationID(msgID, userID)
	if err != nil {
		// Try without user check (user might not be sender)
		convID, err = h.repo.GetConversationIDByMessageID(msgID)
		if err != nil {
			middleware.JSONError(w, "Message not found", http.StatusNotFound)
			return
		}
	}

	var added bool
	if req.Toggle {
		// Use optimized toggle operation
		added, err = h.repo.ToggleReaction(msgID, userID, req.Emoji, req.IsCustom, req.CustomURL)
	} else {
		// Add reaction
		if req.IsCustom {
			err = h.repo.AddCustomReaction(msgID, userID, req.Emoji, req.CustomURL)
		} else {
			err = h.repo.AddReaction(msgID, userID, req.Emoji)
		}
		added = true
	}

	if err != nil {
		middleware.JSONError(w, "Failed to update reaction", http.StatusInternalServerError)
		return
	}

	// Get updated reactions summary (optimized single query)
	reactions, _ := h.repo.GetMessageReactionsSummary(msgID)

	// Broadcast reaction update to conversation members (optimized payload)
	memberIDs, _ := h.repo.GetConversationMemberIDs(convID)
	if len(memberIDs) > 0 {
		payload, _ := json.Marshal(map[string]interface{}{
			"message_id": msgID,
			"user_id":    userID,
			"emoji":      req.Emoji,
			"is_custom":  req.IsCustom,
			"custom_url": req.CustomURL,
			"added":      added,
			"reactions":  reactions,
		})
		h.hub.BroadcastToUsers(memberIDs, &websocket.Event{
			Type:           websocket.EventTypeReaction,
			ConversationID: convID,
			UserID:         userID,
			Payload:        payload,
		})
	}

	middleware.JSON(w, http.StatusOK, map[string]interface{}{
		"success":   true,
		"added":     added,
		"reactions": reactions,
	})
}

// DELETE /api/messages/react - Remove reaction
func (h *ChatHandler) RemoveReaction(w http.ResponseWriter, r *http.Request) {
	userID := h.requireAuth(w, r)
	if userID == 0 {
		return
	}

	msgIDStr := r.URL.Query().Get("msg_id")
	msgID, err := strconv.Atoi(msgIDStr)
	if err != nil {
		middleware.JSONError(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	emoji := r.URL.Query().Get("emoji")
	if emoji == "" {
		middleware.JSONError(w, "Emoji required", http.StatusBadRequest)
		return
	}

	// Get conversation ID for broadcasting
	convID, err := h.repo.GetConversationIDByMessageID(msgID)
	if err != nil {
		middleware.JSONError(w, "Message not found", http.StatusNotFound)
		return
	}

	if err := h.repo.RemoveReaction(msgID, userID, emoji); err != nil {
		middleware.JSONError(w, "Failed to remove reaction", http.StatusInternalServerError)
		return
	}

	// Get updated reactions
	reactions, _ := h.repo.GetMessageReactionsSummary(msgID)

	// Broadcast reaction removal
	memberIDs, _ := h.repo.GetConversationMemberIDs(convID)
	if len(memberIDs) > 0 {
		payload, _ := json.Marshal(map[string]interface{}{
			"message_id": msgID,
			"user_id":    userID,
			"emoji":      emoji,
			"added":      false,
			"reactions":  reactions,
		})
		h.hub.BroadcastToUsers(memberIDs, &websocket.Event{
			Type:           websocket.EventTypeReaction,
			ConversationID: convID,
			UserID:         userID,
			Payload:        payload,
		})
	}

	middleware.JSON(w, http.StatusOK, map[string]interface{}{
		"success":   true,
		"reactions": reactions,
	})
}

// GET /api/messages/reactions - Get reactions for a message
func (h *ChatHandler) GetMessageReactions(w http.ResponseWriter, r *http.Request) {
	userID := h.requireAuth(w, r)
	if userID == 0 {
		return
	}

	msgIDStr := r.URL.Query().Get("msg_id")
	msgID, err := strconv.Atoi(msgIDStr)
	if err != nil {
		middleware.JSONError(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	reactions, err := h.repo.GetMessageReactionsSummary(msgID)
	if err != nil {
		middleware.JSONError(w, "Failed to get reactions", http.StatusInternalServerError)
		return
	}

	middleware.JSONSuccess(w, reactions)
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
		middleware.JSONError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if len(req.TargetConvIDs) == 0 || len(req.TargetConvIDs) > 20 {
		middleware.JSONError(w, "Provide 1-20 target conversations", http.StatusBadRequest)
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

	middleware.JSON(w, http.StatusOK, map[string]interface{}{
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
		middleware.JSONError(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	// Get conversation ID before deleting (for broadcasting)
	convID, err := h.repo.GetMessageConversationID(msgID, userID)
	if err != nil {
		middleware.JSONError(w, "Message not found", http.StatusNotFound)
		return
	}

	if err := h.repo.DeleteMessage(msgID, userID); err != nil {
		middleware.JSONError(w, "Failed to delete message", http.StatusInternalServerError)
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

	middleware.JSONOk(w)
}
