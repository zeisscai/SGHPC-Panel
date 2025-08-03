package api

import (
	"net/http"
	"strings"
)

// AuthMiddleware 是一个认证中间件，用于保护需要认证的API端点
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 获取Authorization头
		authHeader := r.Header.Get("Authorization")
		
		// 检查是否存在Authorization头
		if authHeader == "" {
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}
		
		// 检查Bearer token格式
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}
		
		// 提取token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		
		// 验证token（这里使用简单的验证，实际项目中应该使用JWT或其他安全机制）
		if !isValidToken(token) {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}
		
		// Token有效，继续处理请求
		next.ServeHTTP(w, r)
	}
}

// isValidToken 验证token是否有效
func isValidToken(token string) bool {
	// 这里应该实现实际的token验证逻辑
	// 例如验证JWT token或检查token是否在有效期内
	// 为简化起见，我们只检查token是否以"token_"开头
	return strings.HasPrefix(token, "token_")
}