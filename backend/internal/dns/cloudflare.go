package dns

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"only4.fun/backend/internal/config"
)

const apiBase = "https://api.cloudflare.com/client/v4"

// Client mengelola record DNS di zona Cloudflare.
type Client struct {
	token   string
	zoneID  string
	domain  string
	target  string
	proxied bool
	http    *http.Client
}

type cfResponse struct {
	Success bool `json:"success"`
	Errors  []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
	Result json.RawMessage `json:"result"`
}

type cfRecord struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	Proxied bool   `json:"proxied"`
}

// New membuat client. Bila kredensial/target belum diisi, client dianggap nonaktif
// sehingga pengembangan lokal tetap berjalan tanpa Cloudflare.
func New(cfg config.CloudflareConfig) *Client {
	return &Client{
		token:   cfg.APIKey,
		zoneID:  cfg.ZoneID,
		domain:  strings.Trim(strings.ToLower(cfg.Domain), "."),
		target:  strings.TrimSpace(cfg.Target),
		proxied: cfg.Proxy,
		http:    &http.Client{Timeout: 15 * time.Second},
	}
}

// Enabled melaporkan apakah integrasi Cloudflare siap dipakai.
func (c *Client) Enabled() bool {
	return c != nil && c.token != "" && c.zoneID != "" && c.target != "" && c.domain != ""
}

func (c *Client) fqdn(subdomain string) string {
	return subdomain + "." + c.domain
}

func (c *Client) recordType() string {
	if net.ParseIP(c.target) != nil {
		return "A"
	}
	return "CNAME"
}

// BaseDomain mengembalikan domain utama (mis. only4.fun).
func (c *Client) BaseDomain() string {
	if c == nil {
		return ""
	}
	return c.domain
}

// EnsureRecord memastikan record <subdomain>.<domain> menunjuk ke server ini.
func (c *Client) EnsureRecord(ctx context.Context, subdomain string) error {
	if !c.Enabled() {
		return nil
	}
	name := c.fqdn(subdomain)
	wantType := c.recordType()

	existing, err := c.findRecord(ctx, name)
	if err != nil {
		return err
	}

	payload := map[string]any{
		"type":    wantType,
		"name":    name,
		"content": c.target,
		"proxied": c.proxied,
		"ttl":     1, // 1 = automatic
	}

	if existing == nil {
		return c.do(ctx, http.MethodPost, "/zones/"+c.zoneID+"/dns_records", payload, nil)
	}
	if existing.Type == wantType && existing.Content == c.target && existing.Proxied == c.proxied {
		return nil
	}
	return c.do(ctx, http.MethodPut, "/zones/"+c.zoneID+"/dns_records/"+existing.ID, payload, nil)
}

// DeleteRecord menghapus record subdomain bila ada.
func (c *Client) DeleteRecord(ctx context.Context, subdomain string) error {
	if !c.Enabled() {
		return nil
	}
	existing, err := c.findRecord(ctx, c.fqdn(subdomain))
	if err != nil {
		return err
	}
	if existing == nil {
		return nil
	}
	return c.do(ctx, http.MethodDelete, "/zones/"+c.zoneID+"/dns_records/"+existing.ID, nil, nil)
}

func (c *Client) findRecord(ctx context.Context, name string) (*cfRecord, error) {
	var result []cfRecord
	path := "/zones/" + c.zoneID + "/dns_records?type=" + c.recordType() + "&name=" + name
	if err := c.do(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, nil
	}
	return &result[0], nil
}

func (c *Client) do(ctx context.Context, method, path string, body any, out any) error {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return fmt.Errorf("failed to encode the Cloudflare request: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, apiBase+path, &buf)
	if err != nil {
		return fmt.Errorf("failed to create the Cloudflare request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("failed to reach Cloudflare: %w", err)
	}
	defer resp.Body.Close()

	var parsed cfResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return fmt.Errorf("unreadable Cloudflare response (status %d)", resp.StatusCode)
	}
	if !parsed.Success {
		msg := "unknown error"
		if len(parsed.Errors) > 0 {
			msg = parsed.Errors[0].Message
		}
		return fmt.Errorf("Cloudflare rejected the request: %s", msg)
	}
	if out != nil && len(parsed.Result) > 0 {
		if err := json.Unmarshal(parsed.Result, out); err != nil {
			return fmt.Errorf("unreadable Cloudflare response: %w", err)
		}
	}
	return nil
}

type cfCustomHostname struct {
	ID       string `json:"id"`
	Hostname string `json:"hostname"`
}

// EnsureCustomHostname mendaftarkan domain milik user sebagai custom hostname
// di zona Cloudflare (fitur Cloudflare for SaaS) supaya edge certificate
// diterbitkan otomatis. Bila fitur tidak aktif, error dikembalikan dan caller
// bisa tetap menyimpan domain (SSL diatur di sisi user).
func (c *Client) EnsureCustomHostname(ctx context.Context, hostname string) error {
	if !c.Enabled() {
		return nil
	}
	existing, err := c.findCustomHostname(ctx, hostname)
	if err != nil {
		return err
	}
	if existing != nil {
		return nil
	}
	payload := map[string]any{
		"hostname": hostname,
		"ssl": map[string]any{
			"method": "http",
			"type":   "dv",
		},
	}
	return c.do(ctx, http.MethodPost, "/zones/"+c.zoneID+"/custom_hostnames", payload, nil)
}

// DeleteCustomHostname melepas domain kustom dari zona Cloudflare.
func (c *Client) DeleteCustomHostname(ctx context.Context, hostname string) error {
	if !c.Enabled() {
		return nil
	}
	existing, err := c.findCustomHostname(ctx, hostname)
	if err != nil {
		return err
	}
	if existing == nil {
		return nil
	}
	return c.do(ctx, http.MethodDelete, "/zones/"+c.zoneID+"/custom_hostnames/"+existing.ID, nil, nil)
}

func (c *Client) findCustomHostname(ctx context.Context, hostname string) (*cfCustomHostname, error) {
	var result []cfCustomHostname
	path := "/zones/" + c.zoneID + "/custom_hostnames?hostname=" + url.QueryEscape(hostname)
	if err := c.do(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, nil
	}
	return &result[0], nil
}
