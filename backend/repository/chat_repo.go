package repository

import (
	"context"
	"time"

	"go-to-do/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ChatRepository struct {
	pool *pgxpool.Pool
}

func NewChatRepository(pool *pgxpool.Pool) *ChatRepository {
	return &ChatRepository{pool: pool}
}

func (r *ChatRepository) InitSchema(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS conversations (
		id SERIAL PRIMARY KEY,
		type VARCHAR(20) NOT NULL DEFAULT 'direct',
		name VARCHAR(255),
		avatar TEXT,
		created_by INTEGER REFERENCES users(id),
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS conversation_members (
		conversation_id INTEGER NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		role VARCHAR(20) DEFAULT 'member',
		joined_at TIMESTAMPTZ DEFAULT NOW(),
		last_read_at TIMESTAMPTZ DEFAULT NOW(),
		PRIMARY KEY (conversation_id, user_id)
	);

	CREATE TABLE IF NOT EXISTS messages (
		id SERIAL PRIMARY KEY,
		conversation_id INTEGER NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
		sender_id INTEGER NOT NULL REFERENCES users(id),
		type VARCHAR(20) DEFAULT 'text',
		content TEXT NOT NULL,
		reply_to_id INTEGER REFERENCES messages(id),
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW(),
		deleted_at TIMESTAMPTZ
	);

	CREATE INDEX IF NOT EXISTS idx_conv_members_user ON conversation_members(user_id);
	CREATE INDEX IF NOT EXISTS idx_messages_conv ON messages(conversation_id, created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_messages_sender ON messages(sender_id);
	
	-- Add phone and name to users if not exists
	ALTER TABLE users ADD COLUMN IF NOT EXISTS phone VARCHAR(20);
	ALTER TABLE users ADD COLUMN IF NOT EXISTS name VARCHAR(100) DEFAULT '';
	ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar TEXT;
	ALTER TABLE users ADD COLUMN IF NOT EXISTS last_seen_at TIMESTAMPTZ;
	`
	_, err := r.pool.Exec(ctx, schema)
	return err
}

// Conversation methods

func (r *ChatRepository) CreateDirectConversation(userID1, userID2 int) (*models.Conversation, error) {
	ctx := context.Background()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Check if direct conversation already exists
	var existingID int
	err = tx.QueryRow(ctx, `
		SELECT c.id FROM conversations c
		JOIN conversation_members cm1 ON c.id = cm1.conversation_id AND cm1.user_id = $1
		JOIN conversation_members cm2 ON c.id = cm2.conversation_id AND cm2.user_id = $2
		WHERE c.type = 'direct'
	`, userID1, userID2).Scan(&existingID)

	if err == nil {
		tx.Commit(ctx)
		return r.GetConversationByID(existingID, userID1)
	}

	// Create new conversation
	var convID int
	err = tx.QueryRow(ctx,
		`INSERT INTO conversations (type, created_by) VALUES ('direct', $1) RETURNING id`,
		userID1,
	).Scan(&convID)
	if err != nil {
		return nil, err
	}

	// Add members
	_, err = tx.Exec(ctx,
		`INSERT INTO conversation_members (conversation_id, user_id, role) VALUES ($1, $2, 'member'), ($1, $3, 'member')`,
		convID, userID1, userID2,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.GetConversationByID(convID, userID1)
}

func (r *ChatRepository) CreateGroupConversation(name string, creatorID int, memberIDs []int) (*models.Conversation, error) {
	ctx := context.Background()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var convID int
	err = tx.QueryRow(ctx,
		`INSERT INTO conversations (type, name, created_by) VALUES ('group', $1, $2) RETURNING id`,
		name, creatorID,
	).Scan(&convID)
	if err != nil {
		return nil, err
	}

	// Add creator as admin
	_, err = tx.Exec(ctx,
		`INSERT INTO conversation_members (conversation_id, user_id, role) VALUES ($1, $2, 'admin')`,
		convID, creatorID,
	)
	if err != nil {
		return nil, err
	}

	// Add other members
	for _, memberID := range memberIDs {
		if memberID != creatorID {
			_, err = tx.Exec(ctx,
				`INSERT INTO conversation_members (conversation_id, user_id, role) VALUES ($1, $2, 'member')`,
				convID, memberID,
			)
			if err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.GetConversationByID(convID, creatorID)
}

func (r *ChatRepository) GetConversationByID(convID, userID int) (*models.Conversation, error) {
	var conv models.Conversation
	var createdBy *int
	err := r.pool.QueryRow(
		context.Background(),
		`SELECT c.id, c.type, COALESCE(c.name, ''), COALESCE(c.avatar, ''), c.created_by, c.created_at, c.updated_at
		 FROM conversations c
		 JOIN conversation_members cm ON c.id = cm.conversation_id
		 WHERE c.id = $1 AND cm.user_id = $2`,
		convID, userID,
	).Scan(&conv.ID, &conv.Type, &conv.Name, &conv.Avatar, &createdBy, &conv.CreatedAt, &conv.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if createdBy != nil {
		conv.CreatedBy = *createdBy
	}

	conv.Members, _ = r.GetConversationMembers(convID)
	return &conv, nil
}

func (r *ChatRepository) GetUserConversations(userID int) ([]models.Conversation, error) {
	rows, err := r.pool.Query(
		context.Background(),
		`SELECT c.id, c.type, COALESCE(c.name, ''), COALESCE(c.avatar, ''), c.created_by, c.created_at, c.updated_at
		 FROM conversations c
		 JOIN conversation_members cm ON c.id = cm.conversation_id
		 WHERE cm.user_id = $1
		 ORDER BY c.updated_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []models.Conversation
	for rows.Next() {
		var conv models.Conversation
		var createdBy *int
		if err := rows.Scan(&conv.ID, &conv.Type, &conv.Name, &conv.Avatar, &createdBy, &conv.CreatedAt, &conv.UpdatedAt); err != nil {
			return nil, err
		}
		if createdBy != nil {
			conv.CreatedBy = *createdBy
		}
		conv.Members, _ = r.GetConversationMembers(conv.ID)
		conv.LastMessage, _ = r.GetLastMessage(conv.ID)
		conv.UnreadCount, _ = r.GetUnreadCount(conv.ID, userID)
		conversations = append(conversations, conv)
	}
	return conversations, nil
}

func (r *ChatRepository) GetConversationMembers(convID int) ([]models.ConversationMember, error) {
	rows, err := r.pool.Query(
		context.Background(),
		`SELECT cm.conversation_id, cm.user_id, cm.role, cm.joined_at, cm.last_read_at,
		        u.id, u.email, COALESCE(u.name, ''), COALESCE(u.avatar, ''), u.last_seen_at
		 FROM conversation_members cm
		 JOIN users u ON cm.user_id = u.id
		 WHERE cm.conversation_id = $1`,
		convID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []models.ConversationMember
	for rows.Next() {
		var m models.ConversationMember
		var u models.User
		if err := rows.Scan(&m.ConversationID, &m.UserID, &m.Role, &m.JoinedAt, &m.LastReadAt,
			&u.ID, &u.Email, &u.Name, &u.Avatar, &u.LastSeenAt); err != nil {
			return nil, err
		}
		m.User = &u
		members = append(members, m)
	}
	return members, nil
}

func (r *ChatRepository) GetConversationMemberIDs(convID int) ([]int, error) {
	rows, err := r.pool.Query(
		context.Background(),
		`SELECT user_id FROM conversation_members WHERE conversation_id = $1`,
		convID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// Message methods

func (r *ChatRepository) CreateMessage(convID, senderID int, msgType models.MessageType, content string, replyToID *int) (*models.Message, error) {
	ctx := context.Background()

	var msg models.Message
	err := r.pool.QueryRow(ctx,
		`INSERT INTO messages (conversation_id, sender_id, type, content, reply_to_id) 
		 VALUES ($1, $2, $3, $4, $5) 
		 RETURNING id, conversation_id, sender_id, type, content, reply_to_id, created_at, updated_at`,
		convID, senderID, msgType, content, replyToID,
	).Scan(&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.Type, &msg.Content, &msg.ReplyToID, &msg.CreatedAt, &msg.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Update conversation updated_at
	r.pool.Exec(ctx, `UPDATE conversations SET updated_at = CURRENT_TIMESTAMP WHERE id = $1`, convID)

	return &msg, nil
}

func (r *ChatRepository) GetMessages(convID int, limit, offset int) ([]models.Message, error) {
	rows, err := r.pool.Query(
		context.Background(),
		`SELECT m.id, m.conversation_id, m.sender_id, m.type, m.content, m.reply_to_id, m.created_at, m.updated_at,
		        u.id, u.email, COALESCE(u.name, ''), COALESCE(u.avatar, '')
		 FROM messages m
		 JOIN users u ON m.sender_id = u.id
		 WHERE m.conversation_id = $1 AND m.deleted_at IS NULL
		 ORDER BY m.created_at DESC
		 LIMIT $2 OFFSET $3`,
		convID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var m models.Message
		var u models.User
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.SenderID, &m.Type, &m.Content, &m.ReplyToID, &m.CreatedAt, &m.UpdatedAt,
			&u.ID, &u.Email, &u.Name, &u.Avatar); err != nil {
			return nil, err
		}
		m.Sender = &u
		messages = append(messages, m)
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

func (r *ChatRepository) GetLastMessage(convID int) (*models.Message, error) {
	var msg models.Message
	err := r.pool.QueryRow(
		context.Background(),
		`SELECT id, conversation_id, sender_id, type, content, created_at 
		 FROM messages WHERE conversation_id = $1 AND deleted_at IS NULL
		 ORDER BY created_at DESC LIMIT 1`,
		convID,
	).Scan(&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.Type, &msg.Content, &msg.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *ChatRepository) GetUnreadCount(convID, userID int) (int, error) {
	var count int
	err := r.pool.QueryRow(
		context.Background(),
		`SELECT COUNT(*) FROM messages m
		 JOIN conversation_members cm ON m.conversation_id = cm.conversation_id AND cm.user_id = $2
		 WHERE m.conversation_id = $1 AND m.sender_id != $2 AND m.created_at > cm.last_read_at`,
		convID, userID,
	).Scan(&count)
	return count, err
}

func (r *ChatRepository) MarkAsRead(convID, userID int) error {
	_, err := r.pool.Exec(
		context.Background(),
		`UPDATE conversation_members SET last_read_at = CURRENT_TIMESTAMP 
		 WHERE conversation_id = $1 AND user_id = $2`,
		convID, userID,
	)
	return err
}

func (r *ChatRepository) DeleteMessage(msgID, userID int) error {
	_, err := r.pool.Exec(
		context.Background(),
		`UPDATE messages SET deleted_at = CURRENT_TIMESTAMP 
		 WHERE id = $1 AND sender_id = $2`,
		msgID, userID,
	)
	return err
}

func (r *ChatRepository) UpdateMessage(msgID, userID int, content string) (*models.Message, error) {
	ctx := context.Background()
	var msg models.Message
	err := r.pool.QueryRow(ctx,
		`UPDATE messages SET content = $1, updated_at = CURRENT_TIMESTAMP 
		 WHERE id = $2 AND sender_id = $3 AND deleted_at IS NULL
		 RETURNING id, conversation_id, sender_id, type, content, reply_to_id, created_at, updated_at`,
		content, msgID, userID,
	).Scan(&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.Type, &msg.Content, &msg.ReplyToID, &msg.CreatedAt, &msg.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *ChatRepository) UpdateLastSeen(userID int) error {
	_, err := r.pool.Exec(
		context.Background(),
		`UPDATE users SET last_seen_at = $1 WHERE id = $2`,
		time.Now(), userID,
	)
	return err
}

func (r *ChatRepository) AddMemberToGroup(convID, userID int) error {
	_, err := r.pool.Exec(
		context.Background(),
		`INSERT INTO conversation_members (conversation_id, user_id, role) VALUES ($1, $2, 'member')
		 ON CONFLICT DO NOTHING`,
		convID, userID,
	)
	return err
}

func (r *ChatRepository) RemoveMemberFromGroup(convID, userID int) error {
	_, err := r.pool.Exec(
		context.Background(),
		`DELETE FROM conversation_members WHERE conversation_id = $1 AND user_id = $2`,
		convID, userID,
	)
	return err
}

func (r *ChatRepository) DeleteConversation(convID int) error {
	ctx := context.Background()
	// Delete messages first (due to foreign key)
	_, err := r.pool.Exec(ctx, `DELETE FROM messages WHERE conversation_id = $1`, convID)
	if err != nil {
		return err
	}
	// Delete members
	_, err = r.pool.Exec(ctx, `DELETE FROM conversation_members WHERE conversation_id = $1`, convID)
	if err != nil {
		return err
	}
	// Delete conversation
	_, err = r.pool.Exec(ctx, `DELETE FROM conversations WHERE id = $1`, convID)
	return err
}
