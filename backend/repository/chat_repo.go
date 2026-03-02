package repository

import (
	"context"
	"time"

	"netlink/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ChatRepository struct {
	pool *pgxpool.Pool
}

// GetPool returns the database pool (for internal use by handlers)
func (r *ChatRepository) GetPool() *pgxpool.Pool {
	return r.pool
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

	-- Voice message columns
	ALTER TABLE messages ADD COLUMN IF NOT EXISTS voice_file_path TEXT;
	ALTER TABLE messages ADD COLUMN IF NOT EXISTS voice_duration DOUBLE PRECISION;
	ALTER TABLE messages ADD COLUMN IF NOT EXISTS voice_waveform DOUBLE PRECISION[];
	ALTER TABLE messages ADD COLUMN IF NOT EXISTS voice_file_size BIGINT;
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
	ctx := context.Background()

	// Single query with lateral joins — eliminates N+1 for last message and unread count
	rows, err := r.pool.Query(ctx,
		`SELECT
			c.id, c.type, COALESCE(c.name, ''), COALESCE(c.avatar, ''), c.created_by, c.created_at, c.updated_at,
			lm.id, lm.conversation_id, lm.sender_id, lm.type, lm.content, lm.created_at,
			uc.count
		 FROM conversations c
		 JOIN conversation_members cm ON c.id = cm.conversation_id AND cm.user_id = $1
		 LEFT JOIN LATERAL (
			SELECT m.id, m.conversation_id, m.sender_id, m.type, m.content, m.created_at
			FROM messages m
			WHERE m.conversation_id = c.id AND m.deleted_at IS NULL
			ORDER BY m.created_at DESC LIMIT 1
		 ) lm ON true
		 LEFT JOIN LATERAL (
			SELECT COUNT(*) as count
			FROM messages m2
			WHERE m2.conversation_id = c.id
			  AND m2.sender_id != $1
			  AND m2.created_at > cm.last_read_at
			  AND m2.deleted_at IS NULL
		 ) uc ON true
		 ORDER BY c.updated_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []models.Conversation
	var convIDs []int
	for rows.Next() {
		var conv models.Conversation
		var createdBy *int
		var lmID, lmConvID, lmSenderID *int
		var lmType, lmContent *string
		var lmCreatedAt *time.Time
		var unreadCount int

		if err := rows.Scan(
			&conv.ID, &conv.Type, &conv.Name, &conv.Avatar, &createdBy, &conv.CreatedAt, &conv.UpdatedAt,
			&lmID, &lmConvID, &lmSenderID, &lmType, &lmContent, &lmCreatedAt,
			&unreadCount,
		); err != nil {
			return nil, err
		}

		if createdBy != nil {
			conv.CreatedBy = *createdBy
		}

		if lmID != nil {
			conv.LastMessage = &models.Message{
				ID:             *lmID,
				ConversationID: *lmConvID,
				SenderID:       *lmSenderID,
				Type:           models.MessageType(*lmType),
				Content:        *lmContent,
				CreatedAt:      *lmCreatedAt,
				Reactions:      []models.ReactionSummary{},
			}
		}

		conv.UnreadCount = unreadCount
		conversations = append(conversations, conv)
		convIDs = append(convIDs, conv.ID)
	}

	// Batch-fetch members for all conversations in one query
	if len(convIDs) > 0 {
		memberRows, err := r.pool.Query(ctx,
			`SELECT cm.conversation_id, cm.user_id, cm.role, cm.joined_at, cm.last_read_at,
			        u.id, u.email, COALESCE(u.name, ''), COALESCE(u.avatar, ''), u.last_seen_at
			 FROM conversation_members cm
			 JOIN users u ON cm.user_id = u.id
			 WHERE cm.conversation_id = ANY($1)`,
			convIDs,
		)
		if err == nil {
			defer memberRows.Close()
			membersMap := make(map[int][]models.ConversationMember)
			for memberRows.Next() {
				var m models.ConversationMember
				var u models.User
				if err := memberRows.Scan(&m.ConversationID, &m.UserID, &m.Role, &m.JoinedAt, &m.LastReadAt,
					&u.ID, &u.Email, &u.Name, &u.Avatar, &u.LastSeenAt); err != nil {
					continue
				}
				m.User = &u
				membersMap[m.ConversationID] = append(membersMap[m.ConversationID], m)
			}
			for i := range conversations {
				conversations[i].Members = membersMap[conversations[i].ID]
			}
		}
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

type CreateMessageParams struct {
	VoiceFilePath string
	VoiceDuration float64
	VoiceWaveform []float64
	VoiceFileSize int64
}

func (r *ChatRepository) CreateMessage(convID, senderID int, msgType models.MessageType, content string, replyToID *int, voice ...CreateMessageParams) (*models.Message, error) {
	ctx := context.Background()

	var msg models.Message

	if len(voice) > 0 && voice[0].VoiceFilePath != "" {
		v := voice[0]
		err := r.pool.QueryRow(ctx,
			`INSERT INTO messages (conversation_id, sender_id, type, content, reply_to_id, voice_file_path, voice_duration, voice_waveform, voice_file_size)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			 RETURNING id, conversation_id, sender_id, type, content, reply_to_id, created_at, updated_at,
			           COALESCE(voice_file_path, ''), COALESCE(voice_duration, 0), COALESCE(voice_waveform, ARRAY[]::double precision[]), COALESCE(voice_file_size, 0)`,
			convID, senderID, msgType, content, replyToID, v.VoiceFilePath, v.VoiceDuration, v.VoiceWaveform, v.VoiceFileSize,
		).Scan(&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.Type, &msg.Content, &msg.ReplyToID, &msg.CreatedAt, &msg.UpdatedAt,
			&msg.VoiceFilePath, &msg.VoiceDuration, &msg.VoiceWaveform, &msg.VoiceFileSize)
		if err != nil {
			return nil, err
		}
	} else {
		err := r.pool.QueryRow(ctx,
			`INSERT INTO messages (conversation_id, sender_id, type, content, reply_to_id)
			 VALUES ($1, $2, $3, $4, $5)
			 RETURNING id, conversation_id, sender_id, type, content, reply_to_id, created_at, updated_at`,
			convID, senderID, msgType, content, replyToID,
		).Scan(&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.Type, &msg.Content, &msg.ReplyToID, &msg.CreatedAt, &msg.UpdatedAt)
		if err != nil {
			return nil, err
		}
	}

	// Update conversation updated_at
	r.pool.Exec(ctx, `UPDATE conversations SET updated_at = CURRENT_TIMESTAMP WHERE id = $1`, convID)

	return &msg, nil
}

func (r *ChatRepository) GetMessages(convID int, limit, offset int) ([]models.Message, error) {
	rows, err := r.pool.Query(
		context.Background(),
		`SELECT m.id, m.conversation_id, m.sender_id, m.type, m.content, m.reply_to_id, m.created_at, m.updated_at,
		        COALESCE(m.voice_file_path, ''), COALESCE(m.voice_duration, 0),
		        COALESCE(m.voice_waveform, ARRAY[]::double precision[]), COALESCE(m.voice_file_size, 0),
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
			&m.VoiceFilePath, &m.VoiceDuration, &m.VoiceWaveform, &m.VoiceFileSize,
			&u.ID, &u.Email, &u.Name, &u.Avatar); err != nil {
			return nil, err
		}
		m.Sender = &u
		messages = append(messages, m)
	}
	
	// Populate reactions
	if len(messages) > 0 {
		msgIDs := make([]int, len(messages))
		for i, m := range messages {
			msgIDs[i] = m.ID
		}
		
		reactionsMap, err := r.GetBatchMessageReactions(msgIDs)
		if err == nil {
			for i := range messages {
				if reactions, ok := reactionsMap[messages[i].ID]; ok {
					messages[i].Reactions = reactions
				} else {
					messages[i].Reactions = []models.ReactionSummary{}
				}
			}
		}
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
	
	// Add reactions
	if summaries, err := r.GetMessageReactionsSummary(msg.ID); err == nil {
		msg.Reactions = summaries
	} else {
		msg.Reactions = []models.ReactionSummary{}
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
	
	// Add reactions
	if summaries, err := r.GetMessageReactionsSummary(msg.ID); err == nil {
		msg.Reactions = summaries
	} else {
		msg.Reactions = []models.ReactionSummary{}
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

// Reaction methods - Optimized for 20-30ms latency

// AddReaction adds a reaction to a message (supports both system emojis and custom reactions)
func (r *ChatRepository) AddReaction(messageID, userID int, emoji string) error {
	ctx := context.Background()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO message_reactions (message_id, user_id, emoji, is_custom, custom_url) 
		 VALUES ($1, $2, $3, FALSE, NULL)
		 ON CONFLICT (message_id, user_id, emoji) DO NOTHING`,
		messageID, userID, emoji,
	)
	return err
}

// AddCustomReaction adds a custom reaction to a message
func (r *ChatRepository) AddCustomReaction(messageID, userID int, emoji, customURL string) error {
	ctx := context.Background()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO message_reactions (message_id, user_id, emoji, is_custom, custom_url) 
		 VALUES ($1, $2, $3, TRUE, $4)
		 ON CONFLICT (message_id, user_id, emoji) DO NOTHING`,
		messageID, userID, emoji, customURL,
	)
	return err
}

// RemoveReaction removes a reaction from a message
func (r *ChatRepository) RemoveReaction(messageID, userID int, emoji string) error {
	ctx := context.Background()
	_, err := r.pool.Exec(ctx,
		`DELETE FROM message_reactions 
		 WHERE message_id = $1 AND user_id = $2 AND emoji = $3`,
		messageID, userID, emoji,
	)
	return err
}

// GetMessageReactions returns reactions in optimized format (legacy map format)
func (r *ChatRepository) GetMessageReactions(messageID int) (map[string][]int, error) {
	ctx := context.Background()
	rows, err := r.pool.Query(ctx,
		`SELECT emoji, user_id FROM message_reactions WHERE message_id = $1`,
		messageID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reactions := make(map[string][]int)
	for rows.Next() {
		var emoji string
		var userID int
		if err := rows.Scan(&emoji, &userID); err != nil {
			return nil, err
		}
		reactions[emoji] = append(reactions[emoji], userID)
	}
	return reactions, nil
}

// GetMessageReactionsSummary returns aggregated reaction data for efficient transmission
// This is optimized for minimal payload size and fast queries
func (r *ChatRepository) GetMessageReactionsSummary(messageID int) ([]models.ReactionSummary, error) {
	ctx := context.Background()

	// Use a single optimized query with aggregation
	rows, err := r.pool.Query(ctx,
		`SELECT 
			emoji, 
			is_custom, 
			COALESCE(custom_url, '') as custom_url,
			COUNT(*) as count,
			ARRAY_AGG(user_id ORDER BY created_at) as user_ids
		FROM message_reactions 
		WHERE message_id = $1
		GROUP BY emoji, is_custom, custom_url
		ORDER BY MIN(created_at)`,
		messageID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []models.ReactionSummary
	for rows.Next() {
		var summary models.ReactionSummary
		if err := rows.Scan(&summary.Emoji, &summary.IsCustom, &summary.CustomURL, &summary.Count, &summary.UserIDs); err != nil {
			return nil, err
		}
		summaries = append(summaries, summary)
	}
	return summaries, nil
}

// GetBatchMessageReactions gets reactions for multiple messages in one query
// This is extremely efficient for loading conversation history
func (r *ChatRepository) GetBatchMessageReactions(messageIDs []int) (map[int][]models.ReactionSummary, error) {
	if len(messageIDs) == 0 {
		return make(map[int][]models.ReactionSummary), nil
	}

	ctx := context.Background()
	rows, err := r.pool.Query(ctx,
		`SELECT 
			message_id,
			emoji, 
			is_custom, 
			COALESCE(custom_url, '') as custom_url,
			COUNT(*) as count,
			ARRAY_AGG(user_id ORDER BY created_at) as user_ids
		FROM message_reactions 
		WHERE message_id = ANY($1)
		GROUP BY message_id, emoji, is_custom, custom_url
		ORDER BY message_id, MIN(created_at)`,
		messageIDs,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int][]models.ReactionSummary)
	for rows.Next() {
		var msgID int
		var summary models.ReactionSummary
		if err := rows.Scan(&msgID, &summary.Emoji, &summary.IsCustom, &summary.CustomURL, &summary.Count, &summary.UserIDs); err != nil {
			return nil, err
		}
		result[msgID] = append(result[msgID], summary)
	}
	return result, nil
}

// ToggleReaction adds or removes a reaction in a single operation
// Returns true if added, false if removed
func (r *ChatRepository) ToggleReaction(messageID, userID int, emoji string, isCustom bool, customURL string) (bool, error) {
	ctx := context.Background()

	// Check if reaction exists
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM message_reactions WHERE message_id = $1 AND user_id = $2 AND emoji = $3)`,
		messageID, userID, emoji,
	).Scan(&exists)
	if err != nil {
		return false, err
	}

	if exists {
		// Remove reaction
		_, err = r.pool.Exec(ctx,
			`DELETE FROM message_reactions WHERE message_id = $1 AND user_id = $2 AND emoji = $3`,
			messageID, userID, emoji,
		)
		return false, err
	} else {
		// Add reaction
		_, err = r.pool.Exec(ctx,
			`INSERT INTO message_reactions (message_id, user_id, emoji, is_custom, custom_url) 
			 VALUES ($1, $2, $3, $4, $5)`,
			messageID, userID, emoji, isCustom, customURL,
		)
		return true, err
	}
}

// Forward message
func (r *ChatRepository) ForwardMessage(originalMsgID, targetConvID, senderID int) (*models.Message, error) {
	ctx := context.Background()

	// Get original message
	var content string
	var msgType models.MessageType
	err := r.pool.QueryRow(ctx,
		`SELECT content, type FROM messages WHERE id = $1`,
		originalMsgID,
	).Scan(&content, &msgType)
	if err != nil {
		return nil, err
	}

	// Create forwarded message
	var msg models.Message
	err = r.pool.QueryRow(ctx,
		`INSERT INTO messages (conversation_id, sender_id, type, content, forwarded_from) 
		 VALUES ($1, $2, $3, $4, $5) 
		 RETURNING id, conversation_id, sender_id, type, content, forwarded_from, created_at, updated_at`,
		targetConvID, senderID, msgType, content, originalMsgID,
	).Scan(&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.Type, &msg.Content, &msg.ForwardedFrom, &msg.CreatedAt, &msg.UpdatedAt)

	return &msg, err
}

// Delete message (soft delete)
func (r *ChatRepository) DeleteMessage(messageID, userID int) error {
	_, err := r.pool.Exec(
		context.Background(),
		`UPDATE messages SET deleted_at = NOW(), content = 'Message deleted' 
		 WHERE id = $1 AND sender_id = $2 AND deleted_at IS NULL`,
		messageID, userID,
	)
	return err
}

// GetMessageConversationID gets the conversation ID for a message owned by the user
func (r *ChatRepository) GetMessageConversationID(messageID, userID int) (int, error) {
	var convID int
	err := r.pool.QueryRow(
		context.Background(),
		`SELECT conversation_id FROM messages WHERE id = $1 AND sender_id = $2`,
		messageID, userID,
	).Scan(&convID)
	return convID, err
}

// GetConversationIDByMessageID gets the conversation ID for any message (no sender check)
func (r *ChatRepository) GetConversationIDByMessageID(messageID int) (int, error) {
	var convID int
	err := r.pool.QueryRow(
		context.Background(),
		`SELECT conversation_id FROM messages WHERE id = $1`,
		messageID,
	).Scan(&convID)
	return convID, err
}

// Initialize reactions schema
func (r *ChatRepository) InitReactionsSchema(ctx context.Context) error {
	schema := `
	-- Add forwarded_from column to messages
	DO $$ 
	BEGIN
		IF NOT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name='messages' AND column_name='forwarded_from'
		) THEN
			ALTER TABLE messages ADD COLUMN forwarded_from INTEGER REFERENCES messages(id) ON DELETE SET NULL;
		END IF;
	END $$;

	-- Create message_reactions table with custom reaction support
	CREATE TABLE IF NOT EXISTS message_reactions (
		id SERIAL PRIMARY KEY,
		message_id INTEGER NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		emoji VARCHAR(100) NOT NULL,
		is_custom BOOLEAN DEFAULT FALSE,
		custom_url TEXT,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		UNIQUE(message_id, user_id, emoji)
	);

	-- Create indexes for optimal performance (20-30ms target)
	CREATE INDEX IF NOT EXISTS idx_reactions_message ON message_reactions(message_id);
	CREATE INDEX IF NOT EXISTS idx_reactions_user ON message_reactions(user_id);
	CREATE INDEX IF NOT EXISTS idx_reactions_message_emoji ON message_reactions(message_id, emoji);
	CREATE INDEX IF NOT EXISTS idx_messages_forwarded ON messages(forwarded_from);
	
	-- Add is_custom and custom_url columns if they don't exist (for existing installations)
	DO $$ 
	BEGIN
		IF NOT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name='message_reactions' AND column_name='is_custom'
		) THEN
			ALTER TABLE message_reactions ADD COLUMN is_custom BOOLEAN DEFAULT FALSE;
		END IF;
		
		IF NOT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name='message_reactions' AND column_name='custom_url'
		) THEN
			ALTER TABLE message_reactions ADD COLUMN custom_url TEXT;
		END IF;
		
		-- Increase emoji column size if needed
		ALTER TABLE message_reactions ALTER COLUMN emoji TYPE VARCHAR(100);
	END $$;
	`
	_, err := r.pool.Exec(ctx, schema)
	return err
}
