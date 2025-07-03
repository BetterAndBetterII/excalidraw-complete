package core

import (
	"bytes"
	"context"
	"time"
)

type (
	Document struct {
		Data bytes.Buffer
	}

	// 新增用户实体
	User struct {
		ID        string    `json:"id" db:"id"`
		GitHubID  int64     `json:"github_id" db:"github_id"`
		Username  string    `json:"username" db:"username"`
		Name      string    `json:"name" db:"name"`
		Email     string    `json:"email" db:"email"`
		AvatarURL string    `json:"avatar_url" db:"avatar_url"`
		CreatedAt time.Time `json:"created_at" db:"created_at"`
		UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	}

	// 新增画布实体
	Canvas struct {
		ID          string    `json:"id" db:"id"`
		UserID      string    `json:"user_id" db:"user_id"`
		Title       string    `json:"title" db:"title"`
		Data        []byte    `json:"data" db:"data"`
		IsPublic    bool      `json:"is_public" db:"is_public"`
		CreatedAt   time.Time `json:"created_at" db:"created_at"`
		UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	}

	// 会话实体
	Session struct {
		ID        string    `json:"id" db:"id"`
		UserID    string    `json:"user_id" db:"user_id"`
		ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
		CreatedAt time.Time `json:"created_at" db:"created_at"`
	}

	DocumentStore interface {
		FindID(ctx context.Context, id string) (*Document, error)
		Create(ctx context.Context, document *Document) (string, error)
	}

	// 新增用户存储接口
	UserStore interface {
		CreateUser(ctx context.Context, user *User) error
		GetUserByGitHubID(ctx context.Context, githubID int64) (*User, error)
		GetUserByID(ctx context.Context, id string) (*User, error)
		UpdateUser(ctx context.Context, user *User) error
	}

	// 新增画布存储接口
	CanvasStore interface {
		CreateCanvas(ctx context.Context, canvas *Canvas) error
		GetCanvasById(ctx context.Context, id string) (*Canvas, error)
		GetCanvasByUser(ctx context.Context, userID string, limit int, offset int) ([]*Canvas, error)
		UpdateCanvas(ctx context.Context, canvas *Canvas) error
		DeleteCanvas(ctx context.Context, id string, userID string) error
	}

	// 新增会话存储接口
	SessionStore interface {
		CreateSession(ctx context.Context, session *Session) error
		GetSession(ctx context.Context, id string) (*Session, error)
		DeleteSession(ctx context.Context, id string) error
		CleanExpiredSessions(ctx context.Context) error
	}
)
