package memory

import (
	"context"
	"excalidraw-complete/core"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

type userStore struct {
	users       map[string]*core.User
	usersByGitHubID map[int64]*core.User
	mutex       sync.RWMutex
}

func NewUserStore() core.UserStore {
	return &userStore{
		users:       make(map[string]*core.User),
		usersByGitHubID: make(map[int64]*core.User),
	}
}

func (s *userStore) CreateUser(ctx context.Context, user *core.User) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if user.ID == "" {
		user.ID = ulid.Make().String()
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	
	s.users[user.ID] = user
	s.usersByGitHubID[user.GitHubID] = user
	
	return nil
}

func (s *userStore) GetUserByGitHubID(ctx context.Context, githubID int64) (*core.User, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	user, exists := s.usersByGitHubID[githubID]
	if !exists {
		return nil, nil
	}
	
	// 返回副本以避免并发修改
	userCopy := *user
	return &userCopy, nil
}

func (s *userStore) GetUserByID(ctx context.Context, id string) (*core.User, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	user, exists := s.users[id]
	if !exists {
		return nil, nil
	}
	
	// 返回副本以避免并发修改
	userCopy := *user
	return &userCopy, nil
}

func (s *userStore) UpdateUser(ctx context.Context, user *core.User) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	existingUser, exists := s.users[user.ID]
	if !exists {
		return nil // 用户不存在
	}
	
	user.UpdatedAt = time.Now()
	user.CreatedAt = existingUser.CreatedAt // 保持原有的创建时间
	
	s.users[user.ID] = user
	s.usersByGitHubID[user.GitHubID] = user
	
	return nil
}