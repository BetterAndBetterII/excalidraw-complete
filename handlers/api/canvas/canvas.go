package canvas

import (
	"encoding/json"
	"excalidraw-complete/core"
	"excalidraw-complete/handlers/auth"
	"excalidraw-complete/stores"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/sirupsen/logrus"
)

type CanvasHandler struct {
	store stores.Store
}

type CreateCanvasRequest struct {
	Title    string          `json:"title"`
	Data     json.RawMessage `json:"data"`
	IsPublic bool            `json:"is_public"`
}

type UpdateCanvasRequest struct {
	Title    string          `json:"title"`
	Data     json.RawMessage `json:"data"`
	IsPublic bool            `json:"is_public"`
}

type CanvasResponse struct {
	*core.Canvas
	DataJSON json.RawMessage `json:"data"`
}

type CanvasListResponse struct {
	Canvases []*CanvasResponse `json:"canvases"`
	Total    int               `json:"total"`
}

func NewCanvasHandler(store stores.Store) *CanvasHandler {
	return &CanvasHandler{
		store: store,
	}
}

func (h *CanvasHandler) Routes() chi.Router {
	r := chi.NewRouter()
	
	// 需要认证的路由
	r.Group(func(r chi.Router) {
		r.Use(auth.RequireAuth)
		
		r.Post("/", h.handleCreateCanvas)
		r.Get("/", h.handleListCanvases)
		r.Get("/{id}", h.handleGetCanvas)
		r.Put("/{id}", h.handleUpdateCanvas)
		r.Delete("/{id}", h.handleDeleteCanvas)
	})
	
	return r
}

func (h *CanvasHandler) handleCreateCanvas(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	var req CreateCanvasRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	if req.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}
	
	canvas := &core.Canvas{
		UserID:   user.ID,
		Title:    req.Title,
		Data:     []byte(req.Data),
		IsPublic: req.IsPublic,
	}
	
	err := h.store.GetCanvasStore().CreateCanvas(r.Context(), canvas)
	if err != nil {
		logrus.WithError(err).Error("Failed to create canvas")
		http.Error(w, "Failed to create canvas", http.StatusInternalServerError)
		return
	}
	
	response := &CanvasResponse{
		Canvas:   canvas,
		DataJSON: json.RawMessage(canvas.Data),
	}
	
	render.JSON(w, r, response)
}

func (h *CanvasHandler) handleListCanvases(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// 分页参数
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	
	limit := 20 // 默认值
	offset := 0 // 默认值
	
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}
	
	canvases, err := h.store.GetCanvasStore().GetCanvasByUser(r.Context(), user.ID, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to get canvases")
		http.Error(w, "Failed to get canvases", http.StatusInternalServerError)
		return
	}
	
	// 转换为响应格式
	var responses []*CanvasResponse
	for _, canvas := range canvases {
		response := &CanvasResponse{
			Canvas:   canvas,
			DataJSON: json.RawMessage(canvas.Data),
		}
		responses = append(responses, response)
	}
	
	result := &CanvasListResponse{
		Canvases: responses,
		Total:    len(responses),
	}
	
	render.JSON(w, r, result)
}

func (h *CanvasHandler) handleGetCanvas(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Canvas ID is required", http.StatusBadRequest)
		return
	}
	
	canvas, err := h.store.GetCanvasStore().GetCanvasById(r.Context(), id)
	if err != nil {
		logrus.WithError(err).Error("Failed to get canvas")
		http.Error(w, "Failed to get canvas", http.StatusInternalServerError)
		return
	}
	
	if canvas == nil {
		http.Error(w, "Canvas not found", http.StatusNotFound)
		return
	}
	
	// 检查权限：只有画布所有者或公开画布可以访问
	if canvas.UserID != user.ID && !canvas.IsPublic {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	
	response := &CanvasResponse{
		Canvas:   canvas,
		DataJSON: json.RawMessage(canvas.Data),
	}
	
	render.JSON(w, r, response)
}

func (h *CanvasHandler) handleUpdateCanvas(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Canvas ID is required", http.StatusBadRequest)
		return
	}
	
	var req UpdateCanvasRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	if req.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}
	
	// 检查画布是否存在且属于当前用户
	existingCanvas, err := h.store.GetCanvasStore().GetCanvasById(r.Context(), id)
	if err != nil {
		logrus.WithError(err).Error("Failed to get canvas")
		http.Error(w, "Failed to get canvas", http.StatusInternalServerError)
		return
	}
	
	if existingCanvas == nil {
		http.Error(w, "Canvas not found", http.StatusNotFound)
		return
	}
	
	if existingCanvas.UserID != user.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	
	// 更新画布
	canvas := &core.Canvas{
		ID:       id,
		UserID:   user.ID,
		Title:    req.Title,
		Data:     []byte(req.Data),
		IsPublic: req.IsPublic,
	}
	
	err = h.store.GetCanvasStore().UpdateCanvas(r.Context(), canvas)
	if err != nil {
		logrus.WithError(err).Error("Failed to update canvas")
		http.Error(w, "Failed to update canvas", http.StatusInternalServerError)
		return
	}
	
	response := &CanvasResponse{
		Canvas:   canvas,
		DataJSON: json.RawMessage(canvas.Data),
	}
	
	render.JSON(w, r, response)
}

func (h *CanvasHandler) handleDeleteCanvas(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Canvas ID is required", http.StatusBadRequest)
		return
	}
	
	err := h.store.GetCanvasStore().DeleteCanvas(r.Context(), id, user.ID)
	if err != nil {
		logrus.WithError(err).Error("Failed to delete canvas")
		http.Error(w, "Failed to delete canvas", http.StatusInternalServerError)
		return
	}
	
	render.JSON(w, r, map[string]string{"message": "Canvas deleted successfully"})
}