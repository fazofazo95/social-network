package repository

import (
	"backend/pkg/models"
	"context"
	"database/sql"
	"log"
)

type CommentRepository interface {
	CreateComment(ctx context.Context, comment models.Comment) (int, error)
	GetCommentByID(ctx context.Context, commentID int) (*models.Comment, error)
	GetPostComments(ctx context.Context, postID int, viewerID int) ([]*models.Comment, error)
	GetCommentOwnerID(ctx context.Context, commentID int) (int, error)
	UpdateComment(ctx context.Context, commentID int, content string) error
	DeleteComment(ctx context.Context, commentID int) error
	RestoreComment(ctx context.Context, commentID int) error
}

type sqliteCommentRepo struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) CommentRepository {
	return &sqliteCommentRepo{db: db}
}

func (r *sqliteCommentRepo) CreateComment(ctx context.Context, comment models.Comment) (int, error) {
	query := `
        INSERT INTO comments (user_id, parent_type, parent_id, content, extra_content)
        VALUES (?, ?, ?, ?, ?);`

	result, err := r.db.ExecContext(ctx, query,
		comment.UserID, comment.ParentType, comment.ParentID,
		comment.Content, comment.ExtraContent,
	)
	if err != nil {
		log.Printf("[ERROR] CreateComment insert failed: %v", err)
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("[ERROR] CreateComment could not get LastInsertId: %v", err)
		return 0, err
	}

	return int(id), nil
}

func (r *sqliteCommentRepo) GetCommentByID(ctx context.Context, commentID int) (*models.Comment, error) {
	query := `
        SELECT 
            c.id, c.user_id, c.parent_type, c.parent_id, c.content, 
            COALESCE(c.extra_content, ''), c.created_at,
            u.first_name, u.last_name, COALESCE(u.profile_picture, '')
        FROM active_comments c
        INNER JOIN users u ON c.user_id = u.id
        WHERE c.id = ?;`

	c := new(models.Comment)
	err := r.db.QueryRowContext(ctx, query, commentID).Scan(
		&c.ID, &c.UserID, &c.ParentType, &c.ParentID, &c.Content,
		&c.ExtraContent, &c.CreatedAt,
		&c.AuthorFirstName, &c.AuthorLastName, &c.AuthorProfilePicture,
	)

	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("[ERROR] GetCommentByID failed: %v", err)
		}
		return nil, err
	}
	return c, nil
}

func (r *sqliteCommentRepo) GetPostComments(ctx context.Context, postID int, viewerID int) ([]*models.Comment, error) {
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
            OR (p.group_id IS NOT NULL AND EXISTS (
                SELECT 1 FROM group_members gm
                WHERE gm.group_id = p.group_id AND gm.user_id = ? AND gm.status = 'active'
            ))
        )
        ORDER BY c.created_at ASC;`

	rows, err := r.db.QueryContext(ctx, query, postID, viewerID, viewerID, viewerID, viewerID)
	if err != nil {
		log.Printf("[ERROR] GetPostComments query failed: %v", err)
		return nil, err
	}
	defer rows.Close()

	comments := make([]*models.Comment, 0)
	for rows.Next() {
		c := new(models.Comment)
		err := rows.Scan(
			&c.ID, &c.UserID, &c.ParentType, &c.ParentID, &c.Content,
			&c.ExtraContent, &c.CreatedAt,
			&c.AuthorFirstName, &c.AuthorLastName, &c.AuthorProfilePicture,
		)
		if err != nil {
			log.Printf("[ERROR] GetPostComments scan failed: %v", err)
			return nil, err
		}
		comments = append(comments, c)
	}

	return comments, nil
}

func (r *sqliteCommentRepo) GetCommentOwnerID(ctx context.Context, commentID int) (int, error) {
	var ownerID int
	query := "SELECT user_id FROM comments WHERE id = ?"
	err := r.db.QueryRowContext(ctx, query, commentID).Scan(&ownerID)
	if err != nil {
		log.Printf("[ERROR] GetCommentOwnerID failed: %v", err)
		return 0, err
	}
	return ownerID, nil
}

func (r *sqliteCommentRepo) UpdateComment(ctx context.Context, commentID int, content string) error {
	query := "UPDATE comments SET content = ? WHERE id = ? AND deleted_at IS NULL"
	_, err := r.db.ExecContext(ctx, query, content, commentID)
	if err != nil {
		log.Printf("[ERROR] UpdateComment failed: %v", err)
		return err
	}

	return nil
}

func (r *sqliteCommentRepo) DeleteComment(ctx context.Context, commentID int) error {
	query := "UPDATE comments SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?"
	_, err := r.db.ExecContext(ctx, query, commentID)
	if err != nil {
		log.Printf("[ERROR] DeleteComment failed: %v", err)
	}
	return err
}

func (r *sqliteCommentRepo) RestoreComment(ctx context.Context, commentID int) error {
	query := "UPDATE comments SET deleted_at = NULL WHERE id = ?"
	_, err := r.db.ExecContext(ctx, query, commentID)
	if err != nil {
		log.Printf("[ERROR] RestoreComment failed: %v", err)
	}
	return err
}
