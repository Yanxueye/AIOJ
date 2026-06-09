package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/terminaloj/backend/internal/models"
	"github.com/terminaloj/backend/internal/utils"
)

const (
	ctxUserID   = "x-user-id"
	ctxUsername = "x-username"
	ctxUserRole = "x-user-role"
)

// JWTAuth verifies the Authorization header and stashes the user in the gin
// context. Requests without a valid token are short-circuited with 401.
func JWTAuth(mgr *utils.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			utils.Unauthorized(c, "missing authorization header")
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			utils.Unauthorized(c, "invalid authorization format")
			return
		}
		claims, err := mgr.Parse(strings.TrimSpace(parts[1]))
		if err != nil {
			utils.Unauthorized(c, "invalid or expired token")
			return
		}
		c.Set(ctxUserID, claims.UserID)
		c.Set(ctxUsername, claims.Username)
		c.Set(ctxUserRole, claims.Role)
		c.Next()
	}
}

// CurrentUserID extracts the authenticated user id from the context. It
// returns (0, false) if the middleware was not executed.
func CurrentUserID(c *gin.Context) (uint64, bool) {
	v, ok := c.Get(ctxUserID)
	if !ok {
		return 0, false
	}
	id, ok := v.(uint64)
	return id, ok
}

// CurrentUsername extracts the authenticated username from the context.
func CurrentUsername(c *gin.Context) string {
	v, ok := c.Get(ctxUsername)
	if !ok {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// CurrentUserRole extracts the authenticated user role from the context.
func CurrentUserRole(c *gin.Context) string {
	v, ok := c.Get(ctxUserRole)
	if !ok {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// RequireAdmin only allows admin users through.
func RequireAdmin() gin.HandlerFunc {
	return RequireRoles(models.RoleAdmin)
}

func RequireRoles(allowed ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := CurrentUserRole(c)
		for _, item := range allowed {
			if role == item {
				c.Next()
				return
			}
		}
		if len(allowed) == 0 {
			utils.Forbidden(c, "role required")
			return
		}
		utils.Forbidden(c, "insufficient role")
		return
	}
}

func CanEditProblems(role string) bool {
	switch role {
	case models.RoleProblemEditor, models.RoleAdmin:
		return true
	default:
		return false
	}
}

func CanReviewProblems(role string) bool {
	switch role {
	case models.RoleReviewer, models.RoleAdmin:
		return true
	default:
		return false
	}
}

func CanTriggerRejudge(role string) bool {
	switch role {
	case models.RoleOperator, models.RoleAdmin:
		return true
	default:
		return false
	}
}

func RequireProblemEditor() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !CanEditProblems(CurrentUserRole(c)) {
			utils.Forbidden(c, "problem editor role required")
			return
		}
		c.Next()
	}
}

func RequireReviewer() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !CanReviewProblems(CurrentUserRole(c)) {
			utils.Forbidden(c, "reviewer role required")
			return
		}
		c.Next()
	}
}

func RequireOperator() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !CanTriggerRejudge(CurrentUserRole(c)) {
			utils.Forbidden(c, "operator role required")
			return
		}
		c.Next()
	}
}
