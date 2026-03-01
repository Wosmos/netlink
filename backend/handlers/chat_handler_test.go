package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"netlink/models"
	"netlink/websocket"
)

// mockChatRepo implements ChatRepoInterface for testing
type mockChatRepo struct {
	getUserConversationsFn       func(userID int) ([]models.Conversation, error)
	createDirectConversationFn   func(userID1, userID2 int) (*models.Conversation, error)
	createGroupConversationFn    func(name string, creatorID int, memberIDs []int) (*models.Conversation, error)
	getConversationByIDFn        func(convID, userID int) (*models.Conversation, error)
	getConversationMemberIDsFn   func(convID int) ([]int, error)
	getMessagesFn                func(convID int, limit, offset int) ([]models.Message, error)
	createMessageFn              func(convID, senderID int, msgType models.MessageType, content string, replyToID *int) (*models.Message, error)
	updateMessageFn              func(msgID, userID int, content string) (*models.Message, error)
	markAsReadFn                 func(convID, userID int) error
	updateLastSeenFn             func(userID int) error
	deleteConversationFn         func(convID int) error
	removeMemberFromGroupFn      func(convID, userID int) error
	addReactionFn                func(messageID, userID int, emoji string) error
	addCustomReactionFn          func(messageID, userID int, emoji, customURL string) error
	removeReactionFn             func(messageID, userID int, emoji string) error
	getMessageReactionsSummaryFn func(messageID int) ([]models.ReactionSummary, error)
	toggleReactionFn             func(messageID, userID int, emoji string, isCustom bool, customURL string) (bool, error)
	forwardMessageFn             func(originalMsgID, targetConvID, senderID int) (*models.Message, error)
	deleteMessageFn              func(messageID, userID int) error
	getMessageConvIDFn           func(messageID, userID int) (int, error)
	getConvIDByMsgIDFn           func(messageID int) (int, error)
}

func (m *mockChatRepo) GetUserConversations(userID int) ([]models.Conversation, error) {
	if m.getUserConversationsFn != nil {
		return m.getUserConversationsFn(userID)
	}
	return nil, nil
}

func (m *mockChatRepo) CreateDirectConversation(userID1, userID2 int) (*models.Conversation, error) {
	if m.createDirectConversationFn != nil {
		return m.createDirectConversationFn(userID1, userID2)
	}
	return nil, nil
}

func (m *mockChatRepo) CreateGroupConversation(name string, creatorID int, memberIDs []int) (*models.Conversation, error) {
	if m.createGroupConversationFn != nil {
		return m.createGroupConversationFn(name, creatorID, memberIDs)
	}
	return nil, nil
}

func (m *mockChatRepo) GetConversationByID(convID, userID int) (*models.Conversation, error) {
	if m.getConversationByIDFn != nil {
		return m.getConversationByIDFn(convID, userID)
	}
	return &models.Conversation{ID: convID}, nil
}

func (m *mockChatRepo) GetConversationMemberIDs(convID int) ([]int, error) {
	if m.getConversationMemberIDsFn != nil {
		return m.getConversationMemberIDsFn(convID)
	}
	return []int{1, 2}, nil
}

func (m *mockChatRepo) GetMessages(convID int, limit, offset int) ([]models.Message, error) {
	if m.getMessagesFn != nil {
		return m.getMessagesFn(convID, limit, offset)
	}
	return nil, nil
}

func (m *mockChatRepo) CreateMessage(convID, senderID int, msgType models.MessageType, content string, replyToID *int) (*models.Message, error) {
	if m.createMessageFn != nil {
		return m.createMessageFn(convID, senderID, msgType, content, replyToID)
	}
	return &models.Message{ID: 1, ConversationID: convID, SenderID: senderID, Content: content}, nil
}

func (m *mockChatRepo) UpdateMessage(msgID, userID int, content string) (*models.Message, error) {
	if m.updateMessageFn != nil {
		return m.updateMessageFn(msgID, userID, content)
	}
	return nil, nil
}

func (m *mockChatRepo) MarkAsRead(convID, userID int) error {
	if m.markAsReadFn != nil {
		return m.markAsReadFn(convID, userID)
	}
	return nil
}

func (m *mockChatRepo) UpdateLastSeen(userID int) error {
	if m.updateLastSeenFn != nil {
		return m.updateLastSeenFn(userID)
	}
	return nil
}

func (m *mockChatRepo) DeleteConversation(convID int) error {
	if m.deleteConversationFn != nil {
		return m.deleteConversationFn(convID)
	}
	return nil
}

func (m *mockChatRepo) RemoveMemberFromGroup(convID, userID int) error {
	if m.removeMemberFromGroupFn != nil {
		return m.removeMemberFromGroupFn(convID, userID)
	}
	return nil
}

func (m *mockChatRepo) AddReaction(messageID, userID int, emoji string) error {
	if m.addReactionFn != nil {
		return m.addReactionFn(messageID, userID, emoji)
	}
	return nil
}

func (m *mockChatRepo) AddCustomReaction(messageID, userID int, emoji, customURL string) error {
	if m.addCustomReactionFn != nil {
		return m.addCustomReactionFn(messageID, userID, emoji, customURL)
	}
	return nil
}

func (m *mockChatRepo) RemoveReaction(messageID, userID int, emoji string) error {
	if m.removeReactionFn != nil {
		return m.removeReactionFn(messageID, userID, emoji)
	}
	return nil
}

func (m *mockChatRepo) GetMessageReactionsSummary(messageID int) ([]models.ReactionSummary, error) {
	if m.getMessageReactionsSummaryFn != nil {
		return m.getMessageReactionsSummaryFn(messageID)
	}
	return nil, nil
}

func (m *mockChatRepo) ToggleReaction(messageID, userID int, emoji string, isCustom bool, customURL string) (bool, error) {
	if m.toggleReactionFn != nil {
		return m.toggleReactionFn(messageID, userID, emoji, isCustom, customURL)
	}
	return true, nil
}

func (m *mockChatRepo) ForwardMessage(originalMsgID, targetConvID, senderID int) (*models.Message, error) {
	if m.forwardMessageFn != nil {
		return m.forwardMessageFn(originalMsgID, targetConvID, senderID)
	}
	return nil, nil
}

func (m *mockChatRepo) DeleteMessage(messageID, userID int) error {
	if m.deleteMessageFn != nil {
		return m.deleteMessageFn(messageID, userID)
	}
	return nil
}

func (m *mockChatRepo) GetMessageConversationID(messageID, userID int) (int, error) {
	if m.getMessageConvIDFn != nil {
		return m.getMessageConvIDFn(messageID, userID)
	}
	return 0, errors.New("not found")
}

func (m *mockChatRepo) GetConversationIDByMessageID(messageID int) (int, error) {
	if m.getConvIDByMsgIDFn != nil {
		return m.getConvIDByMsgIDFn(messageID)
	}
	return 0, errors.New("not found")
}

// mockUserRepo implements UserRepoInterface for testing
type mockUserRepo struct {
	getByIDFn    func(id int) (*models.User, error)
	getByEmailFn func(email string) (*models.User, error)
	getByPhoneFn func(phone string) (*models.User, error)
}

func (m *mockUserRepo) GetByID(id int) (*models.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	return &models.User{ID: id, Email: "user@test.com"}, nil
}

func (m *mockUserRepo) GetByEmail(email string) (*models.User, error) {
	if m.getByEmailFn != nil {
		return m.getByEmailFn(email)
	}
	return nil, nil
}

func (m *mockUserRepo) GetByPhone(phone string) (*models.User, error) {
	if m.getByPhoneFn != nil {
		return m.getByPhoneFn(phone)
	}
	return nil, nil
}

// mockHub implements HubInterface for testing
type mockHub struct {
	broadcastedEvents []*websocket.Event
	onlineUsers       []int
}

func (m *mockHub) BroadcastToUsers(userIDs []int, event *websocket.Event) {
	m.broadcastedEvents = append(m.broadcastedEvents, event)
}

func (m *mockHub) GetOnlineUsers() []int {
	return m.onlineUsers
}

func (m *mockHub) Register(client *websocket.Client) {}

// helper to create an authenticated user request
func authedRequest(method, url string, body string) (*http.Request, *mockAuthService) {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, url, bytes.NewBufferString(body))
	} else {
		req = httptest.NewRequest(method, url, nil)
	}

	mock := &mockAuthService{
		getUserFromReqFn: func(r *http.Request) (*models.User, error) {
			return &models.User{ID: 1, Email: "user@test.com"}, nil
		},
	}

	return req, mock
}

func decodeChatResponse(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	return resp
}

// =========== ListConversations Tests ===========

func TestListConversations_Success(t *testing.T) {
	convs := []models.Conversation{
		{ID: 1, Type: "direct", Name: "", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: 2, Type: "group", Name: "Team Chat", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	chatRepo := &mockChatRepo{
		getUserConversationsFn: func(userID int) ([]models.Conversation, error) {
			if userID != 1 {
				t.Errorf("Expected userID=1, got %d", userID)
			}
			return convs, nil
		},
	}

	req, authMock := authedRequest(http.MethodGet, "/api/conversations", "")
	hub := &mockHub{}
	h := NewChatHandler(chatRepo, &mockUserRepo{}, authMock, hub)
	w := httptest.NewRecorder()

	h.ListConversations(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	resp := decodeChatResponse(t, w)
	if resp["success"] != true {
		t.Error("Expected success=true")
	}
	data, ok := resp["data"].([]interface{})
	if !ok {
		t.Fatal("Expected data to be an array")
	}
	if len(data) != 2 {
		t.Errorf("Expected 2 conversations, got %d", len(data))
	}
}

func TestListConversations_Empty(t *testing.T) {
	chatRepo := &mockChatRepo{
		getUserConversationsFn: func(userID int) ([]models.Conversation, error) {
			return nil, nil
		},
	}

	req, authMock := authedRequest(http.MethodGet, "/api/conversations", "")
	h := NewChatHandler(chatRepo, &mockUserRepo{}, authMock, &mockHub{})
	w := httptest.NewRecorder()

	h.ListConversations(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	resp := decodeChatResponse(t, w)
	data := resp["data"].([]interface{})
	if len(data) != 0 {
		t.Errorf("Expected empty array, got %d items", len(data))
	}
}

func TestListConversations_Unauthorized(t *testing.T) {
	unauthMock := &mockAuthService{
		getUserFromReqFn: func(r *http.Request) (*models.User, error) {
			return nil, nil
		},
	}

	h := NewChatHandler(&mockChatRepo{}, &mockUserRepo{}, unauthMock, &mockHub{})
	req := httptest.NewRequest(http.MethodGet, "/api/conversations", nil)
	w := httptest.NewRecorder()

	h.ListConversations(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestListConversations_DBError(t *testing.T) {
	chatRepo := &mockChatRepo{
		getUserConversationsFn: func(userID int) ([]models.Conversation, error) {
			return nil, errors.New("database connection lost")
		},
	}

	req, authMock := authedRequest(http.MethodGet, "/api/conversations", "")
	h := NewChatHandler(chatRepo, &mockUserRepo{}, authMock, &mockHub{})
	w := httptest.NewRecorder()

	h.ListConversations(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

// =========== SendMessage Tests ===========

func TestSendMessage_Success(t *testing.T) {
	msg := &models.Message{
		ID:             10,
		ConversationID: 5,
		SenderID:       1,
		Type:           models.MessageTypeText,
		Content:        "Hello!",
		CreatedAt:      time.Now(),
	}

	chatRepo := &mockChatRepo{
		createMessageFn: func(convID, senderID int, msgType models.MessageType, content string, replyToID *int) (*models.Message, error) {
			if convID != 5 || senderID != 1 || content != "Hello!" {
				t.Errorf("Unexpected args: convID=%d, senderID=%d, content=%q", convID, senderID, content)
			}
			return msg, nil
		},
	}

	hub := &mockHub{}
	req, authMock := authedRequest(http.MethodPost, "/api/conversations/messages?id=5", `{"content":"Hello!"}`)
	h := NewChatHandler(chatRepo, &mockUserRepo{}, authMock, hub)
	w := httptest.NewRecorder()

	h.SendMessage(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	resp := decodeChatResponse(t, w)
	if resp["success"] != true {
		t.Error("Expected success=true")
	}

	// Verify WebSocket broadcast happened
	if len(hub.broadcastedEvents) != 1 {
		t.Errorf("Expected 1 broadcast event, got %d", len(hub.broadcastedEvents))
	}
	if hub.broadcastedEvents[0].Type != websocket.EventTypeMessage {
		t.Errorf("Expected event type=%q, got %q", websocket.EventTypeMessage, hub.broadcastedEvents[0].Type)
	}
}

func TestSendMessage_EmptyContent(t *testing.T) {
	req, authMock := authedRequest(http.MethodPost, "/api/conversations/messages?id=5", `{"content":""}`)
	h := NewChatHandler(&mockChatRepo{}, &mockUserRepo{}, authMock, &mockHub{})
	w := httptest.NewRecorder()

	h.SendMessage(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestSendMessage_InvalidConvID(t *testing.T) {
	req, authMock := authedRequest(http.MethodPost, "/api/conversations/messages?id=abc", `{"content":"hello"}`)
	h := NewChatHandler(&mockChatRepo{}, &mockUserRepo{}, authMock, &mockHub{})
	w := httptest.NewRecorder()

	h.SendMessage(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestSendMessage_NotMember(t *testing.T) {
	chatRepo := &mockChatRepo{
		getConversationByIDFn: func(convID, userID int) (*models.Conversation, error) {
			return nil, errors.New("not found")
		},
	}

	req, authMock := authedRequest(http.MethodPost, "/api/conversations/messages?id=999", `{"content":"hello"}`)
	h := NewChatHandler(chatRepo, &mockUserRepo{}, authMock, &mockHub{})
	w := httptest.NewRecorder()

	h.SendMessage(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

// =========== CreateDirectChat Tests ===========

func TestCreateDirectChat_Success(t *testing.T) {
	conv := &models.Conversation{
		ID:        10,
		Type:      "direct",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	chatRepo := &mockChatRepo{
		createDirectConversationFn: func(userID1, userID2 int) (*models.Conversation, error) {
			if userID1 != 1 || userID2 != 2 {
				t.Errorf("Expected users 1,2 but got %d,%d", userID1, userID2)
			}
			return conv, nil
		},
	}

	hub := &mockHub{}
	req, authMock := authedRequest(http.MethodPost, "/api/conversations/direct", `{"user_id":2}`)
	h := NewChatHandler(chatRepo, &mockUserRepo{}, authMock, hub)
	w := httptest.NewRecorder()

	h.CreateDirectChat(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	// Verify broadcast was sent
	if len(hub.broadcastedEvents) != 1 {
		t.Errorf("Expected 1 broadcast, got %d", len(hub.broadcastedEvents))
	}
}

func TestCreateDirectChat_ByEmail(t *testing.T) {
	targetUser := &models.User{ID: 5, Email: "target@test.com"}
	conv := &models.Conversation{ID: 10, Type: "direct"}

	userRepo := &mockUserRepo{
		getByEmailFn: func(email string) (*models.User, error) {
			if email != "target@test.com" {
				t.Errorf("Expected email=%q, got %q", "target@test.com", email)
			}
			return targetUser, nil
		},
	}

	chatRepo := &mockChatRepo{
		createDirectConversationFn: func(userID1, userID2 int) (*models.Conversation, error) {
			return conv, nil
		},
	}

	req, authMock := authedRequest(http.MethodPost, "/api/conversations/direct", `{"email":"target@test.com"}`)
	h := NewChatHandler(chatRepo, userRepo, authMock, &mockHub{})
	w := httptest.NewRecorder()

	h.CreateDirectChat(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}
}

func TestCreateDirectChat_ByPhone(t *testing.T) {
	targetUser := &models.User{ID: 5, Phone: "1234567890"}
	conv := &models.Conversation{ID: 10, Type: "direct"}

	userRepo := &mockUserRepo{
		getByPhoneFn: func(phone string) (*models.User, error) {
			return targetUser, nil
		},
	}

	chatRepo := &mockChatRepo{
		createDirectConversationFn: func(userID1, userID2 int) (*models.Conversation, error) {
			return conv, nil
		},
	}

	req, authMock := authedRequest(http.MethodPost, "/api/conversations/direct", `{"phone":"1234567890"}`)
	h := NewChatHandler(chatRepo, userRepo, authMock, &mockHub{})
	w := httptest.NewRecorder()

	h.CreateDirectChat(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}
}

func TestCreateDirectChat_SelfChat(t *testing.T) {
	req, authMock := authedRequest(http.MethodPost, "/api/conversations/direct", `{"user_id":1}`)
	h := NewChatHandler(&mockChatRepo{}, &mockUserRepo{}, authMock, &mockHub{})
	w := httptest.NewRecorder()

	h.CreateDirectChat(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	resp := decodeChatResponse(t, w)
	if resp["error"] != "Cannot chat with yourself" {
		t.Errorf("Expected self-chat error, got %v", resp["error"])
	}
}

func TestCreateDirectChat_UserNotFound(t *testing.T) {
	userRepo := &mockUserRepo{
		getByEmailFn: func(email string) (*models.User, error) {
			return nil, nil
		},
	}

	req, authMock := authedRequest(http.MethodPost, "/api/conversations/direct", `{"email":"nobody@test.com"}`)
	h := NewChatHandler(&mockChatRepo{}, userRepo, authMock, &mockHub{})
	w := httptest.NewRecorder()

	h.CreateDirectChat(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestCreateDirectChat_NoIdentifier(t *testing.T) {
	req, authMock := authedRequest(http.MethodPost, "/api/conversations/direct", `{}`)
	h := NewChatHandler(&mockChatRepo{}, &mockUserRepo{}, authMock, &mockHub{})
	w := httptest.NewRecorder()

	h.CreateDirectChat(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// =========== CreateGroup Tests ===========

func TestCreateGroup_Success(t *testing.T) {
	conv := &models.Conversation{ID: 20, Type: "group", Name: "Dev Team"}

	chatRepo := &mockChatRepo{
		createGroupConversationFn: func(name string, creatorID int, memberIDs []int) (*models.Conversation, error) {
			if name != "Dev Team" || creatorID != 1 {
				t.Errorf("Unexpected args: name=%q, creatorID=%d", name, creatorID)
			}
			return conv, nil
		},
	}

	hub := &mockHub{}
	req, authMock := authedRequest(http.MethodPost, "/api/conversations/group", `{"name":"Dev Team","member_ids":[2,3]}`)
	h := NewChatHandler(chatRepo, &mockUserRepo{}, authMock, hub)
	w := httptest.NewRecorder()

	h.CreateGroup(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}
}

func TestCreateGroup_EmptyName(t *testing.T) {
	req, authMock := authedRequest(http.MethodPost, "/api/conversations/group", `{"name":"","member_ids":[2]}`)
	h := NewChatHandler(&mockChatRepo{}, &mockUserRepo{}, authMock, &mockHub{})
	w := httptest.NewRecorder()

	h.CreateGroup(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// =========== GetMessages Tests ===========

func TestGetMessages_Success(t *testing.T) {
	messages := []models.Message{
		{ID: 1, ConversationID: 5, SenderID: 1, Content: "Hello", Type: models.MessageTypeText},
		{ID: 2, ConversationID: 5, SenderID: 2, Content: "Hi!", Type: models.MessageTypeText},
	}

	chatRepo := &mockChatRepo{
		getMessagesFn: func(convID int, limit, offset int) ([]models.Message, error) {
			if convID != 5 {
				t.Errorf("Expected convID=5, got %d", convID)
			}
			return messages, nil
		},
	}

	req, authMock := authedRequest(http.MethodGet, "/api/conversations/messages?id=5", "")
	h := NewChatHandler(chatRepo, &mockUserRepo{}, authMock, &mockHub{})
	w := httptest.NewRecorder()

	h.GetMessages(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	resp := decodeChatResponse(t, w)
	data := resp["data"].([]interface{})
	if len(data) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(data))
	}
}

func TestGetMessages_InvalidConvID(t *testing.T) {
	req, authMock := authedRequest(http.MethodGet, "/api/conversations/messages?id=abc", "")
	h := NewChatHandler(&mockChatRepo{}, &mockUserRepo{}, authMock, &mockHub{})
	w := httptest.NewRecorder()

	h.GetMessages(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// =========== MarkAsRead Tests ===========

func TestMarkAsRead_Success(t *testing.T) {
	var markedConvID, markedUserID int
	chatRepo := &mockChatRepo{
		markAsReadFn: func(convID, userID int) error {
			markedConvID = convID
			markedUserID = userID
			return nil
		},
	}

	hub := &mockHub{}
	req, authMock := authedRequest(http.MethodPost, "/api/conversations/read?id=5", "")
	h := NewChatHandler(chatRepo, &mockUserRepo{}, authMock, hub)
	w := httptest.NewRecorder()

	h.MarkAsRead(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if markedConvID != 5 || markedUserID != 1 {
		t.Errorf("Expected mark convID=5 userID=1, got convID=%d userID=%d", markedConvID, markedUserID)
	}

	// Verify read receipt broadcast
	if len(hub.broadcastedEvents) != 1 {
		t.Errorf("Expected 1 broadcast, got %d", len(hub.broadcastedEvents))
	}
	if hub.broadcastedEvents[0].Type != websocket.EventTypeRead {
		t.Errorf("Expected event type=%q, got %q", websocket.EventTypeRead, hub.broadcastedEvents[0].Type)
	}
}

// =========== DeleteMessage Tests ===========

func TestDeleteMessage_Success(t *testing.T) {
	chatRepo := &mockChatRepo{
		getMessageConvIDFn: func(messageID, userID int) (int, error) {
			return 5, nil
		},
		deleteMessageFn: func(messageID, userID int) error {
			if messageID != 10 || userID != 1 {
				t.Errorf("Unexpected args: msgID=%d, userID=%d", messageID, userID)
			}
			return nil
		},
	}

	hub := &mockHub{}
	req, authMock := authedRequest(http.MethodDelete, "/api/messages/delete?msg_id=10", "")
	h := NewChatHandler(chatRepo, &mockUserRepo{}, authMock, hub)
	w := httptest.NewRecorder()

	h.DeleteMessage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify deletion broadcast
	if len(hub.broadcastedEvents) != 1 {
		t.Errorf("Expected 1 broadcast, got %d", len(hub.broadcastedEvents))
	}
	if hub.broadcastedEvents[0].Type != websocket.EventTypeMessageDelete {
		t.Errorf("Expected event type=%q, got %q", websocket.EventTypeMessageDelete, hub.broadcastedEvents[0].Type)
	}
}

func TestDeleteMessage_NotFound(t *testing.T) {
	chatRepo := &mockChatRepo{
		getMessageConvIDFn: func(messageID, userID int) (int, error) {
			return 0, errors.New("no rows")
		},
	}

	req, authMock := authedRequest(http.MethodDelete, "/api/messages/delete?msg_id=999", "")
	h := NewChatHandler(chatRepo, &mockUserRepo{}, authMock, &mockHub{})
	w := httptest.NewRecorder()

	h.DeleteMessage(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

// =========== EditMessage Tests ===========

func TestEditMessage_Success(t *testing.T) {
	updatedMsg := &models.Message{
		ID:             10,
		ConversationID: 5,
		SenderID:       1,
		Content:        "Updated content",
		Type:           models.MessageTypeText,
	}

	chatRepo := &mockChatRepo{
		updateMessageFn: func(msgID, userID int, content string) (*models.Message, error) {
			if content != "Updated content" {
				t.Errorf("Expected content=%q, got %q", "Updated content", content)
			}
			return updatedMsg, nil
		},
	}

	hub := &mockHub{}
	req, authMock := authedRequest(http.MethodPut, "/api/conversations/messages/edit?id=5&msg_id=10", `{"content":"Updated content"}`)
	h := NewChatHandler(chatRepo, &mockUserRepo{}, authMock, hub)
	w := httptest.NewRecorder()

	h.EditMessage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify edit broadcast
	if len(hub.broadcastedEvents) != 1 {
		t.Errorf("Expected 1 broadcast, got %d", len(hub.broadcastedEvents))
	}
	if hub.broadcastedEvents[0].Type != websocket.EventTypeMessageEdit {
		t.Errorf("Expected event type=%q, got %q", websocket.EventTypeMessageEdit, hub.broadcastedEvents[0].Type)
	}
}

func TestEditMessage_EmptyContent(t *testing.T) {
	req, authMock := authedRequest(http.MethodPut, "/api/conversations/messages/edit?id=5&msg_id=10", `{"content":""}`)
	h := NewChatHandler(&mockChatRepo{}, &mockUserRepo{}, authMock, &mockHub{})
	w := httptest.NewRecorder()

	h.EditMessage(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// =========== DeleteConversation Tests ===========

func TestDeleteConversation_DirectChat(t *testing.T) {
	chatRepo := &mockChatRepo{
		getConversationByIDFn: func(convID, userID int) (*models.Conversation, error) {
			return &models.Conversation{ID: convID, Type: "direct"}, nil
		},
		deleteConversationFn: func(convID int) error {
			if convID != 5 {
				t.Errorf("Expected convID=5, got %d", convID)
			}
			return nil
		},
	}

	req, authMock := authedRequest(http.MethodDelete, "/api/conversations/delete?id=5", "")
	h := NewChatHandler(chatRepo, &mockUserRepo{}, authMock, &mockHub{})
	w := httptest.NewRecorder()

	h.DeleteConversation(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestDeleteConversation_GroupChat(t *testing.T) {
	var removedConvID, removedUserID int
	chatRepo := &mockChatRepo{
		getConversationByIDFn: func(convID, userID int) (*models.Conversation, error) {
			return &models.Conversation{ID: convID, Type: "group"}, nil
		},
		removeMemberFromGroupFn: func(convID, userID int) error {
			removedConvID = convID
			removedUserID = userID
			return nil
		},
	}

	req, authMock := authedRequest(http.MethodDelete, "/api/conversations/delete?id=5", "")
	h := NewChatHandler(chatRepo, &mockUserRepo{}, authMock, &mockHub{})
	w := httptest.NewRecorder()

	h.DeleteConversation(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if removedConvID != 5 || removedUserID != 1 {
		t.Errorf("Expected remove convID=5 userID=1, got convID=%d userID=%d", removedConvID, removedUserID)
	}
}

// =========== GetOnlineUsers Tests ===========

func TestGetOnlineUsers_Success(t *testing.T) {
	hub := &mockHub{onlineUsers: []int{1, 2, 3}}
	req, authMock := authedRequest(http.MethodGet, "/api/users/online", "")
	h := NewChatHandler(&mockChatRepo{}, &mockUserRepo{}, authMock, hub)
	w := httptest.NewRecorder()

	h.GetOnlineUsers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	resp := decodeChatResponse(t, w)
	users := resp["online_users"].([]interface{})
	if len(users) != 3 {
		t.Errorf("Expected 3 online users, got %d", len(users))
	}
}

// =========== ReactToMessage Tests ===========

func TestReactToMessage_ToggleAdd(t *testing.T) {
	chatRepo := &mockChatRepo{
		getMessageConvIDFn: func(messageID, userID int) (int, error) {
			return 0, errors.New("not sender")
		},
		getConvIDByMsgIDFn: func(messageID int) (int, error) {
			return 5, nil
		},
		toggleReactionFn: func(messageID, userID int, emoji string, isCustom bool, customURL string) (bool, error) {
			if emoji != "👍" {
				t.Errorf("Expected emoji=👍, got %q", emoji)
			}
			return true, nil
		},
	}

	hub := &mockHub{}
	req, authMock := authedRequest(http.MethodPost, "/api/messages/react?msg_id=10", `{"emoji":"👍","toggle":true}`)
	h := NewChatHandler(chatRepo, &mockUserRepo{}, authMock, hub)
	w := httptest.NewRecorder()

	h.ReactToMessage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	resp := decodeChatResponse(t, w)
	if resp["success"] != true {
		t.Error("Expected success=true")
	}
	if resp["added"] != true {
		t.Error("Expected added=true")
	}
}

// =========== SendTyping Tests ===========

func TestSendTyping_Success(t *testing.T) {
	hub := &mockHub{}
	req, authMock := authedRequest(http.MethodPost, "/api/conversations/typing?id=5", "")
	h := NewChatHandler(&mockChatRepo{}, &mockUserRepo{}, authMock, hub)
	w := httptest.NewRecorder()

	h.SendTyping(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if len(hub.broadcastedEvents) != 1 {
		t.Errorf("Expected 1 broadcast, got %d", len(hub.broadcastedEvents))
	}
	if hub.broadcastedEvents[0].Type != websocket.EventTypeTyping {
		t.Errorf("Expected event type=%q, got %q", websocket.EventTypeTyping, hub.broadcastedEvents[0].Type)
	}
}
