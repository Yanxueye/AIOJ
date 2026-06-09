package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
)

// FakeAuthMiddleware 为演示环境注入固定用户。
func FakeAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := int64(300241)
		if headerUserID := r.Header.Get("X-Demo-User-ID"); headerUserID != "" {
			if parsed, err := strconv.ParseInt(headerUserID, 10, 64); err == nil && parsed > 0 {
				userID = parsed
			}
		}
		if queryUserID := r.URL.Query().Get("userId"); queryUserID != "" {
			if parsed, err := strconv.ParseInt(queryUserID, 10, 64); err == nil && parsed > 0 {
				userID = parsed
			}
		}
		if r.Method == http.MethodPost && r.Body != nil {
			body, err := io.ReadAll(r.Body)
			if err == nil {
				var carrier UserIDCarrier
				if json.Unmarshal(body, &carrier) == nil && carrier.UserID > 0 {
					userID = carrier.UserID
				}
				r.Body = io.NopCloser(bytes.NewReader(body))
			}
		}
		next.ServeHTTP(w, r.WithContext(WithUserID(r.Context(), userID)))
	})
}
