package handlers

import (
	"net/http"

	"netlink/models"
	"netlink/repository"
	"netlink/websocket"
)

// AuthServiceInterface defines what auth handlers need from the auth service.
type AuthServiceInterface interface {
	Login(email, password string) (*models.Session, error)
	Register(email, password string, name, phone *string) error
	VerifyEmail(token string) error
	ForgotPassword(email string) error
	ResetPassword(token, newPassword string) error
	GetUserFromRequest(r *http.Request) (*models.User, error)
	GetUserFromSession(session *models.Session) (*models.User, error)
	SetSessionCookie(w http.ResponseWriter, session *models.Session)
	ClearSessionCookie(w http.ResponseWriter)
	ValidateEmail(email string) bool
	ValidatePassword(password string) bool
	TestEmailDelivery(to string) error
	Logout(sessionID string) error
}

// ChatRepoInterface defines what chat handlers need from the chat repository.
type ChatRepoInterface interface {
	GetUserConversations(userID int) ([]models.Conversation, error)
	CreateDirectConversation(userID1, userID2 int) (*models.Conversation, error)
	CreateGroupConversation(name string, creatorID int, memberIDs []int) (*models.Conversation, error)
	GetConversationByID(convID, userID int) (*models.Conversation, error)
	GetConversationMemberIDs(convID int) ([]int, error)
	GetMessages(convID int, limit, offset int) ([]models.Message, error)
	CreateMessage(convID, senderID int, msgType models.MessageType, content string, replyToID *int, voice ...repository.CreateMessageParams) (*models.Message, error)
	UpdateMessage(msgID, userID int, content string) (*models.Message, error)
	MarkAsRead(convID, userID int) error
	UpdateLastSeen(userID int) error
	DeleteConversation(convID int) error
	RemoveMemberFromGroup(convID, userID int) error
	AddReaction(messageID, userID int, emoji string) error
	AddCustomReaction(messageID, userID int, emoji, customURL string) error
	RemoveReaction(messageID, userID int, emoji string) error
	GetMessageReactionsSummary(messageID int) ([]models.ReactionSummary, error)
	ToggleReaction(messageID, userID int, emoji string, isCustom bool, customURL string) (bool, error)
	ForwardMessage(originalMsgID, targetConvID, senderID int) (*models.Message, error)
	DeleteMessage(messageID, userID int) error
	GetMessageConversationID(messageID, userID int) (int, error)
	GetConversationIDByMessageID(messageID int) (int, error)
}

// UserRepoInterface defines what chat handlers need from the user repository.
type UserRepoInterface interface {
	GetByID(id int) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetByPhone(phone string) (*models.User, error)
}

// HubInterface defines what chat handlers need from the websocket hub.
type HubInterface interface {
	BroadcastToUsers(userIDs []int, event *websocket.Event)
	GetOnlineUsers() []int
	Register(client *websocket.Client)
}
