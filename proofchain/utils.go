package proofchain

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// readFile reads a file and returns its contents.
func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// filepath returns the base name of a path.
func filepathBase(path string) string {
	return filepath.Base(path)
}

// jsonMarshal marshals a value to JSON.
func jsonMarshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// seg percent-encodes a caller-supplied identifier so it is always a single
// path segment (no "/", "?", "#" or ".." can escape into the route).
func seg(s string) string {
	return url.PathEscape(s)
}

// isLoopbackHost reports whether host (without port) is a loopback address.
func isLoopbackHost(host string) bool {
	host = strings.Trim(host, "[]")
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// validateBaseURL enforces https:// for API endpoints. http:// is accepted
// only for loopback hosts or when allowInsecure is set. The returned URL has
// any trailing slash removed so paths can be appended directly.
func validateBaseURL(raw string, allowInsecure bool) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("proofchain: invalid base URL %q: %w", raw, err)
	}
	switch u.Scheme {
	case "https":
	case "http":
		if !allowInsecure && !isLoopbackHost(u.Hostname()) {
			return "", fmt.Errorf("proofchain: refusing to send credentials over plaintext http:// to %q; use https:// or opt in with the insecure-HTTP option", u.Host)
		}
	default:
		return "", fmt.Errorf("proofchain: base URL %q must use https:// (got scheme %q)", raw, u.Scheme)
	}
	if u.Host == "" {
		return "", fmt.Errorf("proofchain: base URL %q has no host", raw)
	}
	return strings.TrimRight(raw, "/"), nil
}

// errRedirectNotFollowed is returned when the API answers with a redirect.
// Redirects are never followed: credentials must not be replayed to a
// different origin.
var errRedirectNotFollowed = errors.New("proofchain: redirect not followed (credentials are never forwarded)")

// noRedirect is the http.Client CheckRedirect policy used by every client in
// this package.
func noRedirect(*http.Request, []*http.Request) error {
	return http.ErrUseLastResponse
}

// hardenHTTPClient returns a copy of hc (or a fresh client) that never follows
// redirects.
func hardenHTTPClient(hc *http.Client) *http.Client {
	if hc == nil {
		hc = &http.Client{}
	}
	cp := *hc
	cp.CheckRedirect = noRedirect
	return &cp
}

// redactSecret returns the first 8 characters of s followed by "…".
func redactSecret(s string) string {
	if s == "" {
		return ""
	}
	r := []rune(s)
	if len(r) <= 8 {
		return "…"
	}
	return string(r[:8]) + "…"
}
