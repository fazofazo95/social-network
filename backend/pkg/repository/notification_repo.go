package repository

import (
	"backend/pkg/models"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

var (
	ErrNotificationNotFound      = errors.New("notification not found")
	ErrNotificationUnauthorized  = errors.New("notification does not belong to user")
	ErrNotificationNotActionable = errors.New("notification is not actionable")
	ErrNotificationAlreadyClosed = errors.New("notification is already resolved")
	ErrNotificationUserNotFound  = errors.New("user not found")
)

type NotificationRepository interface {
	Create(ctx context.Context, in models.CreateNotificationInput) (*models.NotificationWithActor, error)
	ListByRecipient(ctx context.Context, recipientID, limit, offset int) ([]models.NotificationWithActor, error)
	GetByID(ctx context.Context, notificationID int) (models.NotificationWithActor, error)
	GetByIDForRecipient(ctx context.Context, notificationID, recipientID int) (models.NotificationWithActor, error)
	MarkSeen(ctx context.Context, notificationID, recipientID int) error
	MarkAllSeen(ctx context.Context, recipientID int) (int64, error)
	UpdateStatus(ctx context.Context, notificationID, recipientID int, status string) error
	MarkSeenAndStatus(ctx context.Context, notificationID, recipientID int, status string) error
	GetGroupOwnerID(ctx context.Context, groupID int) (int, error)
	GetGroupActiveMemberIDs(ctx context.Context, groupID int, excludeUserID int) ([]int, error)
	GetGroupNameByID(ctx context.Context, groupID int) (string, error)
	GetUserDisplayName(ctx context.Context, userID int) (string, error)
}

type sqliteNotificationRepo struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) NotificationRepository {
	return &sqliteNotificationRepo{db: db}
}

func (r *sqliteNotificationRepo) Create(ctx context.Context, in models.CreateNotificationInput) (*models.NotificationWithActor, error) {
	metadata := in.Metadata
	if metadata == "" {
		metadata = "{}"
	}
	if !json.Valid([]byte(metadata)) {
		return nil, errors.New("invalid metadata JSON")
	}

	res, err := r.db.ExecContext(ctx, `
		INSERT INTO notifications (recipient_id, actor_id, type, status, group_id, event_id, content, metadata, seen)
		VALUES (?, ?, ?, 'pending', ?, ?, ?, ?, 0)
	`, in.RecipientID, in.ActorID, string(in.Type), in.GroupID, in.EventID, in.Content, metadata)
	if err != nil {
		return nil, err
	}

	id64, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	out, err := r.GetByID(ctx, int(id64))
	if err != nil {
		return nil, err
	}

	return &out, nil
}

func (r *sqliteNotificationRepo) ListByRecipient(ctx context.Context, recipientID, limit, offset int) ([]models.NotificationWithActor, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT n.id, n.recipient_id, n.actor_id, n.type, n.status, n.group_id, n.event_id,
		       n.content, n.metadata, n.seen, COALESCE(datetime(n.created_at), ''), COALESCE(datetime(n.updated_at), ''),
		       COALESCE(u.first_name, ''), COALESCE(u.last_name, ''), COALESCE(u.profile_picture, '')
		FROM notifications n
		LEFT JOIN users u ON u.id = n.actor_id
		WHERE n.recipient_id = ?
		ORDER BY n.created_at DESC, n.id DESC
		LIMIT ? OFFSET ?
	`, recipientID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.NotificationWithActor, 0)
	for rows.Next() {
		item, err := scanNotificationWithActor(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *sqliteNotificationRepo) GetByID(ctx context.Context, notificationID int) (models.NotificationWithActor, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT n.id, n.recipient_id, n.actor_id, n.type, n.status, n.group_id, n.event_id,
		       n.content, n.metadata, n.seen, COALESCE(datetime(n.created_at), ''), COALESCE(datetime(n.updated_at), ''),
		       COALESCE(u.first_name, ''), COALESCE(u.last_name, ''), COALESCE(u.profile_picture, '')
		FROM notifications n
		LEFT JOIN users u ON u.id = n.actor_id
		WHERE n.id = ?
	`, notificationID)

	item, err := scanNotificationWithActor(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.NotificationWithActor{}, ErrNotificationNotFound
		}
		return models.NotificationWithActor{}, err
	}

	return item, nil
}

func (r *sqliteNotificationRepo) GetByIDForRecipient(ctx context.Context, notificationID, recipientID int) (models.NotificationWithActor, error) {
	item, err := r.GetByID(ctx, notificationID)
	if err != nil {
		return models.NotificationWithActor{}, err
	}
	if item.RecipientID != recipientID {
		return models.NotificationWithActor{}, ErrNotificationUnauthorized
	}
	return item, nil
}

func (r *sqliteNotificationRepo) MarkSeen(ctx context.Context, notificationID, recipientID int) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE notifications
		SET seen = 1,
		    status = CASE WHEN status = 'pending' THEN 'read' ELSE status END,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND recipient_id = ?
	`, notificationID, recipientID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotificationNotFound
	}

	return nil
}

func (r *sqliteNotificationRepo) MarkAllSeen(ctx context.Context, recipientID int) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
		UPDATE notifications
		SET seen = 1,
		    status = CASE WHEN status = 'pending' THEN 'read' ELSE status END,
		    updated_at = CURRENT_TIMESTAMP
		WHERE recipient_id = ? AND seen = 0
	`, recipientID)
	if err != nil {
		return 0, err
	}

	return res.RowsAffected()
}

func (r *sqliteNotificationRepo) UpdateStatus(ctx context.Context, notificationID, recipientID int, status string) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE notifications
		SET status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND recipient_id = ?
	`, status, notificationID, recipientID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotificationNotFound
	}

	return nil
}

func (r *sqliteNotificationRepo) MarkSeenAndStatus(ctx context.Context, notificationID, recipientID int, status string) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE notifications
		SET seen = 1,
		    status = ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND recipient_id = ?
	`, status, notificationID, recipientID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotificationNotFound
	}

	return nil
}

func (r *sqliteNotificationRepo) GetGroupOwnerID(ctx context.Context, groupID int) (int, error) {
	var ownerID int
	err := r.db.QueryRowContext(ctx, `SELECT owner_id FROM groups WHERE id = ?`, groupID).Scan(&ownerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrGroupNotFound
		}
		return 0, err
	}
	return ownerID, nil
}

func (r *sqliteNotificationRepo) GetGroupActiveMemberIDs(ctx context.Context, groupID int, excludeUserID int) ([]int, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT user_id
		FROM group_members
		WHERE group_id = ? AND status = 'active' AND user_id <> ?
	`, groupID, excludeUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]int, 0)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ids, nil
}

func (r *sqliteNotificationRepo) GetGroupNameByID(ctx context.Context, groupID int) (string, error) {
	var name string
	err := r.db.QueryRowContext(ctx, `SELECT name FROM groups WHERE id = ?`, groupID).Scan(&name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrGroupNotFound
		}
		return "", err
	}
	return name, nil
}

func (r *sqliteNotificationRepo) GetUserDisplayName(ctx context.Context, userID int) (string, error) {
	var firstName, lastName string
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(first_name, ''), COALESCE(last_name, '')
		FROM users
		WHERE id = ?
	`, userID).Scan(&firstName, &lastName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotificationUserNotFound
		}
		return "", err
	}

	if lastName == "" {
		return firstName, nil
	}

	return firstName + " " + lastName, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanNotificationWithActor(row scanner) (models.NotificationWithActor, error) {
	var item models.NotificationWithActor
	var typ string

	err := row.Scan(
		&item.ID,
		&item.RecipientID,
		&item.ActorID,
		&typ,
		&item.Status,
		&item.GroupID,
		&item.EventID,
		&item.Content,
		&item.Metadata,
		&item.Seen,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.ActorFirstName,
		&item.ActorLastName,
		&item.ActorPicture,
	)
	if err != nil {
		return models.NotificationWithActor{}, err
	}

	item.Type = models.NotificationType(typ)
	return item, nil
}
