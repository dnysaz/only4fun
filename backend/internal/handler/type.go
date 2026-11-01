package handler

import "github.com/gin-gonic/gin"

// GetUserID mengambil user ID (UUID v7) dari context gin.
func GetUserID(c *gin.Context) (string, bool) {
	val, exists := c.Get("userID")
	if !exists {
		return "", false
	}
	userID, ok := val.(string)
	if !ok || userID == "" {
		return "", false
	}
	return userID, true
}

// UserResponse adalah data user yang dikembalikan API.
type UserResponse struct {
	ID        string `json:"id"`
	GitHubID  string `json:"github_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}
