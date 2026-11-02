package handler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"only4.fun/backend/internal/database"

	"github.com/gin-gonic/gin"
)

type GitHubHandler struct {
	sessionStore *database.SessionStore
	// apiBase menunjuk ke API GitHub. Bisa diganti server palsu saat test
	// sehingga pagination teruji tanpa memanggil GitHub sungguhan.
	apiBase string
	// webBaseURL adalah situs GitHub biasa, dipakai untuk halaman kontribusi
	// yang tidak tersedia lewat API.
	webBaseURL string
}

const defaultGitHubAPIBase = "https://api.github.com"

func NewGitHubHandler(sessionStore *database.SessionStore) *GitHubHandler {
	return &GitHubHandler{
		sessionStore: sessionStore,
		apiBase:      defaultGitHubAPIBase,
		webBaseURL:   defaultGitHubWebBase,
	}
}

// GetRepoContents mengambil daftar file/folder di root repo dari GitHub API.
func (h *GitHubHandler) GetRepoContents(c *gin.Context) {
	session, ok := h.sessionFromRequest(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}

	owner := c.Param("owner")
	repo := c.Param("repo")
	if owner == "" || repo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "owner and repo are required"})
		return
	}

	base := h.apiBase
	if base == "" {
		base = defaultGitHubAPIBase
	}
	apiURL := fmt.Sprintf("%s/repos/%s/%s/contents", base, owner, repo)

	req, err := http.NewRequestWithContext(c.Request.Context(), "GET", apiURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create the GitHub request"})
		return
	}
	req.Header.Set("Authorization", "Bearer "+session.Token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to reach GitHub"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("GitHub rejected the request (status %d)", resp.StatusCode)})
		return
	}

	var items []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to read repository contents"})
		return
	}

	type ContentItem struct {
		Name string `json:"name"`
		Type string `json:"type"`
		Size int    `json:"size"`
		Path string `json:"path"`
	}
	result := make([]ContentItem, 0, len(items))
	for _, item := range items {
		result = append(result, ContentItem{
			Name: fmt.Sprintf("%v", item["name"]),
			Type: fmt.Sprintf("%v", item["type"]),
			Size: int(getFloat(item["size"])),
			Path: fmt.Sprintf("%v", item["path"]),
		})
	}

	c.JSON(http.StatusOK, gin.H{"contents": result})
}

// GetRepoReadme mengambil isi README repo dari GitHub API.
func (h *GitHubHandler) GetRepoReadme(c *gin.Context) {
	session, ok := h.sessionFromRequest(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}

	owner := c.Param("owner")
	repo := c.Param("repo")
	if owner == "" || repo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "owner and repo are required"})
		return
	}

	base := h.apiBase
	if base == "" {
		base = defaultGitHubAPIBase
	}
	apiURL := fmt.Sprintf("%s/repos/%s/%s/readme", base, owner, repo)

	req, err := http.NewRequestWithContext(c.Request.Context(), "GET", apiURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create the GitHub request"})
		return
	}
	req.Header.Set("Authorization", "Bearer "+session.Token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to reach GitHub"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		c.JSON(http.StatusOK, gin.H{"readme": "", "found": false})
		return
	}
	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("GitHub rejected the request (status %d)", resp.StatusCode)})
		return
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to read readme data"})
		return
	}

	content, _ := data["content"].(string)
	encoding, _ := data["encoding"].(string)
	if encoding == "base64" && content != "" {
		// Remove newlines from base64 content
		content = strings.ReplaceAll(content, "\n", "")
		decoded, err := base64.StdEncoding.DecodeString(content)
		if err == nil {
			content = string(decoded)
		}
	}

	c.JSON(http.StatusOK, gin.H{"readme": content, "found": true})
}
func getFloat(v interface{}) float64 {
	if f, ok := v.(float64); ok {
		return f
	}
	return 0
}
func (h *GitHubHandler) sessionFromRequest(c *gin.Context) (*database.Session, bool) {
	sessionID, _ := c.Cookie("session_id")
	if sessionID == "" {
		if auth := c.GetHeader("Authorization"); strings.HasPrefix(auth, "Bearer ") {
			sessionID = strings.TrimPrefix(auth, "Bearer ")
		}
	}
	if sessionID == "" {
		return nil, false
	}
	session, err := h.sessionStore.GetBySessionID(sessionID)
	if err != nil {
		return nil, false
	}
	return session, true
}

func (h *GitHubHandler) ListRepos(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	_ = userID

	session, ok := h.sessionFromRequest(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}

	// Pagination: GitHub mengembalikan maksimal 100 repo per halaman, jadi
	// halaman berikutnya diambil lewat ?page=N (default 1).
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "100"))
	if perPage < 1 || perPage > 100 {
		perPage = 100
	}

	// Timeout eksplisit: http.DefaultClient tidak punya timeout, sehingga
	// request GitHub yang menggantung akan menahan handler tanpa batas.
	base := h.apiBase
	if base == "" {
		base = defaultGitHubAPIBase
	}
	apiURL := fmt.Sprintf("%s/user/repos?per_page=%d&page=%d&sort=updated", base, perPage, page)
	req, err := http.NewRequestWithContext(c.Request.Context(), "GET", apiURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create the GitHub request"})
		return
	}
	req.Header.Set("Authorization", "Bearer "+session.Token)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to reach GitHub"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("GitHub rejected the request (status %d). Please sign in again", resp.StatusCode)})
		return
	}

	var repos []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to read repository data from GitHub"})
		return
	}

	type RepoInfo struct {
		Name     string `json:"name"`
		FullName string `json:"full_name"`
		HTMLURL  string `json:"html_url"`
		Private  bool   `json:"private"`
	}
	result := make([]RepoInfo, 0, len(repos))
	for _, r := range repos {
		private, _ := r["private"].(bool)
		result = append(result, RepoInfo{
			Name:     fmt.Sprintf("%v", r["name"]),
			FullName: fmt.Sprintf("%v", r["full_name"]),
			HTMLURL:  fmt.Sprintf("%v", r["html_url"]),
			Private:  private,
		})
	}

	// Halaman penuh berarti kemungkinan masih ada halaman berikutnya.
	c.JSON(http.StatusOK, gin.H{
		"repos":    result,
		"page":     page,
		"has_more": len(result) == perPage,
	})
}
