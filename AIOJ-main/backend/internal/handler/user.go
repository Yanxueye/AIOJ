package handler

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/terminaloj/backend/internal/middleware"
	"github.com/terminaloj/backend/internal/models"
	"github.com/terminaloj/backend/internal/utils"
	"gorm.io/gorm"
)

type UserHandler struct {
	DB *gorm.DB
}

type updateProfileReq struct {
	Email *string `json:"email"`
	Bio   *string `json:"bio"`
}

type adminUpdateRoleReq struct {
	Role string `json:"role" binding:"required"`
}

func (h *UserHandler) Profile(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	var u models.User
	if err := h.DB.First(&u, uid).Error; err != nil {
		utils.NotFound(c, "用户不存在")
		return
	}
	utils.OK(c, buildProfile(h.DB, &u, true))
}

// RatingHistory returns the user's rating change history.
func (h *UserHandler) RatingHistory(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	if limit < 1 || limit > 500 {
		limit = 100
	}

	var history []models.RatingHistory
	if err := h.DB.Where("user_id = ?", uid).Order("created_at DESC").Limit(limit).Find(&history).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}
	utils.OK(c, gin.H{"history": history})
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	var req updateProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数不合法")
		return
	}
	var u models.User
	if err := h.DB.First(&u, uid).Error; err != nil {
		utils.NotFound(c, "用户不存在")
		return
	}
	if req.Email != nil {
		if !emailRe.MatchString(*req.Email) {
			utils.BadRequest(c, "邮箱格式不合法")
			return
		}
		u.Email = *req.Email
	}
	if req.Bio != nil {
		b := *req.Bio
		if len([]rune(b)) > 200 {
			utils.BadRequest(c, "个人简介不能超过 200 字")
			return
		}
		u.Bio = b
	}
	if err := h.DB.Save(&u).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}
	utils.OK(c, buildProfile(h.DB, &u, true))
}

func (h *UserHandler) AdminList(c *gin.Context) {
	var users []models.User
	if err := h.DB.Order("id ASC").Find(&users).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}

	items := make([]gin.H, 0, len(users))
	for _, u := range users {
		items = append(items, gin.H{
			"id":           u.ID,
			"username":     u.Username,
			"email":        u.Email,
			"role":         u.Role,
			"rating":       u.Rating,
			"registeredAt": u.CreatedAt.Format("2006-01-02"),
		})
	}
	utils.OK(c, gin.H{"items": items})
}

func (h *UserHandler) AdminUpdateRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "用户编号不合法")
		return
	}

	var req adminUpdateRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数不合法")
		return
	}
	if !isValidRole(req.Role) {
		utils.BadRequest(c, "角色不合法")
		return
	}

	var user models.User
	if err := h.DB.First(&user, id).Error; err != nil {
		utils.NotFound(c, "用户不存在")
		return
	}
	user.Role = req.Role
	if err := h.DB.Save(&user).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}

	utils.OK(c, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"role":     user.Role,
	})
}

// buildProfile composes a Profile struct matching the frontend contract.
// extended=true includes difficulty/algorithm breakdown + recent activity
// (used for GET /user/profile) while extended=false is used at login time.
func buildProfile(db *gorm.DB, u *models.User, extended bool) models.Profile {
	var solved, total int64
	db.Model(&models.Submission{}).
		Where("user_id = ? AND status = ?", u.ID, models.StatusAccepted).
		Distinct("problem_id").Count(&solved)
	db.Model(&models.Submission{}).Where("user_id = ?", u.ID).Count(&total)
	rate := "0.0"
	if total > 0 {
		rate = fmt.Sprintf("%.1f", float64(solved)*100.0/float64(total))
	}

	var higher int64
	db.Model(&models.User{}).Where("rating > ?", u.Rating).Count(&higher)
	rank := int(higher) + 1

	p := models.Profile{
		ID:                u.ID,
		Username:          u.Username,
		Email:             u.Email,
		Role:              u.Role,
		Avatar:            u.Avatar,
		Bio:               u.Bio,
		Rating:            u.Rating,
		Rank:              rank,
		SolvedCount:       int(solved),
		TotalSubmissions: int(total),
		AcceptRate:        rate,
		RegisteredAt:      u.CreatedAt.Format("2006-01-02"),
	}
	if !extended {
		return p
	}

	byDiff := map[string]int{"简单": 0, "中等": 0, "困难": 0}
	byAlgo := map[string]int{}
	type row struct {
		Difficulty string
		Tags       models.StringSlice
	}
	var rows []row
	db.Raw(`SELECT p.difficulty, p.tags
		FROM submissions s
		JOIN problems p ON p.id = s.problem_id
		WHERE s.user_id = ? AND s.status = 'Accepted'
		GROUP BY s.problem_id`, u.ID).Scan(&rows)
	for _, r := range rows {
		byDiff[r.Difficulty]++
		for _, tag := range r.Tags {
			byAlgo[tag]++
		}
	}
	p.SolvedByDifficulty = byDiff
	p.SolvedByAlgorithm = byAlgo

	p.RecentActivity = buildRecentActivity(db, u.ID)
	p.Favorites = buildFavorites(db, u.ID)
	p.RecentSubmissions = buildRecentSubmissions(db, u.ID)
	return p
}

func buildRecentActivity(db *gorm.DB, uid uint64) []models.DailyCount {
	type agg struct {
		Day   string
		Count int
	}
	var rows []agg
	db.Raw(`SELECT DATE(created_at) AS day, COUNT(*) AS count
		FROM submissions
		WHERE user_id = ? AND created_at >= ?
		GROUP BY day ORDER BY day DESC`,
		uid, time.Now().AddDate(0, 0, -14)).Scan(&rows)
	out := make([]models.DailyCount, 0, len(rows))
	for _, r := range rows {
		out = append(out, models.DailyCount{Date: r.Day, Count: r.Count})
	}
	return out
}

func buildFavorites(db *gorm.DB, uid uint64) []models.FavoriteDigest {
	type row struct {
		ProblemID   uint64
		Title       string
		Difficulty  string
		SubmitCount int
		AcceptCount int
		CreatedAt   time.Time
	}
	var rows []row
	db.Raw(`SELECT f.problem_id, p.title, p.difficulty, p.submit_count, p.accept_count, f.created_at
		FROM favorites f
		JOIN problems p ON p.id = f.problem_id
		WHERE f.user_id = ?
		ORDER BY f.created_at DESC
		LIMIT 12`, uid).Scan(&rows)
	out := make([]models.FavoriteDigest, 0, len(rows))
	for _, r := range rows {
		acceptRate := "0.0"
		if r.SubmitCount > 0 {
			acceptRate = fmt.Sprintf("%.1f", float64(r.AcceptCount)*100.0/float64(r.SubmitCount))
		}
		out = append(out, models.FavoriteDigest{
			ProblemID:   r.ProblemID,
			Title:       r.Title,
			Difficulty:  r.Difficulty,
			AcceptRate:  acceptRate,
			FavoritedAt: r.CreatedAt.Format("2006-01-02 15:04"),
		})
	}
	return out
}

func buildRecentSubmissions(db *gorm.DB, uid uint64) []models.SubmissionTimelineItem {
	var rows []models.Submission
	db.Where("user_id = ? AND source = ?", uid, "submit").Order("created_at DESC").Limit(12).Find(&rows)
	out := make([]models.SubmissionTimelineItem, 0, len(rows))
	for _, item := range rows {
		out = append(out, models.SubmissionTimelineItem{
			SubmissionID: item.ID,
			ProblemID:    item.ProblemID,
			ProblemTitle: item.ProblemTitle,
			Status:       item.Status,
			Language:     item.Language,
			CreatedAt:    item.CreatedAt.Format("2006-01-02 15:04"),
		})
	}
	return out
}

func isValidRole(role string) bool {
	switch role {
	case models.RoleUser, models.RoleAdmin:
		return true
	default:
		return false
	}
}

func (h *UserHandler) Heatmap(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	type agg struct {
		Day   string
		Count int
	}
	var rows []agg
	h.DB.Raw(`SELECT DATE(CONVERT_TZ(created_at, '+00:00', @@session.time_zone)) AS day, COUNT(*) AS count
		FROM submissions
		WHERE user_id = ? AND created_at >= ? AND source = 'submit'
		GROUP BY day ORDER BY day`,
		uid, time.Now().AddDate(-1, 0, 0)).Scan(&rows)
	out := make([]models.DailyCount, 0, len(rows))
	for _, r := range rows {
		out = append(out, models.DailyCount{Date: r.Day, Count: r.Count})
	}
	utils.OK(c, gin.H{"items": out})
}
