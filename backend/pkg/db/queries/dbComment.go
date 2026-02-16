package queries

import (
	"backend/pkg/models"
	"context"
	"database/sql"
)

func GetPostCommnets(ctx context.Context, db *sql.DB, postID int) ([]models.Comment, error) {
	query := `
		SELECT 
			c.id,
			c.user_id,
			c.parent_type,
			c.parent_id,
			c.content,
			COALESCE(c.extra_content, ''),
			c.created_at,
			u.first_name,
			u.last_name,
			COALESCE(u.profile_picture, '')
		FROM comments c
		INNER JOIN users u ON c.user_id = u.id
		WHERE c.parent_type = 'post' AND c.parent_id = ?;
	`

	rows, err := db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		err := rows.Scan(
			&comment.ID,
			&comment.UserID,
			&comment.ParentType,
			&comment.ParentID,
			&comment.Content,
			&comment.ExtraContent,
			&comment.CreatedAt,
			&comment.AuthorFirstName,
			&comment.AuthorLastName,
			&comment.AuthorProfilePicture,
		)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

func GetCommentByID(ctx context.Context, db *sql.DB, commentID int) (models.Comment, error) {
	query := `
		SELECT 
			c.id,
			c.user_id,
			c.parent_type,
			c.parent_id,
			c.content,
			COALESCE(c.extra_content, ''),
			c.created_at,
			u.first_name,
			u.last_name,
			COALESCE(u.profile_picture, '')
		FROM comments c
		INNER JOIN users u ON c.user_id = u.id
		WHERE c.id = ?;
	`

	var comment models.Comment
	err := db.QueryRowContext(ctx, query, commentID).Scan(
		&comment.ID,
		&comment.UserID,
		&comment.ParentType,
		&comment.ParentID,
		&comment.Content,
		&comment.ExtraContent,
		&comment.CreatedAt,
		&comment.AuthorFirstName,
		&comment.AuthorLastName,
		&comment.AuthorProfilePicture,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return models.Comment{}, nil
		}
		return models.Comment{}, err
	}

	return comment, nil
}

func CreateComment(ctx context.Context, db *sql.DB, comment models.Comment) (int, error) {
	query := `
		INSERT INTO comments (user_id, parent_type, parent_id, content, extra_content)
		VALUES (?, ?, ?, ?, ?);
	`

	result, err := db.ExecContext(ctx, query,
		comment.UserID,
		comment.ParentType,
		comment.ParentID,
		comment.Content,
		comment.ExtraContent,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}
