package auth

import (
	"context"
	"excalidraw-complete/core"
	"excalidraw-complete/stores"
	"net/http"

	"github.com/sirupsen/logrus"
)

type contextKey string

const (
	UserContextKey contextKey = "user"
)

// AuthMiddleware 认证中间件
func AuthMiddleware(store stores.Store) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 获取会话token
			sessionCookie, err := r.Cookie("session_token")
			if err != nil {
				// 没有会话，继续处理请求但不设置用户
				next.ServeHTTP(w, r)
				return
			}
			
			// 验证会话
			session, err := store.GetSessionStore().GetSession(r.Context(), sessionCookie.Value)
			if err != nil {
				logrus.WithError(err).Error("Failed to get session")
				next.ServeHTTP(w, r)
				return
			}
			
			if session == nil {
				// 会话不存在或已过期
				next.ServeHTTP(w, r)
				return
			}
			
			// 获取用户信息
			user, err := store.GetUserStore().GetUserByID(r.Context(), session.UserID)
			if err != nil {
				logrus.WithError(err).Error("Failed to get user")
				next.ServeHTTP(w, r)
				return
			}
			
			if user == nil {
				// 用户不存在
				next.ServeHTTP(w, r)
				return
			}
			
			// 将用户信息添加到上下文
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAuth 需要认证的中间件
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// GetUserFromContext 从上下文获取用户信息
func GetUserFromContext(ctx context.Context) *core.User {
	user, ok := ctx.Value(UserContextKey).(*core.User)
	if !ok {
		return nil
	}
	return user
}

// IsAuthenticated 检查用户是否已认证
func IsAuthenticated(ctx context.Context) bool {
	return GetUserFromContext(ctx) != nil
}