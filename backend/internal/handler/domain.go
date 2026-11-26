package handler

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"only4.fun/backend/internal/dns"

	"github.com/gin-gonic/gin"
)

var domainLabelPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

// validateCustomDomain memeriksa format domain yang diarahkan ke project.
func validateCustomDomain(domain string) error {
	if domain == "" {
		return fmt.Errorf("domain is required")
	}
	if len(domain) > 253 {
		return fmt.Errorf("domain is too long")
	}
	lower := strings.ToLower(domain)
	if strings.Contains(lower, "/") || strings.Contains(lower, ":") || strings.Contains(lower, " ") {
		return fmt.Errorf("invalid domain: enter a hostname only, without http:// or a path")
	}
	if strings.HasPrefix(lower, "*.") || strings.HasPrefix(lower, ".") {
		return fmt.Errorf("invalid domain: wildcards are not supported")
	}
	labels := strings.Split(lower, ".")
	if len(labels) < 2 {
		return fmt.Errorf("domain must have an extension, for example example.com")
	}
	for _, label := range labels {
		if !domainLabelPattern.MatchString(label) {
			return fmt.Errorf("invalid domain: %q is not a valid label", label)
		}
	}
	tld := labels[len(labels)-1]
	if len(tld) < 2 {
		return fmt.Errorf("invalid domain: the extension is too short")
	}
	for _, ch := range tld {
		if ch < 'a' || ch > 'z' {
			return fmt.Errorf("invalid domain: the extension must contain letters only")
		}
	}
	return nil
}

// UpdateCustomDomain menyimpan atau menghapus domain kustom project.
// Domain hanya bisa dipakai satu project pada satu waktu.
func (h *ProjectHandler) UpdateCustomDomain(c *gin.Context) {
	project, ok := h.ownedProject(c)
	if !ok {
		return
	}

	var req struct {
		Domain string `json:"domain"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	domain := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(req.Domain)), ".")
	// Menghapus domain.
	if domain == "" {
		if err := h.projectStore.UpdateCustomDomain(project.ID, ""); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove the domain"})
			return
		}
		if h.dns != nil && project.CustomDomain != "" {
			if err := h.dns.DeleteCustomHostname(c.Request.Context(), project.CustomDomain); err != nil {
				log.Printf("failed to remove custom hostname %s: %v", project.CustomDomain, err)
			}
		}
		c.JSON(http.StatusOK, gin.H{"custom_domain": ""})
		return
	}

	if err := validateCustomDomain(domain); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Domain platform sendiri harus memakai project name, bukan domain kustom.
	if h.dns != nil {
		base := strings.ToLower(h.dns.BaseDomain())
		if base != "" && (domain == base || strings.HasSuffix(domain, "."+base)) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("use the project name for %s subdomains instead of a custom domain", base)})
			return
		}
	}
	if existing, err := h.projectStore.GetByCustomDomain(domain); err == nil && existing != nil && existing.ID != project.ID {
		c.JSON(http.StatusConflict, gin.H{"error": "this domain is already used by another project"})
		return
	}

	if err := h.projectStore.UpdateCustomDomain(project.ID, domain); err != nil {
		if strings.Contains(strings.ToUpper(err.Error()), "UNIQUE") {
			c.JSON(http.StatusConflict, gin.H{"error": "this domain is already used by another project"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save the domain"})
		return
	}

	// Domain lama dilepas dari Cloudflare supaya tidak tertinggal.
	if h.dns != nil && project.CustomDomain != "" && project.CustomDomain != domain {
		if err := h.dns.DeleteCustomHostname(c.Request.Context(), project.CustomDomain); err != nil {
			log.Printf("failed to remove the old custom hostname %s: %v", project.CustomDomain, err)
		}
	}

	warning := ""
	if h.dns != nil && h.dns.Enabled() {
		// Best effort: butuh Cloudflare for SaaS. Bila gagal, domain tetap
		// tersimpan dan SSL bisa diatur user di zona Cloudflare miliknya.
		if err := h.dns.EnsureCustomHostname(c.Request.Context(), domain); err != nil {
			log.Printf("failed to register custom hostname %s: %v", domain, err)
			warning = "Domain saved, but automatic Cloudflare certificates are not active yet: " + err.Error()
		}
	}

	c.JSON(http.StatusOK, gin.H{"custom_domain": domain, "warning": warning})
}

// mapDomainStatus menerjemahkan status Cloudflare ke label UI yang stabil.
func mapDomainStatus(status *dns.CustomHostnameStatus) (string, string) {
	switch {
	case status.Status == "active" && status.SSLStatus == "active":
		return "active", "Domain is active and the SSL certificate is issued."
	case status.Status == "active":
		return "ssl_pending", "DNS is verified. The SSL certificate is being issued."
	case status.Status == "pending":
		return "dns_pending", "Waiting for DNS. Point a CNAME record to the target shown below."
	default:
		return "pending", "Waiting for DNS verification."
	}
}

// GetDomainStatus melaporkan status verifikasi domain kustom supaya frontend
// bisa menampilkan indikator seperti "Verifying" / "Active".
func (h *ProjectHandler) GetDomainStatus(c *gin.Context) {
	project, ok := h.ownedProject(c)
	if !ok {
		return
	}

	hostname := project.CustomDomain
	if hostname == "" {
		c.JSON(http.StatusOK, gin.H{"state": "none", "message": "No custom domain set."})
		return
	}
	if h.dns == nil || !h.dns.Enabled() {
		c.JSON(http.StatusOK, gin.H{
			"state": "ssl_unavailable", "hostname": hostname,
			"message": "Automatic SSL is not configured on this server.",
		})
		return
	}

	status, err := h.dns.GetCustomHostnameStatus(c.Request.Context(), hostname)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"state": "unknown", "hostname": hostname,
			"message": "Could not check the domain status: " + err.Error(),
		})
		return
	}
	if status == nil {
		c.JSON(http.StatusOK, gin.H{
			"state": "pending", "hostname": hostname,
			"message": "Domain saved. Waiting for Cloudflare to register it.",
		})
		return
	}

	state, message := mapDomainStatus(status)

	c.JSON(http.StatusOK, gin.H{
		"state": state, "hostname": hostname, "message": message,
		"hostname_status": status.Status, "ssl_status": status.SSLStatus,
	})
}
