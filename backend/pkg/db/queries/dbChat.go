package queries

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"backend/pkg/models"
)

var ErrInvalidChatMessage = errors.New("invalid chat message")
var ErrChatNotFound = errors.New("chat not found")
var ErrChatForbidden = errors.New("chat access forbidden")
var ErrDirectChatNotAllowed = errors.New("direct chat not allowed")
var ErrGroupChatNotFound = errors.New("group chat not found")

func normalizePair(a, b int) (int, int) {
	if a < b {
		return a, b
	}
	return b, a
}

func validateMessageInput(in models.SendMessageInput) (models.SendMessageInput, error) {
	in.MessageType = strings.ToLower(strings.TrimSpace(in.MessageType))
	in.Body = strings.TrimSpace(in.Body)
	in.MediaURL = strings.TrimSpace(in.MediaURL)

	if in.MessageType == "" {
		in.MessageType = "text"
	}

	switch in.MessageType {
	case "text":
		if in.Body == "" {
			return in, ErrInvalidChatMessage
		}
	case "image":
		if in.MediaURL == "" {
			return in, ErrInvalidChatMessage
		}
	case "text_image":
		if in.Body == "" || in.MediaURL == "" {
			return in, ErrInvalidChatMessage
		}
	default:
		return in, ErrInvalidChatMessage
	}

	return in, nil
}

func canUsersDirectChat(ctx context.Context, tx *sql.Tx, userA, userB int) (bool, error) {
	var blockedCount int
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM followers
		WHERE ((follower_id = ? AND followed_id = ?) OR (follower_id = ? AND followed_id = ?))
		  AND status = 'blocked'
	`, userA, userB, userB, userA).Scan(&blockedCount); err != nil {
		return false, err
	}
	if blockedCount > 0 {
		return false, nil
	}

	var acceptedCount int
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM followers
		WHERE ((follower_id = ? AND followed_id = ?) OR (follower_id = ? AND followed_id = ?))
		  AND status = 'accepted'
	`, userA, userB, userB, userA).Scan(&acceptedCount); err != nil {
		return false, err
	}

	return acceptedCount > 0, nil
}

func ensureDirectChatTx(ctx context.Context, tx *sql.Tx, userID, targetUserID int) (int, error) {
	if userID == targetUserID {
		return 0, ErrDirectChatNotAllowed
	}

	allowed, err := canUsersDirectChat(ctx, tx, userID, targetUserID)
	if err != nil {
		return 0, err
	}
	if !allowed {
		return 0, ErrDirectChatNotAllowed
	}

	lowID, highID := normalizePair(userID, targetUserID)

	var chatID int
	err = tx.QueryRowContext(ctx, `
		SELECT id
		FROM chats
		WHERE type = 'direct' AND user_low_id = ? AND user_high_id = ?
		LIMIT 1
	`, lowID, highID).Scan(&chatID)
	if err == nil {
		return chatID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	res, err := tx.ExecContext(ctx, `
		INSERT INTO chats (type, user_low_id, user_high_id, created_by, last_message_at)
		VALUES ('direct', ?, ?, ?, CURRENT_TIMESTAMP)
	`, lowID, highID, userID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "idx_chats_direct_pair_unique") {
			if err := tx.QueryRowContext(ctx, `
				SELECT id FROM chats
				WHERE type = 'direct' AND user_low_id = ? AND user_high_id = ?
				LIMIT 1
			`, lowID, highID).Scan(&chatID); err != nil {
				return 0, err
			}
			return chatID, nil
		}
		return 0, err
	}

	chatID64, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	chatID = int(chatID64)

	if _, err := tx.ExecContext(ctx, `
		INSERT OR IGNORE INTO chat_participants (chat_id, user_id)
		VALUES (?, ?), (?, ?)
	`, chatID, lowID, chatID, highID); err != nil {
		return 0, err
	}

	return chatID, nil
}

func ensureUserCanAccessChatTx(ctx context.Context, tx *sql.Tx, userID, chatID int) error {
	var n int
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM chat_participants
		WHERE chat_id = ? AND user_id = ? AND left_at IS NULL
	`, chatID, userID).Scan(&n); err != nil {
		return err
	}
	if n == 0 {
		return ErrChatForbidden
	}
	return nil
}

func sendMessageToChatTx(ctx context.Context, tx *sql.Tx, chatID, senderID int, in models.SendMessageInput) (models.ChatMessage, error) {
	in, err := validateMessageInput(in)
	if err != nil {
		return models.ChatMessage{}, err
	}

	res, err := tx.ExecContext(ctx, `
		INSERT INTO chat_messages (chat_id, sender_id, message_type, body, media_url)
		VALUES (?, ?, ?, ?, ?)
	`, chatID, senderID, in.MessageType, nullableString(in.Body), nullableString(in.MediaURL))
	if err != nil {
		return models.ChatMessage{}, err
	}

	messageID64, err := res.LastInsertId()
	if err != nil {
		return models.ChatMessage{}, err
	}
	messageID := int(messageID64)

	if _, err := tx.ExecContext(ctx, `
		UPDATE chats
		SET last_message_id = ?, last_message_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, messageID, chatID); err != nil {
		return models.ChatMessage{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE chat_participants
		SET last_read_message_id = ?
		WHERE chat_id = ? AND user_id = ?
	`, messageID, chatID, senderID); err != nil {
		return models.ChatMessage{}, err
	}

	var out models.ChatMessage
	var body sql.NullString
	var media sql.NullString
	if err := tx.QueryRowContext(ctx, `
		SELECT id, chat_id, sender_id, message_type, body, media_url, COALESCE(datetime(created_at), '')
		FROM chat_messages
		WHERE id = ?
	`, messageID).Scan(&out.ID, &out.ChatID, &out.SenderID, &out.MessageType, &body, &media, &out.CreatedAt); err != nil {
		return models.ChatMessage{}, err
	}
	if body.Valid {
		out.Body = body.String
	}
	if media.Valid {
		out.MediaURL = media.String
	}

	return out, nil
}

func nullableString(s string) interface{} {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

func SendDirectMessage(ctx context.Context, db *sql.DB, senderID, targetUserID int, in models.SendMessageInput) (models.ChatMessage, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return models.ChatMessage{}, err
	}
	defer tx.Rollback()

	chatID, err := ensureDirectChatTx(ctx, tx, senderID, targetUserID)
	if err != nil {
		return models.ChatMessage{}, err
	}

	msg, err := sendMessageToChatTx(ctx, tx, chatID, senderID, in)
	if err != nil {
		return models.ChatMessage{}, err
	}

	if err := tx.Commit(); err != nil {
		return models.ChatMessage{}, err
	}

	return msg, nil
}

func SendGroupMessage(ctx context.Context, db *sql.DB, senderID, groupID int, in models.SendMessageInput) (models.ChatMessage, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return models.ChatMessage{}, err
	}
	defer tx.Rollback()

	var chatID int
	err = tx.QueryRowContext(ctx, `
		SELECT id
		FROM chats
		WHERE type = 'group' AND group_id = ?
		LIMIT 1
	`, groupID).Scan(&chatID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.ChatMessage{}, ErrGroupChatNotFound
		}
		return models.ChatMessage{}, err
	}

	var active int
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM group_members
		WHERE group_id = ? AND user_id = ? AND status = 'active'
	`, groupID, senderID).Scan(&active); err != nil {
		return models.ChatMessage{}, err
	}
	if active == 0 {
		return models.ChatMessage{}, ErrChatForbidden
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT OR IGNORE INTO chat_participants (chat_id, user_id)
		VALUES (?, ?)
	`, chatID, senderID); err != nil {
		return models.ChatMessage{}, err
	}

	msg, err := sendMessageToChatTx(ctx, tx, chatID, senderID, in)
	if err != nil {
		return models.ChatMessage{}, err
	}

	if err := tx.Commit(); err != nil {
		return models.ChatMessage{}, err
	}

	return msg, nil
}

func ListChats(ctx context.Context, db *sql.DB, userID, limit, offset int) ([]models.ChatSummary, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := db.QueryContext(ctx, `
		SELECT
			c.id,
			c.type,
			COALESCE(c.group_id, 0),
			COALESCE(c.last_message_id, 0),
			COALESCE(cm.sender_id, 0),
			COALESCE(cm.message_type, ''),
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
			COALESCE(u.first_name, ''),
			COALESCE(u.last_name, ''),
			COALESCE(u.profile_picture, ''),
			CASE
				WHEN c.last_message_id IS NULL THEN 1
				WHEN cm.sender_id = ? THEN 1
				WHEN cp.last_read_message_id IS NOT NULL AND cp.last_read_message_id >= c.last_message_id THEN 1
				ELSE 0
			END as seen
		FROM chats c
		JOIN chat_participants cp
			ON cp.chat_id = c.id
			AND cp.user_id = ?
			AND cp.left_at IS NULL
		LEFT JOIN chat_messages cm ON cm.id = c.last_message_id
		LEFT JOIN users u
			ON c.type = 'direct'
			AND u.id = CASE WHEN c.user_low_id = ? THEN c.user_high_id ELSE c.user_low_id END
		ORDER BY COALESCE(c.last_message_at, c.created_at) DESC, c.id DESC
		LIMIT ? OFFSET ?
	`, userID, userID, userID, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.ChatSummary, 0)
	for rows.Next() {
		var item models.ChatSummary
		var seenInt int
		if err := rows.Scan(
			&item.ChatID,
			&item.Type,
			&item.GroupID,
			&item.LastMessageID,
			&item.LastMessageSender,
			&item.LastMessageType,
			&item.LastMessagePreview,
			&item.LastMessageAt,
			&item.OtherUserID,
			&item.OtherUserFirstName,
			&item.OtherUserLastName,
			&item.OtherUserPicture,
			&seenInt,
		); err != nil {
			return nil, err
		}
		item.Seen = seenInt == 1
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func GetChatMessages(ctx context.Context, db *sql.DB, userID, chatID, beforeID, limit int) ([]models.ChatMessage, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var exists int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM chats WHERE id = ?`, chatID).Scan(&exists); err != nil {
		return nil, err
	}
	if exists == 0 {
		return nil, ErrChatNotFound
	}

	if err := ensureUserCanAccessChatTx(ctx, tx, userID, chatID); err != nil {
		return nil, err
	}

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

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	desc := make([]models.ChatMessage, 0)
	for rows.Next() {
		var m models.ChatMessage
		var body sql.NullString
		var media sql.NullString
		if err := rows.Scan(&m.ID, &m.ChatID, &m.SenderID, &m.MessageType, &body, &media, &m.CreatedAt); err != nil {
			return nil, err
		}
		if body.Valid {
			m.Body = body.String
		}
		if media.Valid {
			m.MediaURL = media.String
		}
		desc = append(desc, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	for i, j := 0, len(desc)-1; i < j; i, j = i+1, j-1 {
		desc[i], desc[j] = desc[j], desc[i]
	}

	return desc, nil
}

func MarkChatRead(ctx context.Context, db *sql.DB, userID, chatID, lastMessageID int) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM chats WHERE id = ?`, chatID).Scan(&exists); err != nil {
		return err
	}
	if exists == 0 {
		return ErrChatNotFound
	}

	if err := ensureUserCanAccessChatTx(ctx, tx, userID, chatID); err != nil {
		return err
	}

	if lastMessageID <= 0 {
		if err := tx.QueryRowContext(ctx, `
			SELECT COALESCE(last_message_id, 0)
			FROM chats
			WHERE id = ?
		`, chatID).Scan(&lastMessageID); err != nil {
			return err
		}
	}

	if lastMessageID > 0 {
		var messageChatID int
		if err := tx.QueryRowContext(ctx, `
			SELECT chat_id
			FROM chat_messages
			WHERE id = ?
		`, lastMessageID).Scan(&messageChatID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrChatNotFound
			}
			return err
		}
		if messageChatID != chatID {
			return fmt.Errorf("%w: message does not belong to chat", ErrInvalidChatMessage)
		}
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE chat_participants
		SET last_read_message_id = CASE
			WHEN last_read_message_id IS NULL OR last_read_message_id < ? THEN ?
			ELSE last_read_message_id
		END
		WHERE chat_id = ? AND user_id = ?
	`, lastMessageID, lastMessageID, chatID, userID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
