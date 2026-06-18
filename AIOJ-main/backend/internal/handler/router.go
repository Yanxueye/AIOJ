package handler

import (
	"fmt"
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
	ai := &AIHandler{DB: db, Client: aisvc.NewClient(cfg.AI), Judger: jdClient}
	audit := &AuditHandler{DB: db}
	knowledge := &KnowledgeHandler{DB: db}
	recommendation := &RecommendationHandler{DB: db}
	tag := &TagHandler{}

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

		// Algorithm tags — public, used for problem creation and AI alignment.
		api.GET("/tags", tag.List)
		api.GET("/tags/names", tag.Names)
	}
	// Agent internal API (called by agent-service, auth via X-User-ID, no JWT required)
	agentAuth := func(c *gin.Context) {
		if uid := c.GetHeader("X-User-ID"); uid != "" {
			var id uint64
			if _, err := fmt.Sscanf(uid, "%d", &id); err == nil {
				c.Set("x-user-id", id)
			}
		}
		c.Next()
	}
	agentGroup := r.Group("/api", agentAuth)
	{
		agentGroup.POST("/agent/problems", ai.QueryUserProblems)
		agentGroup.POST("/agent/judge", ai.SubmitAndJudge)
		agentGroup.POST("/agent/code", ai.GetUserCode)
		agentGroup.POST("/agent/search-problems", ai.SearchProblems)
	}

	authed := r.Group("/api", middleware.JWTAuth(jwt))
	{
		authed.GET("/user/profile", user.Profile)
		authed.PUT("/user/profile", user.UpdateProfile)
		authed.GET("/user/rating-history", user.RatingHistory)
		authed.GET("/user/heatmap", user.Heatmap)
		authed.GET("/learning-path", recommendation.LearningPath)
		authed.GET("/weakness-analysis", recommendation.WeaknessAnalysis)
		authed.GET("/study-plans/checkins", studyPlan.Checkins)
		authed.POST("/study-plans", studyPlan.Create)
		authed.PUT("/study-plans/:id", studyPlan.Update)
		authed.DELETE("/study-plans/:id", studyPlan.Delete)
		authed.POST("/study-plans/:id/favorite", studyPlan.Favorite)
		authed.POST("/knowledge/problems-by-tags", knowledge.ProblemsByTags)
		authed.GET("/admin/users", middleware.RequireAdminDB(db), user.AdminList)
		authed.PUT("/admin/users/:id/role", middleware.RequireAdminDB(db), user.AdminUpdateRole)
		authed.GET("/admin/audit-logs", middleware.RequireAdminDB(db), audit.List)

		authed.GET("/problems/:id", problem.Detail)
		authed.POST("/problems/:id/favorite", problem.Favorite)
		authed.DELETE("/problems/:id/favorite", problem.Unfavorite)
		authed.POST("/problems/:id/solution", problem.UpsertSolution)
		authed.GET("/problems/:id/my-solution", problem.UserSolutionForProblem)
		authed.GET("/my/solutions", problem.MySolutions)
		authed.GET("/my/solutions/:id", problem.MySolutionDetail)
		authed.GET("/solutions/:id", problem.SolutionDetail)
		authed.POST("/solutions/:sid/like", problem.LikeSolution)
		authed.DELETE("/solutions/:sid", middleware.RequireAdminDB(db), problem.DeleteSolution)
		authed.POST("/problems", middleware.RequireAdminDB(db), problem.Create)
		authed.GET("/admin/problems/:id", middleware.RequireAdminDB(db), problem.AdminDetail)
		authed.GET("/admin/problems/:id/versions", middleware.RequireAdminDB(db), problem.Versions)
		authed.PUT("/problems/:id", middleware.RequireAdminDB(db), problem.Update)
		authed.POST("/admin/problems/:id/publish", middleware.RequireAdminDB(db), problem.Publish)
		authed.POST("/admin/problems/:id/rollback", middleware.RequireAdminDB(db), problem.Rollback)
		authed.DELETE("/problems/:id", middleware.RequireAdminDB(db), problem.Delete)

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


		aiRateLimit := middleware.PerUserRateLimit(10, 3)
		authed.POST("/ai/chat", aiRateLimit, ai.Chat)
		authed.GET("/ai/history", ai.History)
		authed.GET("/ai/conversations/:id/messages", ai.Messages)
		authed.DELETE("/ai/conversations/:id", ai.DeleteConversation)
		authed.POST("/ai/code-diagnosis", aiRateLimit, ai.CodeDiagnosis)
		authed.POST("/ai/generate-solution", aiRateLimit, ai.GenerateSolution)
		authed.POST("/ai/knowledge-graph", aiRateLimit, ai.KnowledgeGraph)
		authed.POST("/ai/solve", aiRateLimit, ai.Solve)
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
