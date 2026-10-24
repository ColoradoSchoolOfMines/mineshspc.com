package internal

import (
	"net/http"
	"strings"
	"testing"
)

func TestStaticCache(t *testing.T) {
	handler := staticHandler()
	versioned := doRequest(handler, http.MethodGet, staticURL("index.css"))
	if versioned.Code != http.StatusOK || versioned.Header().Get("Cache-Control") != "public, max-age=31536000, immutable" {
		t.Fatalf("versioned asset: status %d, cache %q", versioned.Code, versioned.Header().Get("Cache-Control"))
	}
	if !strings.Contains(versioned.Body.String(), "./fonts/subset-GothamBold.woff2") {
		t.Fatal("CSS font URL changed unexpectedly")
	}
	font := doRequest(handler, http.MethodGet, staticURL("fonts/subset-GothamBold.woff2"))
	if font.Code != http.StatusOK || font.Header().Get("Cache-Control") != "public, max-age=31536000, immutable" {
		t.Fatalf("versioned font: status %d, cache %q", font.Code, font.Header().Get("Cache-Control"))
	}
	legacy := doRequest(handler, http.MethodGet, "/static/index.css")
	if legacy.Code != http.StatusOK || legacy.Header().Get("Cache-Control") != "public, max-age=3600" {
		t.Fatalf("legacy asset: status %d, cache %q", legacy.Code, legacy.Header().Get("Cache-Control"))
	}
	missing := doRequest(handler, http.MethodGet, "/static/vwrong/index.css")
	if missing.Code != http.StatusNotFound || missing.Header().Get("Cache-Control") != "" {
		t.Fatalf("wrong version: status %d, cache %q", missing.Code, missing.Header().Get("Cache-Control"))
	}
}
