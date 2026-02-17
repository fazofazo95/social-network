package queries

import (
	"backend/pkg/models"
	"context"
	"database/sql"
)

func GetPostComments(ctx context.Context, db *sql.DB, postID int, viewerID int) ([]models.Comment, error) {

	query := `
        SELECT 
            c.id, c.user_id, c.parent_type, c.parent_id, c.content, 
            COALESCE(c.extra_content, ''), c.created_at,
            u.first_name, u.last_name, COALESCE(u.profile_picture, '')
        FROM active_comments c
        INNER JOIN users u ON c.user_id = u.id
        INNER JOIN active_posts p ON c.parent_id = p.id AND c.parent_type = 'post'
        WHERE c.parent_id = ? 
        AND (
            p.user_id = ? 
            OR p.privacy = 'public'
            OR (p.privacy = 'followers' AND EXISTS (
                SELECT 1 FROM followers WHERE follower_id = ? AND followed_id = p.user_id
            ))
            OR (p.privacy = 'custom' AND EXISTS (
                SELECT 1 FROM post_permissions WHERE post_id = p.id AND user_id = ?
            ))
        )
        ORDER BY c.created_at ASC;`

	rows, err := db.QueryContext(ctx, query, postID, viewerID, viewerID, viewerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var c models.Comment
		err := rows.Scan(
			&c.ID, &c.UserID, &c.ParentType, &c.ParentID, &c.Content,
			&c.ExtraContent, &c.CreatedAt,
			&c.AuthorFirstName, &c.AuthorLastName, &c.AuthorProfilePicture,
		)
		if err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, nil
}

func GetCommentByID(ctx context.Context, db *sql.DB, commentID int) (*models.Comment, error) {
	query := `
        SELECT 
            c.id, c.user_id, c.parent_type, c.parent_id, c.content, 
            COALESCE(c.extra_content, ''), c.created_at,
            u.first_name, u.last_name, COALESCE(u.profile_picture, '')
        FROM active_comments c
        INNER JOIN users u ON c.user_id = u.id
        WHERE c.id = ?;`

	var c models.Comment
	err := db.QueryRowContext(ctx, query, commentID).Scan(
		&c.ID, &c.UserID, &c.ParentType, &c.ParentID, &c.Content,
		&c.ExtraContent, &c.CreatedAt,
		&c.AuthorFirstName, &c.AuthorLastName, &c.AuthorProfilePicture,
	)

	if err != nil {
		return nil, err
	}
	return &c, nil
}

func CreateComment(ctx context.Context, db *sql.DB, comment models.Comment) (int, error) {
	query := `
        INSERT INTO comments (user_id, parent_type, parent_id, content, extra_content)
        VALUES (?, ?, ?, ?, ?);`

	result, err := db.ExecContext(ctx, query,
		comment.UserID, comment.ParentType, comment.ParentID,
		comment.Content, comment.ExtraContent,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	return int(id), err
}

func UpdateComment(ctx context.Context, db *sql.DB, commentID int, content string) error {
	query := "UPDATE comments SET content = ? WHERE id = ? AND deleted_at IS NULL"
	_, err := db.ExecContext(ctx, query, content, commentID)
	return err
}

func DeleteComment(ctx context.Context, db *sql.DB, commentID int) error {
	query := "UPDATE comments SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?"
	_, err := db.ExecContext(ctx, query, commentID)
	return err
}

func RestoreComment(ctx context.Context, db *sql.DB, commentID int) error {
	query := "UPDATE comments SET deleted_at = NULL WHERE id = ?"
	_, err := db.ExecContext(ctx, query, commentID)
	return err
}

func GetCommentOwnerID(ctx context.Context, db *sql.DB, commentID int) (int, error) {
	var ownerID int
	query := "SELECT user_id FROM comments WHERE id = ?"
	err := db.QueryRowContext(ctx, query, commentID).Scan(&ownerID)
	if err != nil {
		return 0, err
	}
	return ownerID, nil
}
