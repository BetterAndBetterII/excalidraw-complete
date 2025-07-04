package sqlite

import (
	"context"
	"database/sql"
	"excalidraw-complete/core"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/sirupsen/logrus"
)

type sessionStore struct {
	db *sql.DB
}

func NewSessionStore(db *sql.DB) core.SessionStore {
	// 创建会话表
	sts := `CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		expires_at DATETIME NOT NULL,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);`
	
	_, err := db.Exec(sts)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create sessions table")
	}
	
	return &sessionStore{db: db}
}

func (s *sessionStore) CreateSession(ctx context.Context, session *core.Session) error {
	if session.ID == "" {
		session.ID = ulid.Make().String()
	}
	session.CreatedAt = time.Now()

	query := `INSERT INTO sessions (id, user_id, expires_at, created_at) 
		VALUES (?, ?, ?, ?)`
	
	_, err := s.db.ExecContext(ctx, query, 
		session.ID, session.UserID, session.ExpiresAt, session.CreatedAt)
	
	if err != nil {
		logrus.WithError(err).Error("Failed to create session")
		return err
	}
	
	logrus.WithFields(logrus.Fields{
		"session_id": session.ID,
		"user_id":    session.UserID,
	}).Info("Session created successfully")
	return nil
}

func (s *sessionStore) GetSession(ctx context.Context, id string) (*core.Session, error) {
	session := &core.Session{}
	query := `SELECT id, user_id, expires_at, created_at 
		FROM sessions WHERE id = ? AND expires_at > ?`
	
	err := s.db.QueryRowContext(ctx, query, id, time.Now()).Scan(
		&session.ID, &session.UserID, &session.ExpiresAt, &session.CreatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 会话不存在或已过期
		}
		logrus.WithError(err).Error("Failed to get session")
		return nil, err
	}
	
	return session, nil
}

func (s *sessionStore) DeleteSession(ctx context.Context, id string) error {
	query := `DELETE FROM sessions WHERE id = ?`
	
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		logrus.WithError(err).Error("Failed to delete session")
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logrus.WithError(err).Error("Failed to get rows affected")
		return err
	}
	
	if rowsAffected > 0 {
		logrus.WithField("session_id", id).Info("Session deleted successfully")
	}
	
	return nil
}

func (s *sessionStore) CleanExpiredSessions(ctx context.Context) error {
	query := `DELETE FROM sessions WHERE expires_at <= ?`
	
	result, err := s.db.ExecContext(ctx, query, time.Now())
	if err != nil {
		logrus.WithError(err).Error("Failed to clean expired sessions")
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logrus.WithError(err).Error("Failed to get rows affected")
		return err
	}
	
	if rowsAffected > 0 {
		logrus.WithField("cleaned_sessions", rowsAffected).Info("Expired sessions cleaned")
	}
	
	return nil
}