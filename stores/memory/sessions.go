package memory

import (
	"context"
	"excalidraw-complete/core"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

type sessionStore struct {
	sessions map[string]*core.Session
	mutex    sync.RWMutex
}

func NewSessionStore() core.SessionStore {
	return &sessionStore{
		sessions: make(map[string]*core.Session),
	}
}

func (s *sessionStore) CreateSession(ctx context.Context, session *core.Session) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if session.ID == "" {
		session.ID = ulid.Make().String()
	}
	session.CreatedAt = time.Now()
	
	s.sessions[session.ID] = session
	
	return nil
}

func (s *sessionStore) GetSession(ctx context.Context, id string) (*core.Session, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	session, exists := s.sessions[id]
	if !exists {
		return nil, nil
	}
	
	// 检查会话是否过期
	if session.ExpiresAt.Before(time.Now()) {
		// 删除过期的会话
		delete(s.sessions, id)
		return nil, nil
	}
	
	// 返回副本以避免并发修改
	sessionCopy := *session
	return &sessionCopy, nil
}

func (s *sessionStore) DeleteSession(ctx context.Context, id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	delete(s.sessions, id)
	
	return nil
}

func (s *sessionStore) CleanExpiredSessions(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	now := time.Now()
	for id, session := range s.sessions {
		if session.ExpiresAt.Before(now) {
			delete(s.sessions, id)
		}
	}
	
	return nil
}