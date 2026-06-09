package handler

import (
	"github.com/gin-gonic/gin"
	aisvc "github.com/terminaloj/backend/internal/ai"
	"github.com/terminaloj/backend/internal/config"
	"github.com/terminaloj/backend/internal/middleware"
	"github.com/terminaloj/backend/internal/mq"
	"github.com/terminaloj/backend/internal/utils"
	"gorm.io/gorm"
)

// BuildRouter wires middlewares + route groups. Grouping keeps the auth
// boundary explicit and makes it trivial to add versioned endpoints later.
func BuildRouter(db *gorm.DB, broker *mq.Broker, jwt *utils.JWTManager, cfg *config.Config) *gin.Engine {
	gin.SetMode(cfg.Server.Mode)
	r := gin.New()
	r.Use(gin.Logger(), middleware.Recovery(), middleware.CORS())

	auth := &AuthHandler{DB: db, JWT: jwt}
	user := &UserHandler{DB: db}
	problem := &ProblemHandler{DB: db}
	announcement := &AnnouncementHandler{DB: db}
	submission := &SubmissionHandler{DB: db, Broker: broker}
	ai := &AIHandler{DB: db, Client: aisvc.NewClient(cfg.AI)}

	r.GET("/healthz", func(c *gin.Context) { utils.OK(c, gin.H{"status": "ok"}) })

	api := r.Group("/api")
	{
		api.POST("/auth/login", auth.Login)
		api.POST("/auth/register", auth.Register)
		api.GET("/announcements", announcement.List)

		// Public GET for list/detail — optional token is read by handler to
		// enrich the response with the caller's `accepted` flag.
		api.GET("/problems", optionalAuth(jwt), problem.List)
	}

	authed := r.Group("/api", middleware.JWTAuth(jwt))
	{
		authed.GET("/user/profile", user.Profile)
		authed.PUT("/user/profile", user.UpdateProfile)

		authed.GET("/problems/:id", problem.Detail)
		authed.POST("/problems", middleware.RequireAdmin(), problem.Create)
		authed.GET("/admin/problems/:id", middleware.RequireAdmin(), problem.AdminDetail)
		authed.PUT("/problems/:id", middleware.RequireAdmin(), problem.Update)
		authed.DELETE("/problems/:id", middleware.RequireAdmin(), problem.Delete)

		authed.POST("/submissions",
			middleware.PerUserRateLimit(cfg.RateLimit.SubmitPerMinute, cfg.RateLimit.SubmitBurst),
			submission.Submit,
		)
		authed.GET("/submissions", submission.List)
		authed.GET("/submissions/:id", submission.Detail)

		authed.POST("/ai/chat", ai.Chat)
		authed.GET("/ai/history", ai.History)
		authed.GET("/ai/conversations/:id/messages", ai.Messages)
		authed.POST("/ai/code-diagnosis", ai.CodeDiagnosis)
		authed.POST("/ai/knowledge-graph", ai.KnowledgeGraph)
		authed.POST("/ai/solve", ai.Solve)
	}
	return r
}

// optionalAuth parses the JWT if present but does not abort when missing or
// invalid. Used on endpoints that behave differently for logged-in users.
func optionalAuth(mgr *utils.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.Next()
			return
		}
		if len(header) > 7 && header[:7] == "Bearer " {
			if claims, err := mgr.Parse(header[7:]); err == nil {
				c.Set("x-user-id", claims.UserID)
				c.Set("x-username", claims.Username)
			}
		}
		c.Next()
	}
}
