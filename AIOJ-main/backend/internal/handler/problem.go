package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/terminaloj/backend/internal/middleware"
	"github.com/terminaloj/backend/internal/models"
	"github.com/terminaloj/backend/internal/utils"
	"gorm.io/gorm"
)

type ProblemHandler struct {
	DB *gorm.DB
}

type problemPayload struct {
	ID              uint64             `json:"id" binding:"required"`
	Title           string             `json:"title" binding:"required"`
	Difficulty      string             `json:"difficulty" binding:"required"`
	DifficultyScore int                `json:"difficultyScore"`
	Tags            models.StringSlice `json:"tags"`
	Content         string             `json:"content" binding:"required"`
	TimeLimit       int                `json:"timeLimit"`
	MemoryLimit     int                `json:"memoryLimit"`
	OutputLimitKB   int32              `json:"outputLimitKb"`
	Source          string             `json:"source"`
	TestCases       models.TestCases   `json:"testCases" binding:"required"`
}

type updateProblemReq struct {
	Title           string             `json:"title" binding:"required"`
	Difficulty      string             `json:"difficulty" binding:"required"`
	DifficultyScore int                `json:"difficultyScore"`
	Tags            models.StringSlice `json:"tags"`
	Content         string             `json:"content" binding:"required"`
	TimeLimit       int                `json:"timeLimit"`
	MemoryLimit     int                `json:"memoryLimit"`
	OutputLimitKB   int32              `json:"outputLimitKb"`
	Source          string             `json:"source"`
	TestCases       models.TestCases   `json:"testCases" binding:"required"`
}

type problemListItem struct {
	ID              uint64             `json:"id"`
	Title           string             `json:"title"`
	Difficulty      string             `json:"difficulty"`
	DifficultyScore int                `json:"difficultyScore"`
	Tags            models.StringSlice `json:"tags"`
	AcceptRate      string             `json:"acceptRate"`
	SubmitCount     int                `json:"submitCount"`
	Accepted        bool               `json:"accepted"`
}

func (h *ProblemHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	keyword := c.Query("keyword")
	difficulty := c.Query("difficulty")
	tag := c.Query("tag")

	q := h.DB.Model(&models.Problem{})
	if keyword != "" {
		like := "%" + keyword + "%"
		if id, err := strconv.ParseUint(keyword, 10, 64); err == nil {
			q = q.Where("id = ? OR title LIKE ?", id, like)
		} else {
			q = q.Where("title LIKE ?", like)
		}
	}
	if difficulty != "" {
		q = q.Where("difficulty = ?", difficulty)
	}
	if tag != "" {
		q = q.Where("JSON_CONTAINS(tags, JSON_QUOTE(?))", tag)
	}

	var total int64
	q.Count(&total)

	var rows []models.Problem
	if err := q.Order("id ASC").Offset((page - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}

	uid, logged := middleware.CurrentUserID(c)
	acceptedSet := map[uint64]bool{}
	if logged && len(rows) > 0 {
		ids := make([]uint64, 0, len(rows))
		for _, p := range rows {
			ids = append(ids, p.ID)
		}
		var solved []uint64
		h.DB.Model(&models.Submission{}).
			Where("user_id = ? AND status = ? AND problem_id IN ?", uid, models.StatusAccepted, ids).
			Distinct("problem_id").Pluck("problem_id", &solved)
		for _, id := range solved {
			acceptedSet[id] = true
		}
	}

	list := make([]problemListItem, 0, len(rows))
	for _, p := range rows {
		list = append(list, problemListItem{
			ID:              p.ID,
			Title:           p.Title,
			Difficulty:      p.Difficulty,
			DifficultyScore: p.DifficultyScore,
			Tags:            p.Tags,
			AcceptRate:      p.AcceptRate(),
			SubmitCount:     p.SubmitCount,
			Accepted:        acceptedSet[p.ID],
		})
	}
	utils.OK(c, gin.H{"list": list, "total": total})
}

func (h *ProblemHandler) Detail(c *gin.Context) {
	p, err := h.loadProblem(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "题号不合法")
		return
	}
	if p == nil {
		utils.NotFound(c, "题目不存在")
		return
	}

	accepted := false
	if uid, ok := middleware.CurrentUserID(c); ok {
		var cnt int64
		h.DB.Model(&models.Submission{}).
			Where("user_id = ? AND problem_id = ? AND status = ?", uid, p.ID, models.StatusAccepted).
			Count(&cnt)
		accepted = cnt > 0
	}

	utils.OK(c, gin.H{
		"id":              p.ID,
		"title":           p.Title,
		"difficulty":      p.Difficulty,
		"difficultyScore": p.DifficultyScore,
		"tags":            p.Tags,
		"acceptRate":      p.AcceptRate(),
		"submitCount":     p.SubmitCount,
		"accepted":        accepted,
		"content":         p.Content,
		"timeLimit":       p.TimeLimit,
		"memoryLimit":     p.MemoryLimit,
		"source":          p.Source,
		"outputLimitKb":   p.OutputLimitKBOrDefault(),
	})
}

func (h *ProblemHandler) AdminDetail(c *gin.Context) {
	p, err := h.loadProblem(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "题号不合法")
		return
	}
	if p == nil {
		utils.NotFound(c, "题目不存在")
		return
	}
	utils.OK(c, buildAdminProblemDetail(*p))
}

func (h *ProblemHandler) Create(c *gin.Context) {
	var req problemPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "invalid problem payload")
		return
	}
	if !normalizeProblemPayload(&req) {
		utils.BadRequest(c, "at least one test case is required")
		return
	}

	var count int64
	h.DB.Model(&models.Problem{}).Where("id = ?", req.ID).Count(&count)
	if count > 0 {
		utils.BadRequest(c, "problem id already exists")
		return
	}

	problem := buildProblem(req)
	if err := h.DB.Create(&problem).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}
	utils.OK(c, buildAdminProblemDetail(problem))
}

func (h *ProblemHandler) Update(c *gin.Context) {
	p, err := h.loadProblem(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "题号不合法")
		return
	}
	if p == nil {
		utils.NotFound(c, "题目不存在")
		return
	}

	var req updateProblemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "invalid problem payload")
		return
	}

	payload := problemPayload{
		ID:              p.ID,
		Title:           req.Title,
		Difficulty:      req.Difficulty,
		DifficultyScore: req.DifficultyScore,
		Tags:            req.Tags,
		Content:         req.Content,
		TimeLimit:       req.TimeLimit,
		MemoryLimit:     req.MemoryLimit,
		OutputLimitKB:   req.OutputLimitKB,
		Source:          req.Source,
		TestCases:       req.TestCases,
	}
	if !normalizeProblemPayload(&payload) {
		utils.BadRequest(c, "at least one test case is required")
		return
	}

	p.Title = payload.Title
	p.Difficulty = payload.Difficulty
	p.DifficultyScore = payload.DifficultyScore
	p.Tags = payload.Tags
	p.Content = payload.Content
	p.TimeLimit = payload.TimeLimit
	p.MemoryLimit = payload.MemoryLimit
	p.OutputLimitKB = payload.OutputLimitKB
	p.Source = payload.Source
	p.TestCases = payload.TestCases

	if err := h.DB.Save(p).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}
	utils.OK(c, buildAdminProblemDetail(*p))
}

func (h *ProblemHandler) Delete(c *gin.Context) {
	p, err := h.loadProblem(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "题号不合法")
		return
	}
	if p == nil {
		utils.NotFound(c, "题目不存在")
		return
	}

	if err := h.DB.Delete(p).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}
	utils.OK(c, gin.H{"deleted": true, "id": p.ID})
}

func (h *ProblemHandler) loadProblem(rawID string) (*models.Problem, error) {
	id, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil {
		return nil, err
	}

	var p models.Problem
	if err := h.DB.First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func normalizeProblemPayload(req *problemPayload) bool {
	if len(req.TestCases) == 0 {
		return false
	}
	if req.TimeLimit <= 0 {
		req.TimeLimit = 1000
	}
	if req.MemoryLimit <= 0 {
		req.MemoryLimit = 256
	}
	if req.OutputLimitKB <= 0 {
		req.OutputLimitKB = 1024
	}
	if req.DifficultyScore <= 0 {
		req.DifficultyScore = 800
	}
	if req.Source == "" {
		req.Source = "Admin"
	}
	return true
}

func buildProblem(req problemPayload) models.Problem {
	return models.Problem{
		ID:              req.ID,
		Title:           req.Title,
		Difficulty:      req.Difficulty,
		DifficultyScore: req.DifficultyScore,
		Tags:            req.Tags,
		Content:         req.Content,
		TimeLimit:       req.TimeLimit,
		MemoryLimit:     req.MemoryLimit,
		OutputLimitKB:   req.OutputLimitKB,
		Source:          req.Source,
		TestCases:       req.TestCases,
	}
}

func buildAdminProblemDetail(problem models.Problem) gin.H {
	return gin.H{
		"id":              problem.ID,
		"title":           problem.Title,
		"difficulty":      problem.Difficulty,
		"difficultyScore": problem.DifficultyScore,
		"tags":            problem.Tags,
		"content":         problem.Content,
		"timeLimit":       problem.TimeLimit,
		"memoryLimit":     problem.MemoryLimit,
		"outputLimitKb":   problem.OutputLimitKBOrDefault(),
		"source":          problem.Source,
		"testCases":       problem.TestCases,
		"submitCount":     problem.SubmitCount,
		"acceptRate":      problem.AcceptRate(),
	}
}

type AnnouncementHandler struct {
	DB *gorm.DB
}

func (h *AnnouncementHandler) List(c *gin.Context) {
	var rows []models.Announcement
	if err := h.DB.Order("id DESC").Limit(20).Find(&rows).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}
	utils.OK(c, rows)
}
