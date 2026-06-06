// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package mcp

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// okHandler is the protected handler stand-in: it records that it ran.
func okHandler(ran *bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*ran = true
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
}

func TestBearerAuthMiddleware(t *testing.T) {
	const token = "s3cr3t-token"

	cases := []struct {
		name       string
		path       string
		authHeader string
		wantStatus int
		wantNext   bool // did the protected handler run?
	}{
		{"valid token reaches handler", "/mcp", "Bearer " + token, http.StatusOK, true},
		{"root path also gated", "/", "Bearer " + token, http.StatusOK, true},
		{"missing header rejected", "/mcp", "", http.StatusUnauthorized, false},
		{"wrong token rejected", "/mcp", "Bearer nope", http.StatusUnauthorized, false},
		{"no Bearer prefix rejected", "/mcp", token, http.StatusUnauthorized, false},
		{"health open without token", "/health", "", http.StatusOK, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var ran bool
			h := bearerAuthMiddleware(token, okHandler(&ran))

			req := httptest.NewRequest(http.MethodPost, tc.path, nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			if ran != tc.wantNext {
				t.Errorf("protected handler ran = %v, want %v", ran, tc.wantNext)
			}
			if tc.wantStatus == http.StatusUnauthorized {
				if got := rec.Header().Get("WWW-Authenticate"); got == "" {
					t.Error("401 response missing WWW-Authenticate challenge")
				}
			}
		})
	}
}
