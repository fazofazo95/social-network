package queries

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"backend/pkg/models"
)

var ErrGroupNameTaken = errors.New("group name already in use")
var ErrGroupNotFound = errors.New("group not found")
var ErrNotGroupOwner = errors.New("only the group owner can delete the group")

func CreateGroup(ctx context.Context, db *sql.DB, ownerID int, in models.CreateGroupInput) (models.GroupResponse, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return models.GroupResponse{}, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `
		INSERT INTO groups (name, description, owner_id, visibility, join_mode, group_members)
		VALUES (?, ?, ?, ?, ?, 1)
	`, in.Name, in.Description, ownerID, in.Visibility, in.JoinMode)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "groups.name") {
			return models.GroupResponse{}, ErrGroupNameTaken
		}
		return models.GroupResponse{}, err
	}

	groupID64, err := res.LastInsertId()
	if err != nil {
		return models.GroupResponse{}, err
	}
	groupID := int(groupID64)

	_, err = tx.ExecContext(ctx, `
		INSERT INTO group_members (group_id, user_id, role, status)
		VALUES (?, ?, 'owner', 'active')
	`, groupID, ownerID)
	if err != nil {
		return models.GroupResponse{}, err
	}

	var out models.GroupResponse
	out.ID = groupID
	out.Name = in.Name
	out.Description = in.Description
	out.OwnerID = ownerID
	out.Visibility = in.Visibility
	out.JoinMode = in.JoinMode
	out.GroupMembers = 1

	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(datetime(created_at), '') FROM groups WHERE id = ?`, groupID).Scan(&out.CreatedAt); err != nil {
		return models.GroupResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return models.GroupResponse{}, err
	}

	return out, nil
}

func DeleteGroup(ctx context.Context, db *sql.DB, requesterID, groupID int) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var ownerID int
	if err := tx.QueryRowContext(ctx, `SELECT owner_id FROM groups WHERE id = ?`, groupID).Scan(&ownerID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrGroupNotFound
		}
		return err
	}

	if ownerID != requesterID {
		return ErrNotGroupOwner
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM group_join_requests WHERE group_id = ?`, groupID); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM group_members WHERE group_id = ?`, groupID); err != nil {
		return err
	}

	res, err := tx.ExecContext(ctx, `DELETE FROM groups WHERE id = ?`, groupID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrGroupNotFound
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
