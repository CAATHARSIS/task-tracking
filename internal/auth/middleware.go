package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (s *JWTService) JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Пропускаем статические файлы и API
		if strings.HasPrefix(c.Request.URL.Path, "/static") ||
			strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.Next()
			return
		}

		// Публичные маршруты
		publicRoutes := []string{"/", "/login", "/register"}
		for _, route := range publicRoutes {
			if c.Request.URL.Path == route {
				c.Next()
				return
			}
		}

		token, err := c.Cookie("auth_token")
		if err != nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		claims, err := s.ParseToken(token)
		if err != nil {
			c.SetCookie("auth_token", "", -1, "/", "", false, true)
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("IsAuthenticated", true)
		c.Next()
	}
}

func MethodOverride() gin.HandlerFunc {
    return func(c *gin.Context) {
        if c.Request.Method == "POST" {
            method := c.PostForm("_method")
            if method == "PUT" || method == "PATCH" || method == "DELETE" {
                c.Request.Method = method
            }
        }
        c.Next()
    }
}
