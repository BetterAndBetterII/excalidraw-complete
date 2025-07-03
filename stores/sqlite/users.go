package sqlite

import (
	"context"
	"database/sql"
	"excalidraw-complete/core"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/sirupsen/logrus"
)

type userStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) core.UserStore {
	// 创建用户表
	sts := `CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		github_id INTEGER UNIQUE NOT NULL,
		username TEXT NOT NULL,
		name TEXT NOT NULL,
		email TEXT,
		avatar_url TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_users_github_id ON users(github_id);`
	
	_, err := db.Exec(sts)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create users table")
	}
	
	return &userStore{db: db}
}

func (s *userStore) CreateUser(ctx context.Context, user *core.User) error {
	if user.ID == "" {
		user.ID = ulid.Make().String()
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	query := `INSERT INTO users (id, github_id, username, name, email, avatar_url, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	
	_, err := s.db.ExecContext(ctx, query, 
		user.ID, user.GitHubID, user.Username, user.Name, 
		user.Email, user.AvatarURL, user.CreatedAt, user.UpdatedAt)
	
	if err != nil {
		logrus.WithError(err).Error("Failed to create user")
		return err
	}
	
	logrus.WithField("user_id", user.ID).Info("User created successfully")
	return nil
}

func (s *userStore) GetUserByGitHubID(ctx context.Context, githubID int64) (*core.User, error) {
	user := &core.User{}
	query := `SELECT id, github_id, username, name, email, avatar_url, created_at, updated_at 
		FROM users WHERE github_id = ?`
	
	err := s.db.QueryRowContext(ctx, query, githubID).Scan(
		&user.ID, &user.GitHubID, &user.Username, &user.Name,
		&user.Email, &user.AvatarURL, &user.CreatedAt, &user.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 用户不存在
		}
		logrus.WithError(err).Error("Failed to get user by GitHub ID")
		return nil, err
	}
	
	return user, nil
}

func (s *userStore) GetUserByID(ctx context.Context, id string) (*core.User, error) {
	user := &core.User{}
	query := `SELECT id, github_id, username, name, email, avatar_url, created_at, updated_at 
		FROM users WHERE id = ?`
	
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.GitHubID, &user.Username, &user.Name,
		&user.Email, &user.AvatarURL, &user.CreatedAt, &user.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 用户不存在
		}
		logrus.WithError(err).Error("Failed to get user by ID")
		return nil, err
	}
	
	return user, nil
}

func (s *userStore) UpdateUser(ctx context.Context, user *core.User) error {
	user.UpdatedAt = time.Now()
	
	query := `UPDATE users SET username = ?, name = ?, email = ?, avatar_url = ?, updated_at = ? 
		WHERE id = ?`
	
	_, err := s.db.ExecContext(ctx, query, 
		user.Username, user.Name, user.Email, user.AvatarURL, user.UpdatedAt, user.ID)
	
	if err != nil {
		logrus.WithError(err).Error("Failed to update user")
		return err
	}
	
	logrus.WithField("user_id", user.ID).Info("User updated successfully")
	return nil
}