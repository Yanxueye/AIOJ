package handler

import (
	"fmt"
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

func (h *UserHandler) Profile(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	var u models.User
	if err := h.DB.First(&u, uid).Error; err != nil {
		utils.NotFound(c, "用户不存在")
		return
	}
	utils.OK(c, buildProfile(h.DB, &u, true))
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
