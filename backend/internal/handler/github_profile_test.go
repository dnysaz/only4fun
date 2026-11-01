package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

const sampleContributions = `
<h2 class="f4 text-normal mb-2"> 1,234 contributions in the last year </h2>
<table>
  <tbody>
    <tr>
      <td tabindex="0" data-ix="0" aria-describedby="contribution-graph-legend-level-0" data-date="2025-11-02" id="contribution-day-component-0-0" data-level="0"></td>
      <td tabindex="0" data-ix="1" aria-describedby="contribution-graph-legend-level-1" data-date="2026-11-01" id="contribution-day-component-0-1" data-level="2"></td>
    </tr>
  </tbody>
</table>
<tool-tip for="contribution-day-component-0-0" id="tooltip-0-0">No contributions on Sunday, November 2, 2025</tool-tip>
<tool-tip for="contribution-day-component-0-1" id="tooltip-0-1">7 contributions on Sunday, November 1, 2026</tool-tip>
`

func TestParseContributions(t *testing.T) {
	days, total := parseContributions(sampleContributions)

	if len(days) != 2 {
		t.Fatalf("parseContributions menghasilkan %d hari, ingin 2", len(days))
	}
	if days[0].Date != "2025-11-02" || days[0].Level != 0 || days[0].Count != 0 {
		t.Errorf("hari pertama = %+v, ingin 2025-11-02 level 0 count 0", days[0])
	}
	if days[1].Date != "2026-11-01" || days[1].Level != 2 || days[1].Count != 7 {
		t.Errorf("hari kedua = %+v, ingin 2026-11-01 level 2 count 7", days[1])
	}
	// Angka total diambil dari judul, bukan hasil penjumlahan tooltip.
	if total != 1234 {
		t.Errorf("total = %d, ingin 1234", total)
	}
}

func TestParseContributionsWithoutTooltips(t *testing.T) {
	page := `<td data-date="2026-01-01" id="d" data-level="4"></td>`
	days, total := parseContributions(page)

	if len(days) != 1 {
		t.Fatalf("parseContributions menghasilkan %d hari, ingin 1", len(days))
	}
	if days[0].Count != -1 {
		t.Errorf("count = %d, ingin -1 (tidak tersedia)", days[0].Count)
	}
	if total != 0 {
		t.Errorf("total = %d, ingin 0 karena tidak ada angka per hari", total)
	}
}

func TestParseContributionsEmptyPage(t *testing.T) {
	days, total := parseContributions("<html><body>not found</body></html>")
	if len(days) != 0 || total != 0 {
		t.Errorf("halaman tanpa kalender harus menghasilkan 0 hari, dapat %d hari total %d", len(days), total)
	}
}

// TestGetProfileReturnsBioReadmeAndCalendar memastikan endpoint menggabungkan
// data /user, README profil, dan kalender kontribusi dari sumber yang benar.
func TestGetProfileReturnsBioReadmeAndCalendar(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/user":
			if r.Header.Get("Authorization") != "Bearer gh-token" {
				t.Errorf("Authorization = %q, ingin token session", r.Header.Get("Authorization"))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"login":            "octo-test-user",
				"name":             "Octo Test",
				"bio":              "Building things",
				"company":          "@only4fun",
				"location":         "Denpasar",
				"blog":             "https://example.com",
				"twitter_username": "octo",
				"public_repos":     12,
				"followers":        34,
				"following":        5,
				"created_at":       "2020-01-02T03:04:05Z",
			})
		case "/repos/octo-test-user/octo-test-user/readme":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte("<h1>Hello</h1>"))
		case "/users/octo-test-user/contributions":
			_, _ = w.Write([]byte(sampleContributions))
		default:
			http.NotFound(w, r)
		}
	}))
	defer upstream.Close()

	store := newTestStore(t)
	handler, sessionID := newTestGitHubHandler(t, store)
	handler.apiBase = upstream.URL
	handler.webBaseURL = upstream.URL

	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/github/profile", nil)
	c.Request.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
	c.Set("userID", "user-id")
	handler.GetProfile(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("GetProfile = %d, ingin 200: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Profile GitHubProfile `json:"profile"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	p := body.Profile
	if p.Login != "octo-test-user" || p.Bio != "Building things" {
		t.Errorf("profil = %+v, ingin login dan bio terisi", p)
	}
	if p.PublicRepos != 12 || p.Followers != 34 || p.Following != 5 {
		t.Errorf("statistik = %d/%d/%d, ingin 12/34/5", p.PublicRepos, p.Followers, p.Following)
	}
	if p.ReadmeHTML != "<h1>Hello</h1>" {
		t.Errorf("readme = %q, ingin HTML dari GitHub", p.ReadmeHTML)
	}
	if len(p.Contributions) != 2 || p.TotalContributions != 1234 {
		t.Errorf("kontribusi = %d hari total %d, ingin 2 hari total 1234", len(p.Contributions), p.TotalContributions)
	}
}

// TestGetProfileRequiresSession memastikan endpoint menolak request tanpa sesi.
func TestGetProfileRequiresSession(t *testing.T) {
	store := newTestStore(t)
	gin.SetMode(gin.TestMode)

	handler := NewGitHubHandler(nil)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/github/profile", nil)
	handler.GetProfile(c)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("GetProfile tanpa session = %d, ingin 401", rec.Code)
	}
	_ = store
}
