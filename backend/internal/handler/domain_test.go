package handler

import "testing"

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
