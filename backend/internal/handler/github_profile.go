package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ContributionDay adalah satu kotak pada kalender kontribusi GitHub.
type ContributionDay struct {
	Date  string `json:"date"`
	Level int    `json:"level"`
	// Count bernilai -1 bila GitHub tidak menyertakan angka pastinya.
	Count int `json:"count"`
}

// GitHubProfile adalah data profil publik yang ditampilkan di halaman Profile.
type GitHubProfile struct {
	Login              string            `json:"login"`
	Name               string            `json:"name"`
	Bio                string            `json:"bio"`
	Company            string            `json:"company"`
	Location           string            `json:"location"`
	Blog               string            `json:"blog"`
	Twitter            string            `json:"twitter"`
	PublicRepos        int               `json:"public_repos"`
	Followers          int               `json:"followers"`
	Following          int               `json:"following"`
	CreatedAt          string            `json:"created_at"`
	ReadmeHTML         string            `json:"readme_html"`
	Contributions      []ContributionDay `json:"contributions"`
	TotalContributions int               `json:"total_contributions"`
}

const (
	profileCacheTTL      = 10 * time.Minute
	defaultGitHubWebBase = "https://github.com"
)

type profileCacheEntry struct {
	profile *GitHubProfile
	stored  time.Time
}

var profileCache = struct {
	sync.Mutex
	m map[string]profileCacheEntry
}{m: map[string]profileCacheEntry{}}

func (h *GitHubHandler) webBase() string {
	if h.webBaseURL != "" {
		return h.webBaseURL
	}
	return defaultGitHubWebBase
}

// githubGet menjalankan request ke GitHub dengan token session dan memakai
// context request asal supaya ikut batal saat klien memutus koneksi.
func (h *GitHubHandler) githubGet(c *gin.Context, url, accept, token string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if accept != "" {
		req.Header.Set("Accept", accept)
	}
	return (&http.Client{Timeout: 15 * time.Second}).Do(req)
}

// GetProfile mengembalikan bio, statistik, README profil, dan kalender
// Contributions. Data is briefly cached so GitHub rate limit is not quickly exhausted.
func (h *GitHubHandler) GetProfile(c *gin.Context) {
	session, ok := h.sessionFromRequest(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}

	base := h.apiBase
	if base == "" {
		base = defaultGitHubAPIBase
	}

	// Ambil identitas + bio dari token yang tersimpan di session.
	resp, err := h.githubGet(c, base+"/user", "application/vnd.github+json", session.Token)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to reach GitHub"})
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusBadGateway, gin.H{"error": "GitHub rejected the profile request"})
		return
	}

	var raw struct {
		Login       string `json:"login"`
		Name        string `json:"name"`
		Bio         string `json:"bio"`
		Company     string `json:"company"`
		Location    string `json:"location"`
		Blog        string `json:"blog"`
		Twitter     string `json:"twitter_username"`
		PublicRepos int    `json:"public_repos"`
		Followers   int    `json:"followers"`
		Following   int    `json:"following"`
		CreatedAt   string `json:"created_at"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to read the GitHub profile"})
		return
	}
	if raw.Login == "" {
		c.JSON(http.StatusBadGateway, gin.H{"error": "GitHub did not return a username"})
		return
	}

	profileCache.Lock()
	if entry, found := profileCache.m[raw.Login]; found && time.Since(entry.stored) < profileCacheTTL {
		cached := *entry.profile
		profileCache.Unlock()
		c.JSON(http.StatusOK, gin.H{"profile": cached})
		return
	}
	profileCache.Unlock()

	profile := &GitHubProfile{
		Login:       raw.Login,
		Name:        raw.Name,
		Bio:         raw.Bio,
		Company:     raw.Company,
		Location:    raw.Location,
		Blog:        raw.Blog,
		Twitter:     raw.Twitter,
		PublicRepos: raw.PublicRepos,
		Followers:   raw.Followers,
		Following:   raw.Following,
		CreatedAt:   raw.CreatedAt,
	}

	// README profil: repo bernama sama dengan username. 404 berarti tidak ada.
	readmeURL := base + "/repos/" + raw.Login + "/" + raw.Login + "/readme"
	if readmeResp, err := h.githubGet(c, readmeURL, "application/vnd.github.html", session.Token); err == nil {
		body, _ := io.ReadAll(io.LimitReader(readmeResp.Body, 512*1024))
		readmeResp.Body.Close()
		if readmeResp.StatusCode == http.StatusOK {
			profile.ReadmeHTML = string(body)
		}
	}

	// Kalender kontribusi hanya tersedia lewat halaman publik GitHub.
	contribURL := h.webBase() + "/users/" + raw.Login + "/contributions"
	if contribResp, err := h.githubGet(c, contribURL, "text/html", ""); err == nil {
		body, _ := io.ReadAll(io.LimitReader(contribResp.Body, 2*1024*1024))
		contribResp.Body.Close()
		if contribResp.StatusCode == http.StatusOK {
			profile.Contributions, profile.TotalContributions = parseContributions(string(body))
		}
	}
	if profile.Contributions == nil {
		profile.Contributions = []ContributionDay{}
	}

	profileCache.Lock()
	profileCache.m[raw.Login] = profileCacheEntry{profile: profile, stored: time.Now()}
	// Buang entri lama supaya map tidak tumbuh tanpa batas.
	for login, entry := range profileCache.m {
		if time.Since(entry.stored) > profileCacheTTL {
			delete(profileCache.m, login)
		}
	}
	profileCache.Unlock()

	c.JSON(http.StatusOK, gin.H{"profile": *profile})
}

var (
	// <td ... data-date="2026-11-01" ... data-level="2" ...>
	contribCellRe = regexp.MustCompile(`<td[^>]*data-date="[^"]*"[^>]*>`)
	// <tool-tip for="contribution-day-component-0-0" ...>3 contributions on ...</tool-tip>
	contribTipRe = regexp.MustCompile(`(?s)<tool-tip[^>]*for="([^"]+)"[^>]*>(.*?)</tool-tip>`)
	attrRe       = regexp.MustCompile(`(\w[\w-]*)="([^"]*)"`)
	totalRe      = regexp.MustCompile(`(?i)([\d,]+)\s+contributions? in the last year`)
	leadingNumRe = regexp.MustCompile(`(?i)^\s*(no|([\d,]+))\s+contributions?`)
)

// parseContributions membaca kalender kontribusi dari HTML halaman profil.
// Angka per hari diambil dari tooltip bila markup-nya tersedia; bila tidak,
// Count dibiarkan -1 dan UI hanya memakai Level.
func parseContributions(page string) ([]ContributionDay, int) {
	tips := map[string]int{}
	for _, match := range contribTipRe.FindAllStringSubmatch(page, -1) {
		if num := leadingNumRe.FindStringSubmatch(match[2]); num != nil {
			count := 0
			if strings.EqualFold(strings.TrimSpace(num[1]), "no") {
				count = 0
			} else {
				cleaned := strings.ReplaceAll(num[2], ",", "")
				if parsed, err := strconv.Atoi(cleaned); err == nil {
					count = parsed
				} else {
					continue
				}
			}
			tips[match[1]] = count
		}
	}

	days := make([]ContributionDay, 0, 371)
	sum := 0
	for _, cell := range contribCellRe.FindAllString(page, -1) {
		attrs := map[string]string{}
		for _, a := range attrRe.FindAllStringSubmatch(cell, -1) {
			attrs[a[1]] = a[2]
		}
		date := attrs["data-date"]
		if date == "" {
			continue
		}
		level, _ := strconv.Atoi(attrs["data-level"])
		count := -1
		if value, ok := tips[attrs["id"]]; ok {
			count = value
			sum += value
		}
		days = append(days, ContributionDay{Date: date, Level: level, Count: count})
	}

	total := sum
	if match := totalRe.FindStringSubmatch(page); match != nil {
		if parsed, err := strconv.Atoi(strings.ReplaceAll(match[1], ",", "")); err == nil {
			total = parsed
		}
	}
	return days, total
}
