package memory

import (
	"context"
	"excalidraw-complete/core"
	"sort"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

type canvasStore struct {
	canvases map[string]*core.Canvas
	mutex    sync.RWMutex
}

func NewCanvasStore() core.CanvasStore {
	return &canvasStore{
		canvases: make(map[string]*core.Canvas),
	}
}

func (s *canvasStore) CreateCanvas(ctx context.Context, canvas *core.Canvas) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if canvas.ID == "" {
		canvas.ID = ulid.Make().String()
	}
	canvas.CreatedAt = time.Now()
	canvas.UpdatedAt = time.Now()
	
	s.canvases[canvas.ID] = canvas
	
	return nil
}

func (s *canvasStore) GetCanvasById(ctx context.Context, id string) (*core.Canvas, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	canvas, exists := s.canvases[id]
	if !exists {
		return nil, nil
	}
	
	// 返回副本以避免并发修改
	canvasCopy := *canvas
	if canvas.Data != nil {
		canvasCopy.Data = make([]byte, len(canvas.Data))
		copy(canvasCopy.Data, canvas.Data)
	}
	return &canvasCopy, nil
}

func (s *canvasStore) GetCanvasByUser(ctx context.Context, userID string, limit int, offset int) ([]*core.Canvas, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	var userCanvases []*core.Canvas
	
	// 收集用户的画布
	for _, canvas := range s.canvases {
		if canvas.UserID == userID {
			canvasCopy := *canvas
			if canvas.Data != nil {
				canvasCopy.Data = make([]byte, len(canvas.Data))
				copy(canvasCopy.Data, canvas.Data)
			}
			userCanvases = append(userCanvases, &canvasCopy)
		}
	}
	
	// 按更新时间降序排序
	sort.Slice(userCanvases, func(i, j int) bool {
		return userCanvases[i].UpdatedAt.After(userCanvases[j].UpdatedAt)
	})
	
	// 分页
	if offset >= len(userCanvases) {
		return []*core.Canvas{}, nil
	}
	
	end := offset + limit
	if end > len(userCanvases) {
		end = len(userCanvases)
	}
	
	return userCanvases[offset:end], nil
}

func (s *canvasStore) UpdateCanvas(ctx context.Context, canvas *core.Canvas) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	existingCanvas, exists := s.canvases[canvas.ID]
	if !exists || existingCanvas.UserID != canvas.UserID {
		return nil // 画布不存在或用户不匹配
	}
	
	canvas.UpdatedAt = time.Now()
	canvas.CreatedAt = existingCanvas.CreatedAt // 保持原有的创建时间
	
	s.canvases[canvas.ID] = canvas
	
	return nil
}

func (s *canvasStore) DeleteCanvas(ctx context.Context, id string, userID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	canvas, exists := s.canvases[id]
	if !exists || canvas.UserID != userID {
		return nil // 画布不存在或用户不匹配
	}
	
	delete(s.canvases, id)
	
	return nil
}