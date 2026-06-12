package handler

import (
	"github.com/gin-gonic/gin"
	aisvc "github.com/terminaloj/backend/internal/ai"
	"github.com/terminaloj/backend/internal/config"
	"github.com/terminaloj/backend/internal/judger"
	"github.com/terminaloj/backend/internal/middleware"
	"github.com/terminaloj/backend/internal/mq"
	"github.com/terminaloj/backend/internal/utils"
	"gorm.io/gorm"
)

// BuildRouter wires middlewares + route groups. Grouping keeps the auth
// boundary explicit and makes it trivial to add versioned endpoints later.
func BuildRouter(db *gorm.DB, broker *mq.Broker, jdClient judger.JudgerClient, jwt *utils.JWTManager, cfg *config.Config) *gin.Engine {
	gin.SetMode(cfg.Server.Mode)
	r := gin.New()
	r.Use(gin.Logger(), middleware.Recovery(), middleware.CORS())

	auth := &AuthHandler{DB: db, JWT: jwt}
	user := &UserHandler{DB: db}
	problem := &ProblemHandler{DB: db}
	studyPlan := &StudyPlanHandler{DB: db}
	announcement := &AnnouncementHandler{DB: db}
	submission := &SubmissionHandler{DB: db, Broker: broker, Judger: jdClient}
	ai := &AIHandler{DB: db, Client: aisvc.NewClient(cfg.AI)}
	audit := &AuditHandler{DB: db}
	knowledge := &KnowledgeHandler{DB: db}
	recommendation := &RecommendationHandler{DB: db}

	r.GET("/healthz", func(c *gin.Context) { utils.OK(c, gin.H{"status": "ok"}) })

	api := r.Group("/api")
	{
		api.POST("/auth/login", auth.Login)
		api.POST("/auth/register", auth.Register)
		api.GET("/announcements", announcement.List)
		api.GET("/daily-challenge", studyPlan.DailyChallenge)
		api.GET("/study-plans", optionalAuth(jwt), studyPlan.List)
		api.GET("/study-plans/:id", optionalAuth(jwt), studyPlan.Detail)

		// Public GET for list/detail — optional token is read by handler to
		// enrich the response with the caller's `accepted` flag.
		api.GET("/problems", optionalAuth(jwt), problem.List)

		// Knowledge graph — optional token enriches with user mastery data.
		api.GET("/knowledge", knowledge.List)
		api.GET("/knowledge/graph", optionalAuth(jwt), knowledge.Graph)
		api.GET("/knowledge/:id/problems", knowledge.ProblemsForKP)
		api.GET("/recommendations/daily", optionalAuth(jwt), recommendation.DailyRecommendation)
	}

	authed := r.Group("/api", middleware.JWTAuth(jwt))
	{
		authed.GET("/user/profile", user.Profile)
		authed.PUT("/user/profile", user.UpdateProfile)
		authed.GET("/user/heatmap", user.Heatmap)
		authed.GET("/learning-path", recommendation.LearningPath)
		authed.GET("/weakness-analysis", recommendation.WeaknessAnalysis)
		authed.GET("/study-plans/checkins", studyPlan.Checkins)
		authed.GET("/admin/users", middleware.RequireAdmin(), user.AdminList)
		authed.PUT("/admin/users/:id/role", middleware.RequireAdmin(), user.AdminUpdateRole)
		authed.GET("/admin/audit-logs", middleware.RequireAdmin(), audit.List)

		authed.GET("/problems/:id", problem.Detail)
		authed.POST("/problems/:id/favorite", problem.Favorite)
		authed.DELETE("/problems/:id/favorite", problem.Unfavorite)
		authed.POST("/problems/:id/solution", problem.UpsertSolution)
		authed.GET("/problems/:id/my-solution", problem.UserSolutionForProblem)
		authed.GET("/my/solutions", problem.MySolutions)
		authed.GET("/my/solutions/:id", problem.MySolutionDetail)
		authed.GET("/solutions/:id", problem.SolutionDetail)
		authed.POST("/solutions/:sid/like", problem.LikeSolution)
		authed.DELETE("/solutions/:sid", middleware.RequireAdmin(), problem.DeleteSolution)
		authed.POST("/problems", middleware.RequireAdmin(), problem.Create)
		authed.GET("/admin/problems/:id", middleware.RequireAdmin(), problem.AdminDetail)
		authed.GET("/admin/problems/:id/versions", middleware.RequireAdmin(), problem.Versions)
		authed.PUT("/problems/:id", middleware.RequireAdmin(), problem.Update)
		authed.POST("/admin/problems/:id/publish", middleware.RequireAdmin(), problem.Publish)
		authed.POST("/admin/problems/:id/rollback", middleware.RequireAdmin(), problem.Rollback)
		authed.POST("/admin/problems/:id/rejudge", middleware.RequireAdmin(), problem.Rejudge)
		authed.GET("/admin/problems/:id/rejudge-jobs", middleware.RequireAdmin(), problem.RejudgeJobs)
		authed.DELETE("/problems/:id", middleware.RequireAdmin(), problem.Delete)

		authed.POST("/problems/:id/run", submission.Run)
		authed.POST("/submissions",
			middleware.PerUserRateLimit(cfg.RateLimit.SubmitPerMinute, cfg.RateLimit.SubmitBurst),
			submission.Submit,
		)
		authed.GET("/submissions", submission.List)
		authed.GET("/submissions/:id", submission.Detail)
		authed.GET("/submissions/:id/stream", submission.Stream)
		authed.GET("/submissions/:id/cases", submission.Cases)
		authed.GET("/submissions/:id/output", submission.Output)

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
				c.Set("x-user-role", claims.Role)
			}
		}
		c.Next()
	}
}
