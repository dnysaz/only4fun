package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestValidateCustomDomain(t *testing.T) {
	valid := []string{
		"example.com",
		"www.example.com",
		"my-site.example.co.id",
		"a.b.example.dev",
	}
	for _, domain := range valid {
		if err := validateCustomDomain(domain); err != nil {
			t.Errorf("validateCustomDomain(%q) = %v, ingin valid", domain, err)
		}
	}

	invalid := []string{
		"",
		"example",
		"http://example.com",
		"example.com/path",
		"example.com:8080",
		"-bad.example.com",
		"bad-.example.com",
		"*.example.com",
		"example..com",
		"example.c0m",
	}
	for _, domain := range invalid {
		if err := validateCustomDomain(domain); err == nil {
			t.Errorf("validateCustomDomain(%q) = nil, ingin error", domain)
		}
	}
}

func TestUpdateCustomDomain(t *testing.T) {
	store := newTestStore(t)
	h := newTestHandler(t, store)
	userID := store.createUser(t, "owner")

	projectID := store.createProject(t, userID, "https://github.com/demo/site-a", "site-a", nil)
	otherID := store.createProject(t, userID, "https://github.com/demo/site-b", "site-b", nil)

	call := func(id, body string) *httptest.ResponseRecorder {
		t.Helper()
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Params = gin.Params{{Key: "id", Value: id}}
		c.Request = httptest.NewRequest(http.MethodPut, "/api/projects/"+id+"/domain", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("userID", userID)
		h.UpdateCustomDomain(c)
		return rec
	}

	if rec := call(projectID, `{"domain":"https://example.com"}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid domain status = %d, want %d; body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}

	rec := call(projectID, `{"domain":"WWW.Example.COM."}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("save status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"custom_domain":"www.example.com"`) {
		t.Fatalf("save response = %s, want normalized domain", rec.Body.String())
	}
	project, err := store.projects.GetByID(projectID)
	if err != nil || project.CustomDomain != "www.example.com" {
		t.Fatalf("stored domain = %q, err = %v", project.CustomDomain, err)
	}

	if rec := call(otherID, `{"domain":"www.example.com"}`); rec.Code != http.StatusConflict {
		t.Fatalf("duplicate status = %d, want %d; body = %s", rec.Code, http.StatusConflict, rec.Body.String())
	}

	if rec := call(projectID, `{"domain":""}`); rec.Code != http.StatusOK {
		t.Fatalf("remove status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	project, err = store.projects.GetByID(projectID)
	if err != nil || project.CustomDomain != "" {
		t.Fatalf("domain after removal = %q, err = %v", project.CustomDomain, err)
	}

	if rec := call(otherID, `{"domain":"www.example.com"}`); rec.Code != http.StatusOK {
		t.Fatalf("reuse after removal status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}
