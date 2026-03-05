package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"backend/pkg/models"
)

var (
	ErrInvalidChatMessage   = errors.New("invalid chat message")
	ErrChatNotFound         = errors.New("chat not found")
	ErrChatForbidden        = errors.New("chat access forbidden")
	ErrDirectChatNotAllowed = errors.New("direct chat not allowed")
	ErrGroupChatNotFound    = errors.New("group chat not found")
	ErrChatAlreadyExists    = errors.New("chat already exists")
)

type ChatRepository interface {
	// --- Checks & Permissions ---
	CanUsersChat(ctx context.Context, u1, u2 int) (bool, error)
	IsGroupMemberActive(ctx context.Context, groupID, userID int) (bool, error)
	UserHasChatAccess(ctx context.Context, userID, chatID int) (bool, error)

	// --- Chat Management ---
	GetDirectChatID(ctx context.Context, lowID, highID int) (int, error)
	CreateDirectChat(ctx context.Context, lowID, highID, creatorID int) (int, error)
	GetChatIDByGroupID(ctx context.Context, groupID int) (int, error)
	EnsureParticipant(ctx context.Context, chatID, userID int) error

	// --- Messaging (The Core) ---
	// Transaction: Insert Message + Update Chat LastMessage
	SaveMessage(ctx context.Context, chatID, senderID int, in models.SendMessageInput) (*models.ChatMessage, error)

	// --- Retrieval ---
	FetchMessages(ctx context.Context, chatID, beforeID, limit int) ([]*models.ChatMessage, error)
	FetchChatSummaries(ctx context.Context, userID, limit, offset int) ([]*models.ChatSummary, error)

	// --- Read Status ---
	GetLatestMessageID(ctx context.Context, chatID int) (int, error)
	GetMessageChatID(ctx context.Context, messageID int) (int, error)
	UpdateLastReadMessage(ctx context.Context, userID, chatID, messageID int) error

	GetGroupParticipantIDs(ctx context.Context, groupID int) ([]int, error)
}

type sqliteChatRepo struct {
	db *sql.DB
}

func NewChatRepository(db *sql.DB) ChatRepository {
	return &sqliteChatRepo{db: db}
}

func (r *sqliteChatRepo) CanUsersChat(ctx context.Context, u1, u2 int) (bool, error) {
	var blocked int
	r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM followers
		WHERE ((follower_id = ? AND followed_id = ?) OR (follower_id = ? AND followed_id = ?))
		  AND status = 'blocked'
	`, u1, u2, u2, u1).Scan(&blocked)
	if blocked > 0 {
		return false, nil
	}

	var accepted int
	r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM followers
		WHERE ((follower_id = ? AND followed_id = ?) OR (follower_id = ? AND followed_id = ?))
		  AND status = 'accepted'
	`, u1, u2, u2, u1).Scan(&accepted)
	return accepted > 0, nil
}

func (r *sqliteChatRepo) SaveMessage(ctx context.Context, chatID, senderID int, in models.SendMessageInput) (*models.ChatMessage, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 1. Insert Message
	res, err := tx.ExecContext(ctx, `
		INSERT INTO chat_messages (chat_id, sender_id, message_type, body, media_url)
		VALUES (?, ?, ?, ?, ?)
	`, chatID, senderID, in.MessageType, in.Body, in.MediaURL)
	if err != nil {
		log.Printf("error inserting message: %v", err)
		return nil, err
	}
	msgID, _ := res.LastInsertId()

	// 2. Update Chat & Participants
	tx.ExecContext(ctx, `
		UPDATE chats
		SET last_message_id = ?, last_message_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, msgID, chatID)
	tx.ExecContext(ctx, `
		UPDATE chat_participants
		SET last_read_message_id = ?
		WHERE chat_id = ? AND user_id = ?
	`, msgID, chatID, senderID)

	if err := tx.Commit(); err != nil {
		log.Printf("error committing transaction: %v", err)
		return nil, err
	}

	out := new(models.ChatMessage)
	var body sql.NullString
	var media sql.NullString
	if err := r.db.QueryRowContext(ctx, `
		SELECT id, chat_id, sender_id, message_type, body, media_url, COALESCE(datetime(created_at), '')
		FROM chat_messages
		WHERE id = ?
	`, msgID).Scan(&out.ID, &out.ChatID, &out.SenderID, &out.MessageType, &body, &media, &out.CreatedAt); err != nil {
		log.Printf("error fetching saved message: %v", err)
		return nil, err
	}
	if body.Valid {
		out.Body = body.String
	}
	if media.Valid {
		out.MediaURL = media.String
	}

	return out, nil
}

func (r *sqliteChatRepo) GetDirectChatID(ctx context.Context, lowID, highID int) (int, error) {
	var chatID int
	err := r.db.QueryRowContext(ctx, `
        SELECT id
        FROM chats
        WHERE type = 'direct' AND user_low_id = ? AND user_high_id = ?
        LIMIT 1
    `, lowID, highID).Scan(&chatID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrChatNotFound
		}
		return 0, err
	}
	return chatID, nil
}

func (r *sqliteChatRepo) CreateDirectChat(ctx context.Context, lowID, highID, creatorID int) (int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `
        INSERT INTO chats (type, user_low_id, user_high_id, created_by, last_message_at)
        VALUES ('direct', ?, ?, ?, CURRENT_TIMESTAMP)
    `, lowID, highID, creatorID)

	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "idx_chats_direct_pair_unique") {
			return 0, ErrChatAlreadyExists
		}
		return 0, err
	}

	chatID64, _ := res.LastInsertId()
	chatID := int(chatID64)

	_, err = tx.ExecContext(ctx, `
        INSERT OR IGNORE INTO chat_participants (chat_id, user_id)
        VALUES (?, ?), (?, ?)
    `, chatID, lowID, chatID, highID)

	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return chatID, nil
}

func (r *sqliteChatRepo) GetChatIDByGroupID(ctx context.Context, groupID int) (int, error) {
	var chatID int
	err := r.db.QueryRowContext(ctx, `SELECT id FROM chats WHERE type = 'group' AND group_id = ?`, groupID).Scan(&chatID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrGroupChatNotFound
		}
		return 0, err
	}
	return chatID, nil
}

func (r *sqliteChatRepo) IsGroupMemberActive(ctx context.Context, groupID, userID int) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
        SELECT COUNT(*) FROM group_members 
        WHERE group_id = ? AND user_id = ? AND status = 'active'
    `, groupID, userID).Scan(&count)
	return count > 0, err
}

func (r *sqliteChatRepo) EnsureParticipant(ctx context.Context, chatID, userID int) error {
	_, err := r.db.ExecContext(ctx, `
        INSERT OR IGNORE INTO chat_participants (chat_id, user_id) VALUES (?, ?)
    `, chatID, userID)
	return err
}

func (r *sqliteChatRepo) UserHasChatAccess(ctx context.Context, userID, chatID int) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
        SELECT COUNT(*) FROM chat_participants 
        WHERE chat_id = ? AND user_id = ? AND left_at IS NULL
    `, chatID, userID).Scan(&count)

	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *sqliteChatRepo) FetchMessages(ctx context.Context, chatID, beforeID, limit int) ([]*models.ChatMessage, error) {
	query := `
        SELECT id, chat_id, sender_id, message_type, body, media_url, COALESCE(datetime(created_at), '')
        FROM chat_messages
        WHERE chat_id = ?
    `
	args := []interface{}{chatID}
	if beforeID > 0 {
		query += " AND id < ?"
		args = append(args, beforeID)
	}
	query += " ORDER BY id DESC LIMIT ?"
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*models.ChatMessage
	for rows.Next() {
		m := new(models.ChatMessage)
		var body, media sql.NullString
		if err := rows.Scan(&m.ID, &m.ChatID, &m.SenderID, &m.MessageType, &body, &media, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.Body = body.String
		m.MediaURL = media.String
		messages = append(messages, m)
	}
	return messages, nil
}

func (r *sqliteChatRepo) GetLatestMessageID(ctx context.Context, chatID int) (int, error) {
	var id int
	err := r.db.QueryRowContext(ctx, `SELECT COALESCE(last_message_id, 0) FROM chats WHERE id = ?`, chatID).Scan(&id)
	return id, err
}

func (r *sqliteChatRepo) GetMessageChatID(ctx context.Context, messageID int) (int, error) {
	var chatID int
	err := r.db.QueryRowContext(ctx, `SELECT chat_id FROM chat_messages WHERE id = ?`, messageID).Scan(&chatID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrChatNotFound
	}
	return chatID, err
}

func (r *sqliteChatRepo) UpdateLastReadMessage(ctx context.Context, userID, chatID, messageID int) error {
	_, err := r.db.ExecContext(ctx, `
        UPDATE chat_participants
        SET last_read_message_id = CASE
            WHEN last_read_message_id IS NULL OR last_read_message_id < ? THEN ?
            ELSE last_read_message_id
        END
        WHERE chat_id = ? AND user_id = ?
    `, messageID, messageID, chatID, userID)
	return err
}

func (r *sqliteChatRepo) FetchChatSummaries(ctx context.Context, userID, limit, offset int) ([]*models.ChatSummary, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT
            c.id, c.type, COALESCE(c.group_id, 0), COALESCE(c.last_message_id, 0),
            COALESCE(cm.sender_id, 0), COALESCE(cm.message_type, ''),
            CASE
                WHEN cm.body IS NOT NULL AND length(trim(cm.body)) > 0 THEN cm.body
                WHEN cm.media_url IS NOT NULL AND length(trim(cm.media_url)) > 0 THEN '[image]'
                ELSE ''
            END,
            COALESCE(datetime(cm.created_at), ''),
            CASE
                WHEN c.type = 'direct' THEN CASE WHEN c.user_low_id = ? THEN c.user_high_id ELSE c.user_low_id END
                ELSE 0
            END as other_user_id,
            COALESCE(u.first_name, ''), COALESCE(u.last_name, ''), COALESCE(u.profile_picture, ''),
            CASE
                WHEN c.last_message_id IS NULL THEN 1
                WHEN cm.sender_id = ? THEN 1
                WHEN cp.last_read_message_id IS NOT NULL AND cp.last_read_message_id >= c.last_message_id THEN 1
                ELSE 0
            END as seen
        FROM chats c
        JOIN chat_participants cp ON cp.chat_id = c.id AND cp.user_id = ? AND cp.left_at IS NULL
        LEFT JOIN chat_messages cm ON cm.id = c.last_message_id
        LEFT JOIN users u ON c.type = 'direct' AND u.id = CASE WHEN c.user_low_id = ? THEN c.user_high_id ELSE c.user_low_id END
        ORDER BY COALESCE(c.last_message_at, c.created_at) DESC, c.id DESC
        LIMIT ? OFFSET ?
    `, userID, userID, userID, userID, limit, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*models.ChatSummary
	for rows.Next() {
		item := new(models.ChatSummary)
		var seenInt int
		if err := rows.Scan(
			&item.ChatID, &item.Type, &item.GroupID, &item.LastMessageID,
			&item.LastMessageSender, &item.LastMessageType, &item.LastMessagePreview,
			&item.LastMessageAt, &item.OtherUserID, &item.OtherUserFirstName,
			&item.OtherUserLastName, &item.OtherUserPicture, &seenInt,
		); err != nil {
			return nil, err
		}
		item.Seen = seenInt == 1
		out = append(out, item)
	}
	return out, nil
}

func (r *sqliteChatRepo) GetGroupParticipantIDs(ctx context.Context, groupID int) ([]int, error) {
	query := `
        SELECT user_id 
        FROM group_members 
        WHERE group_id = ? AND status = 'accepted';`

	rows, err := r.db.QueryContext(ctx, query, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch group participants: %w", err)
	}
	defer rows.Close()

	var participantIDs []int
	for rows.Next() {
		var uid int
		if err := rows.Scan(&uid); err != nil {
			return nil, err
		}
		participantIDs = append(participantIDs, uid)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return participantIDs, nil
}
