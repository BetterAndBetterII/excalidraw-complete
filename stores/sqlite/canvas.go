package sqlite

import (
	"context"
	"database/sql"
	"excalidraw-complete/core"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/sirupsen/logrus"
)

type canvasStore struct {
	db *sql.DB
}

func NewCanvasStore(db *sql.DB) core.CanvasStore {
	// 创建画布表
	sts := `CREATE TABLE IF NOT EXISTS canvas (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		title TEXT NOT NULL,
		data BLOB,
		is_public BOOLEAN DEFAULT FALSE,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_canvas_user_id ON canvas(user_id);
	CREATE INDEX IF NOT EXISTS idx_canvas_created_at ON canvas(created_at);`
	
	_, err := db.Exec(sts)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create canvas table")
	}
	
	return &canvasStore{db: db}
}

func (s *canvasStore) CreateCanvas(ctx context.Context, canvas *core.Canvas) error {
	if canvas.ID == "" {
		canvas.ID = ulid.Make().String()
	}
	canvas.CreatedAt = time.Now()
	canvas.UpdatedAt = time.Now()

	query := `INSERT INTO canvas (id, user_id, title, data, is_public, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	
	_, err := s.db.ExecContext(ctx, query, 
		canvas.ID, canvas.UserID, canvas.Title, canvas.Data, 
		canvas.IsPublic, canvas.CreatedAt, canvas.UpdatedAt)
	
	if err != nil {
		logrus.WithError(err).Error("Failed to create canvas")
		return err
	}
	
	logrus.WithFields(logrus.Fields{
		"canvas_id": canvas.ID,
		"user_id":   canvas.UserID,
		"title":     canvas.Title,
	}).Info("Canvas created successfully")
	return nil
}

func (s *canvasStore) GetCanvasById(ctx context.Context, id string) (*core.Canvas, error) {
	canvas := &core.Canvas{}
	query := `SELECT id, user_id, title, data, is_public, created_at, updated_at 
		FROM canvas WHERE id = ?`
	
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&canvas.ID, &canvas.UserID, &canvas.Title, &canvas.Data,
		&canvas.IsPublic, &canvas.CreatedAt, &canvas.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 画布不存在
		}
		logrus.WithError(err).Error("Failed to get canvas by ID")
		return nil, err
	}
	
	return canvas, nil
}

func (s *canvasStore) GetCanvasByUser(ctx context.Context, userID string, limit int, offset int) ([]*core.Canvas, error) {
	query := `SELECT id, user_id, title, data, is_public, created_at, updated_at 
		FROM canvas WHERE user_id = ? ORDER BY updated_at DESC LIMIT ? OFFSET ?`
	
	rows, err := s.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to get canvas by user")
		return nil, err
	}
	defer rows.Close()

	var canvases []*core.Canvas
	for rows.Next() {
		canvas := &core.Canvas{}
		err := rows.Scan(
			&canvas.ID, &canvas.UserID, &canvas.Title, &canvas.Data,
			&canvas.IsPublic, &canvas.CreatedAt, &canvas.UpdatedAt)
		if err != nil {
			logrus.WithError(err).Error("Failed to scan canvas row")
			return nil, err
		}
		canvases = append(canvases, canvas)
	}
	
	return canvases, nil
}

func (s *canvasStore) UpdateCanvas(ctx context.Context, canvas *core.Canvas) error {
	canvas.UpdatedAt = time.Now()
	
	query := `UPDATE canvas SET title = ?, data = ?, is_public = ?, updated_at = ? 
		WHERE id = ? AND user_id = ?`
	
	result, err := s.db.ExecContext(ctx, query, 
		canvas.Title, canvas.Data, canvas.IsPublic, canvas.UpdatedAt, 
		canvas.ID, canvas.UserID)
	
	if err != nil {
		logrus.WithError(err).Error("Failed to update canvas")
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logrus.WithError(err).Error("Failed to get rows affected")
		return err
	}
	
	if rowsAffected == 0 {
		logrus.WithFields(logrus.Fields{
			"canvas_id": canvas.ID,
			"user_id":   canvas.UserID,
		}).Warn("No canvas updated - either canvas doesn't exist or user doesn't own it")
		return nil
	}
	
	logrus.WithFields(logrus.Fields{
		"canvas_id": canvas.ID,
		"user_id":   canvas.UserID,
	}).Info("Canvas updated successfully")
	return nil
}

func (s *canvasStore) DeleteCanvas(ctx context.Context, id string, userID string) error {
	query := `DELETE FROM canvas WHERE id = ? AND user_id = ?`
	
	result, err := s.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		logrus.WithError(err).Error("Failed to delete canvas")
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logrus.WithError(err).Error("Failed to get rows affected")
		return err
	}
	
	if rowsAffected == 0 {
		logrus.WithFields(logrus.Fields{
			"canvas_id": id,
			"user_id":   userID,
		}).Warn("No canvas deleted - either canvas doesn't exist or user doesn't own it")
		return nil
	}
	
	logrus.WithFields(logrus.Fields{
		"canvas_id": id,
		"user_id":   userID,
	}).Info("Canvas deleted successfully")
	return nil
}