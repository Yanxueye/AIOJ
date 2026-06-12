package handler

import (
	"errors"
	"slices"
	"strconv"
	"time"

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
	ID              uint64             `json:"id"`
	Title           string             `json:"title" binding:"required"`
	Difficulty      string             `json:"difficulty" binding:"required"`
	DifficultyScore int                `json:"difficultyScore"`
	Tags            models.StringSlice `json:"tags"`
	Content         string             `json:"content" binding:"required"`
	Constraints     string             `json:"constraints"`
	TimeLimit       int                `json:"timeLimit"`
	MemoryLimit     int                `json:"memoryLimit"`
	OutputLimitKB   int32              `json:"outputLimitKb"`
	Source          string             `json:"source"`
	Editorial       string             `json:"editorial"`
	Samples         []samplePayload    `json:"samples"`
	TestCases       []testCasePayload  `json:"testCases" binding:"required"`
	Templates       []templatePayload  `json:"templates"`
	Status          string             `json:"status"`
	ReviewComment   string             `json:"reviewComment"`
}

type updateProblemReq struct {
	Title           string             `json:"title" binding:"required"`
	Difficulty      string             `json:"difficulty" binding:"required"`
	DifficultyScore int                `json:"difficultyScore"`
	Tags            models.StringSlice `json:"tags"`
	Content         string             `json:"content" binding:"required"`
	Constraints     string             `json:"constraints"`
	TimeLimit       int                `json:"timeLimit"`
	MemoryLimit     int                `json:"memoryLimit"`
	OutputLimitKB   int32              `json:"outputLimitKb"`
	Source          string             `json:"source"`
	Editorial       string             `json:"editorial"`
	Samples         []samplePayload    `json:"samples"`
	TestCases       []testCasePayload  `json:"testCases" binding:"required"`
	Templates       []templatePayload  `json:"templates"`
	Status          string             `json:"status"`
	ReviewComment   string             `json:"reviewComment"`
}

type samplePayload struct {
	Input       string `json:"input"`
	Expected    string `json:"expected"`
	Explanation string `json:"explanation"`
}

type testCasePayload struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
	IsHidden bool   `json:"isHidden"`
}

type templatePayload struct {
	Language string `json:"language"`
	Code     string `json:"code"`
}

type publishReq struct {
	ReviewComment string `json:"reviewComment"`
	VersionID     uint64 `json:"versionId"`
}

type rollbackReq struct {
	VersionID uint64 `json:"versionId" binding:"required"`
}

type rejudgeReq struct {
	Reason          string `json:"reason"`
	TargetVersionID uint64 `json:"targetVersionId"`
}

type solutionReq struct {
	Title       string `json:"title" binding:"required"`
	Content     string `json:"content" binding:"required"`
	Language    string `json:"language"`
	IsPublished bool   `json:"isPublished"`
}

type problemListItem struct {
	ID              uint64             `json:"id"`
	Title           string             `json:"title"`
	Difficulty      string             `json:"difficulty"`
	DifficultyScore int                `json:"difficultyScore"`
	Tags            models.StringSlice `json:"tags"`
	Status          string             `json:"status"`
	AcceptRate      string             `json:"acceptRate"`
	SubmitCount     int                `json:"submitCount"`
	Accepted        bool               `json:"accepted"`
	Favorite        bool               `json:"favorite"`
	Attempted       bool               `json:"attempted"`
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
	statusFilter := c.Query("status")

	q := h.DB.Model(&models.Problem{}).Where("status = ? OR status = ''", models.ProblemStatusPublished)
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
	if err := q.Preload("PublishedVersion").Order("id ASC").Offset((page - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}

	uid, logged := middleware.CurrentUserID(c)
	acceptedSet := map[uint64]bool{}
	attemptedSet := map[uint64]bool{}
	favoriteSet := map[uint64]bool{}
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
		var attempted []uint64
		h.DB.Model(&models.Submission{}).
			Where("user_id = ? AND problem_id IN ? AND source = ?", uid, ids, "submit").
			Distinct("problem_id").Pluck("problem_id", &attempted)
		for _, id := range attempted {
			attemptedSet[id] = true
		}
		var favorites []uint64
		h.DB.Model(&models.Favorite{}).
			Where("user_id = ? AND problem_id IN ?", uid, ids).
			Pluck("problem_id", &favorites)
		for _, id := range favorites {
			favoriteSet[id] = true
		}
	}

	list := make([]problemListItem, 0, len(rows))
	for _, p := range rows {
		item := problemListItem{
			ID:              p.ID,
			Title:           p.Title,
			Difficulty:      p.Difficulty,
			DifficultyScore: p.DifficultyScore,
			Tags:            p.Tags,
			Status:          p.Status,
			AcceptRate:      p.AcceptRate(),
			SubmitCount:     p.SubmitCount,
			Accepted:        acceptedSet[p.ID],
			Favorite:        favoriteSet[p.ID],
			Attempted:       attemptedSet[p.ID],
		}
		if !matchesStatusFilter(statusFilter, item) {
			continue
		}
		list = append(list, item)
	}
	utils.OK(c, gin.H{"list": list, "total": len(list)})
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

	if err := h.DB.
		Preload("PublishedVersion.Samples").
		Preload("PublishedVersion.Templates").
		Preload("CurrentVersion.Samples").
		Preload("CurrentVersion.Templates").
		First(p, p.ID).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}

	accepted := false
	favorite := false
	if uid, ok := middleware.CurrentUserID(c); ok {
		var cnt int64
		h.DB.Model(&models.Submission{}).
			Where("user_id = ? AND problem_id = ? AND status = ?", uid, p.ID, models.StatusAccepted).
			Count(&cnt)
		accepted = cnt > 0
		h.DB.Model(&models.Favorite{}).
			Where("user_id = ? AND problem_id = ?", uid, p.ID).
			Count(&cnt)
		favorite = cnt > 0
	}

	version := chosenVersion(p)
	if version == nil {
		version = &models.ProblemVersion{
			Title:           p.Title,
			Difficulty:      p.Difficulty,
			DifficultyScore: p.DifficultyScore,
			Tags:            p.Tags,
			Content:         "# 题目内容待补全\n\n当前题目来自旧版数据，管理员可在后台完善题面。",
			TimeLimit:       1000,
			MemoryLimit:     256,
			OutputLimitKB:   1024,
		}
	}
	view := publicProblemView(p, version, accepted)
	view["favorite"] = favorite
	var currentUID uint64
	if uid, ok := middleware.CurrentUserID(c); ok {
		currentUID = uid
		view["mySolution"] = h.userSolution(p.ID, uid)
	}
	view["relatedProblems"] = h.relatedProblems(p, currentUID, 4)
	solutions := h.publishedSolutions(p.ID)
	// Prepend editorial as official solution if it exists
	if version.Editorial != "" {
		editorial := gin.H{
			"id":          uint64(0),
			"userId":      uint64(0),
			"username":    "官方",
			"title":       "官方题解",
			"content":     version.Editorial,
			"language":    "",
			"isPublished": true,
			"isOfficial":  true,
			"likeCount":   0,
			"updatedAt":   "",
		}
		solutions = append([]gin.H{editorial}, solutions...)
	}
	view["solutions"] = solutions
	utils.OK(c, view)
}

func (h *ProblemHandler) UpsertSolution(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "题号不合法")
		return
	}

	var req solutionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数不合法")
		return
	}

	if req.IsPublished {
		var acceptedCount int64
		h.DB.Model(&models.Submission{}).
			Where("user_id = ? AND problem_id = ? AND status = ?", uid, id, models.StatusAccepted).
			Count(&acceptedCount)
		if acceptedCount == 0 {
			utils.Forbidden(c, "通过该题后才能发布题解")
			return
		}
	}

	var user models.User
	if err := h.DB.First(&user, uid).Error; err != nil {
		utils.NotFound(c, "用户不存在")
		return
	}

	var solution models.ProblemSolution
	err = h.DB.Where("problem_id = ? AND user_id = ?", id, uid).First(&solution).Error
	if err == gorm.ErrRecordNotFound {
		solution = models.ProblemSolution{
			ProblemID:   id,
			UserID:      uid,
			Username:    user.Username,
			Title:       req.Title,
			Content:     req.Content,
			Language:    req.Language,
			IsPublished: req.IsPublished,
		}
		if err := h.DB.Create(&solution).Error; err != nil {
			utils.Server(c, err.Error())
			return
		}
		utils.OK(c, solution)
		return
	}
	if err != nil {
		utils.Server(c, err.Error())
		return
	}

	solution.Title = req.Title
	solution.Content = req.Content
	solution.Language = req.Language
	solution.IsPublished = req.IsPublished
	if err := h.DB.Save(&solution).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}
	utils.OK(c, solution)
}

func (h *ProblemHandler) LikeSolution(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	sid, err := strconv.ParseUint(c.Param("sid"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "题解ID不合法")
		return
	}

	var solution models.ProblemSolution
	if err := h.DB.First(&solution, sid).Error; err != nil {
		utils.NotFound(c, "题解不存在")
		return
	}

	// Use transaction to handle concurrent likes safely
	tx := h.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var existing models.SolutionLike
	err = tx.Where("solution_id = ? AND user_id = ?", sid, uid).First(&existing).Error
	if err == nil {
		// Already liked, unlike
		tx.Delete(&existing)
		tx.Model(&solution).UpdateColumn("like_count", gorm.Expr("GREATEST(like_count - 1, 0)"))
		tx.Commit()
		// Re-read actual count from DB
		h.DB.First(&solution, sid)
		utils.OK(c, gin.H{"liked": false, "likeCount": solution.LikeCount})
		return
	}

	like := models.SolutionLike{SolutionID: sid, UserID: uid}
	if err := tx.Create(&like).Error; err != nil {
		tx.Rollback()
		// Likely duplicate key from concurrent request, treat as already liked
		h.DB.First(&solution, sid)
		utils.OK(c, gin.H{"liked": true, "likeCount": solution.LikeCount})
		return
	}
	tx.Model(&solution).UpdateColumn("like_count", gorm.Expr("like_count + 1"))
	tx.Commit()
	// Re-read actual count from DB
	h.DB.First(&solution, sid)
	utils.OK(c, gin.H{"liked": true, "likeCount": solution.LikeCount})
}

func (h *ProblemHandler) DeleteSolution(c *gin.Context) {
	sid, err := strconv.ParseUint(c.Param("sid"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "题解ID不合法")
		return
	}
	var solution models.ProblemSolution
	if err := h.DB.First(&solution, sid).Error; err != nil {
		utils.NotFound(c, "题解不存在")
		return
	}
	// Delete associated likes first
	h.DB.Where("solution_id = ?", sid).Delete(&models.SolutionLike{})
	if err := h.DB.Delete(&solution).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}
	utils.OK(c, gin.H{"deleted": true, "id": sid})
}

func (h *ProblemHandler) MySolutions(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	var rows []models.ProblemSolution
	if err := h.DB.Where("user_id = ?", uid).Order("updated_at DESC").Find(&rows).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}
	items := make([]gin.H, 0, len(rows))
	for _, item := range rows {
		var problem models.Problem
		_ = h.DB.Select("id, title").First(&problem, item.ProblemID).Error
		items = append(items, gin.H{
			"id":          item.ID,
			"problemId":   item.ProblemID,
			"problemTitle": problem.Title,
			"title":       item.Title,
			"content":     item.Content,
			"language":    item.Language,
			"isPublished": item.IsPublished,
			"updatedAt":   item.UpdatedAt.Format("2006-01-02 15:04"),
		})
	}
	utils.OK(c, gin.H{"items": items})
}

func (h *ProblemHandler) MySolutionDetail(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "题解编号不合法")
		return
	}
	var item models.ProblemSolution
	if err := h.DB.Where("id = ? AND user_id = ?", id, uid).First(&item).Error; err != nil {
		utils.NotFound(c, "题解不存在")
		return
	}
	var problem models.Problem
	_ = h.DB.Select("id, title").First(&problem, item.ProblemID).Error
	utils.OK(c, gin.H{
		"id":           item.ID,
		"problemId":    item.ProblemID,
		"problemTitle": problem.Title,
		"userId":       item.UserID,
		"username":     item.Username,
		"title":        item.Title,
		"content":      item.Content,
		"language":     item.Language,
		"isPublished":  item.IsPublished,
		"updatedAt":    item.UpdatedAt.Format("2006-01-02 15:04"),
	})
}

func (h *ProblemHandler) SolutionDetail(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "题解编号不合法")
		return
	}
	var item models.ProblemSolution
	if err := h.DB.Where("id = ?", id).First(&item).Error; err != nil {
		utils.NotFound(c, "题解不存在")
		return
	}
	if !item.IsPublished && item.UserID != uid {
		utils.Forbidden(c, "无权查看该题解")
		return
	}
	utils.OK(c, gin.H{
		"id":          item.ID,
		"problemId":   item.ProblemID,
		"userId":      item.UserID,
		"username":    item.Username,
		"title":       item.Title,
		"content":     item.Content,
		"language":    item.Language,
		"isPublished": item.IsPublished,
		"updatedAt":   item.UpdatedAt.Format("2006-01-02 15:04"),
	})
}

func (h *ProblemHandler) UserSolutionForProblem(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	problemID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "题号不合法")
		return
	}
	var item models.ProblemSolution
	if err := h.DB.Where("problem_id = ? AND user_id = ?", problemID, uid).First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.OK(c, gin.H{})
			return
		}
		utils.Server(c, err.Error())
		return
	}
	utils.OK(c, gin.H{
		"id":          item.ID,
		"problemId":   item.ProblemID,
		"userId":      item.UserID,
		"username":    item.Username,
		"title":       item.Title,
		"content":     item.Content,
		"language":    item.Language,
		"isPublished": item.IsPublished,
		"updatedAt":   item.UpdatedAt.Format("2006-01-02 15:04"),
	})
}

func (h *ProblemHandler) Favorite(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "题号不合法")
		return
	}
	var p models.Problem
	if err := h.DB.First(&p, id).Error; err != nil {
		utils.NotFound(c, "题目不存在")
		return
	}
	var count int64
	h.DB.Model(&models.Favorite{}).Where("user_id = ? AND problem_id = ?", uid, id).Count(&count)
	if count == 0 {
		if err := h.DB.Create(&models.Favorite{UserID: uid, ProblemID: id}).Error; err != nil {
			utils.Server(c, err.Error())
			return
		}
	}
	utils.OK(c, gin.H{"problemId": id, "favorite": true})
}

func (h *ProblemHandler) Unfavorite(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "题号不合法")
		return
	}
	if err := h.DB.Where("user_id = ? AND problem_id = ?", uid, id).Delete(&models.Favorite{}).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}
	utils.OK(c, gin.H{"problemId": id, "favorite": false})
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
	if err := h.DB.
		Preload("CurrentVersion.Samples").
		Preload("CurrentVersion.TestCases").
		Preload("CurrentVersion.Templates").
		Preload("Versions", func(db *gorm.DB) *gorm.DB { return db.Order("version_no DESC") }).
		First(p, p.ID).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}
	utils.OK(c, adminProblemView(p, chosenVersion(p)))
}

func (h *ProblemHandler) Versions(c *gin.Context) {
	p, err := h.loadProblem(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "题号不合法")
		return
	}
	if p == nil {
		utils.NotFound(c, "题目不存在")
		return
	}

	var versions []models.ProblemVersion
	if err := h.DB.Where("problem_id = ?", p.ID).Order("version_no DESC").Find(&versions).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}

	items := make([]gin.H, 0, len(versions))
	for _, item := range versions {
		items = append(items, gin.H{
			"id":          item.ID,
			"versionNo":   item.VersionNo,
			"title":       item.Title,
			"difficulty":  item.Difficulty,
			"createdAt":   item.CreatedAt.UTC().Format(time.RFC3339),
			"publishedAt": timePtrToString(item.PublishedAt),
		})
	}
	utils.OK(c, gin.H{"problemId": p.ID, "items": items})
}

func (h *ProblemHandler) RejudgeJobs(c *gin.Context) {
	p, err := h.loadProblem(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "题号不合法")
		return
	}
	if p == nil {
		utils.NotFound(c, "题目不存在")
		return
	}

	var jobs []models.RejudgeJob
	if err := h.DB.Where("problem_id = ?", p.ID).Order("id DESC").Find(&jobs).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}
	utils.OK(c, gin.H{"problemId": p.ID, "items": jobs})
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

	editorID, _ := middleware.CurrentUserID(c)
	now := time.Now().UTC()
	status := defaultProblemStatus(req.Status)

	// Create problem first (let database assign auto-increment ID)
	problem := models.Problem{
		Title:           req.Title,
		Difficulty:      req.Difficulty,
		DifficultyScore: req.DifficultyScore,
		Tags:            req.Tags,
		Source:          req.Source,
		Status:          status,
		ReviewComment:   req.ReviewComment,
	}
	if err := h.DB.Create(&problem).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}

	// Now create version with the assigned problem ID
	version := buildProblemVersion(problem.ID, 1, req, editorID, nil)
	if err := h.DB.Create(&version).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}

	// Update problem with version references
	problem.CurrentVersionID = &version.ID
	if status == models.ProblemStatusPublished {
		problem.PublishedVersionID = &version.ID
		problem.PublishedAt = &now
		problem.PublishedBy = &editorID
		version.PublishedAt = &now
	}
	problem.LastEditedBy = &editorID
	if err := h.DB.Save(&problem).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}

	if err := h.persistVersionChildren(version.ID, req); err != nil {
		utils.Server(c, err.Error())
		return
	}

	problem.CurrentVersion = &version
	h.writeAuditLog(c, "problem", strconv.FormatUint(problem.ID, 10), "create", "created problem with version 1")
	utils.OK(c, adminProblemView(&problem, &version))
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
		Constraints:     req.Constraints,
		TimeLimit:       req.TimeLimit,
		MemoryLimit:     req.MemoryLimit,
		OutputLimitKB:   req.OutputLimitKB,
		Source:          req.Source,
		Editorial:       req.Editorial,
		Samples:         req.Samples,
		TestCases:       req.TestCases,
		Templates:       req.Templates,
		Status:          req.Status,
		ReviewComment:   req.ReviewComment,
	}
	if !normalizeProblemPayload(&payload) {
		utils.BadRequest(c, "at least one test case is required")
		return
	}

	editorID, _ := middleware.CurrentUserID(c)
	nextVersionNo := latestVersionNo(p) + 1
	version := buildProblemVersion(p.ID, nextVersionNo, payload, editorID, nil)
	if err := h.DB.Create(&version).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}
	if err := h.persistVersionChildren(version.ID, payload); err != nil {
		utils.Server(c, err.Error())
		return
	}

	p.Title = payload.Title
	p.Difficulty = payload.Difficulty
	p.DifficultyScore = payload.DifficultyScore
	p.Tags = payload.Tags
	p.Source = payload.Source
	p.CurrentVersionID = &version.ID
	p.LastEditedBy = uint64Ptr(editorID)
	p.ReviewComment = payload.ReviewComment
	if payload.Status != "" {
		p.Status = payload.Status
	}

	if err := h.DB.Save(p).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}
	p.CurrentVersion = &version
	h.writeAuditLog(c, "problem", strconv.FormatUint(p.ID, 10), "update", "created new draft version")
	utils.OK(c, adminProblemView(p, &version))
}

func (h *ProblemHandler) Publish(c *gin.Context) {
	p, err := h.loadProblem(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "题号不合法")
		return
	}
	if p == nil {
		utils.NotFound(c, "题目不存在")
		return
	}
	if p.CurrentVersionID == nil {
		utils.BadRequest(c, "题目没有可发布版本")
		return
	}

	var req publishReq
	_ = c.ShouldBindJSON(&req)
	editorID, _ := middleware.CurrentUserID(c)
	now := time.Now().UTC()

	targetVersionID := p.CurrentVersionID
	if req.VersionID > 0 {
		targetVersionID = &req.VersionID
	}
	if targetVersionID == nil {
		utils.BadRequest(c, "题目没有可发布版本")
		return
	}

	p.Status = models.ProblemStatusPublished
	p.CurrentVersionID = targetVersionID
	p.PublishedVersionID = targetVersionID
	p.PublishedAt = &now
	p.PublishedBy = uint64Ptr(editorID)
	p.ReviewComment = req.ReviewComment
	if err := h.DB.Save(p).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}

	if err := h.DB.Model(&models.ProblemVersion{}).Where("id = ?", *targetVersionID).Update("published_at", now).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}

	if err := h.DB.
		Preload("PublishedVersion.Samples").
		Preload("PublishedVersion.TestCases").
		Preload("PublishedVersion.Templates").
		First(p, p.ID).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}
	h.writeAuditLog(c, "problem", strconv.FormatUint(p.ID, 10), "publish", "published version "+strconv.FormatUint(*targetVersionID, 10))
	utils.OK(c, adminProblemView(p, chosenVersion(p)))
}

func (h *ProblemHandler) Rollback(c *gin.Context) {
	p, err := h.loadProblem(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "题号不合法")
		return
	}
	if p == nil {
		utils.NotFound(c, "题目不存在")
		return
	}

	var req rollbackReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "invalid rollback payload")
		return
	}

	var version models.ProblemVersion
	if err := h.DB.Where("id = ? AND problem_id = ?", req.VersionID, p.ID).First(&version).Error; err != nil {
		utils.NotFound(c, "目标版本不存在")
		return
	}

	editorID, _ := middleware.CurrentUserID(c)
	p.CurrentVersionID = &version.ID
	p.Title = version.Title
	p.Difficulty = version.Difficulty
	p.DifficultyScore = version.DifficultyScore
	p.Tags = version.Tags
	p.Source = version.Source
	p.LastEditedBy = uint64Ptr(editorID)
	if err := h.DB.Save(p).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}

	if err := h.DB.
		Preload("CurrentVersion.Samples").
		Preload("CurrentVersion.TestCases").
		Preload("CurrentVersion.Templates").
		Preload("Versions", func(db *gorm.DB) *gorm.DB { return db.Order("version_no DESC") }).
		First(p, p.ID).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}
	h.writeAuditLog(c, "problem", strconv.FormatUint(p.ID, 10), "rollback", "rolled back current version to "+strconv.FormatUint(version.ID, 10))
	utils.OK(c, adminProblemView(p, chosenVersion(p)))
}

func (h *ProblemHandler) Rejudge(c *gin.Context) {
	p, err := h.loadProblem(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "题号不合法")
		return
	}
	if p == nil {
		utils.NotFound(c, "题目不存在")
		return
	}

	var req rejudgeReq
	_ = c.ShouldBindJSON(&req)

	var total int64
	if err := h.DB.Model(&models.Submission{}).Where("problem_id = ? AND source = ?", p.ID, "submit").Count(&total).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}

	editorID, _ := middleware.CurrentUserID(c)
	job := models.RejudgeJob{
		ProblemID:        p.ID,
		TargetVersionID:  nilIfZero(req.TargetVersionID),
		Status:           "pending",
		Reason:           req.Reason,
		TriggeredBy:      uint64Ptr(editorID),
		TotalSubmissions: int(total),
	}
	if err := h.DB.Create(&job).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}
	h.writeAuditLog(c, "rejudge_job", strconv.FormatUint(job.ID, 10), "create", "created rejudge job for problem "+strconv.FormatUint(p.ID, 10))
	utils.OK(c, job)
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

	if err := h.DB.Transaction(func(tx *gorm.DB) error {
		var versions []models.ProblemVersion
		if err := tx.Where("problem_id = ?", p.ID).Find(&versions).Error; err != nil {
			return err
		}
		versionIDs := make([]uint64, 0, len(versions))
		for _, item := range versions {
			versionIDs = append(versionIDs, item.ID)
		}
		if len(versionIDs) > 0 {
			if err := tx.Where("version_id IN ?", versionIDs).Delete(&models.ProblemSample{}).Error; err != nil {
				return err
			}
			if err := tx.Where("version_id IN ?", versionIDs).Delete(&models.ProblemTestCase{}).Error; err != nil {
				return err
			}
			if err := tx.Where("version_id IN ?", versionIDs).Delete(&models.ProblemTemplate{}).Error; err != nil {
				return err
			}
			if err := tx.Where("problem_id = ?", p.ID).Delete(&models.ProblemVersion{}).Error; err != nil {
				return err
			}
		}
		return tx.Delete(p).Error
	}); err != nil {
		utils.Server(c, err.Error())
		return
	}
	h.writeAuditLog(c, "problem", strconv.FormatUint(p.ID, 10), "delete", "deleted problem and all versions")
	utils.OK(c, gin.H{"deleted": true, "id": p.ID})
}

func (h *ProblemHandler) loadProblem(rawID string) (*models.Problem, error) {
	id, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil {
		return nil, err
	}

	var p models.Problem
	if err := h.DB.Preload("CurrentVersion").Preload("PublishedVersion").Preload("Versions").First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (h *ProblemHandler) persistVersionChildren(versionID uint64, req problemPayload) error {
	samples := make([]models.ProblemSample, 0, len(req.Samples))
	for i, item := range req.Samples {
		samples = append(samples, models.ProblemSample{
			VersionID:   versionID,
			CaseNo:      i + 1,
			Input:       item.Input,
			Expected:    item.Expected,
			Explanation: item.Explanation,
		})
	}
	tests := make([]models.ProblemTestCase, 0, len(req.TestCases))
	for i, item := range req.TestCases {
		tests = append(tests, models.ProblemTestCase{
			VersionID: versionID,
			CaseNo:    i + 1,
			Input:     item.Input,
			Expected:  item.Expected,
			IsHidden:  item.IsHidden,
		})
	}
	templates := make([]models.ProblemTemplate, 0, len(req.Templates))
	for _, item := range req.Templates {
		if item.Language == "" || item.Code == "" {
			continue
		}
		templates = append(templates, models.ProblemTemplate{
			VersionID: versionID,
			Language:  item.Language,
			Code:      item.Code,
		})
	}
	if len(samples) > 0 {
		if err := h.DB.Create(&samples).Error; err != nil {
			return err
		}
	}
	if len(tests) > 0 {
		if err := h.DB.Create(&tests).Error; err != nil {
			return err
		}
	}
	if len(templates) > 0 {
		if err := h.DB.Create(&templates).Error; err != nil {
			return err
		}
	}
	return nil
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
	if len(req.Samples) == 0 && len(req.TestCases) > 0 {
		limit := min(2, len(req.TestCases))
		req.Samples = make([]samplePayload, 0, limit)
		for i := 0; i < limit; i++ {
			req.Samples = append(req.Samples, samplePayload{
				Input:    req.TestCases[i].Input,
				Expected: req.TestCases[i].Expected,
			})
		}
	}
	if len(req.Templates) == 0 {
		req.Templates = []templatePayload{
			{Language: "cpp", Code: "#include <bits/stdc++.h>\nusing namespace std;\n\nint main() {\n    return 0;\n}\n"},
			{Language: "python", Code: "import sys\ninput = sys.stdin.readline\n\ndef solve():\n    pass\n\nsolve()\n"},
			{Language: "go", Code: "package main\n\nfunc main() {\n}\n"},
		}
	}
	req.Status = defaultProblemStatus(req.Status)
	return true
}

func buildProblemVersion(problemID uint64, versionNo int, req problemPayload, editorID uint64, publishedAt *time.Time) models.ProblemVersion {
	return models.ProblemVersion{
		ProblemID:       problemID,
		VersionNo:       versionNo,
		Title:           req.Title,
		Difficulty:      req.Difficulty,
		DifficultyScore: req.DifficultyScore,
		Tags:            req.Tags,
		Content:         req.Content,
		Constraints:     req.Constraints,
		Source:          req.Source,
		TimeLimit:       req.TimeLimit,
		MemoryLimit:     req.MemoryLimit,
		OutputLimitKB:   req.OutputLimitKB,
		Editorial:       req.Editorial,
		CreatedBy:       uint64Ptr(editorID),
		PublishedAt:     publishedAt,
	}
}

func chosenVersion(problem *models.Problem) *models.ProblemVersion {
	if problem.CurrentVersion != nil {
		return problem.CurrentVersion
	}
	if problem.PublishedVersion != nil {
		return problem.PublishedVersion
	}
	return nil
}

func latestVersionNo(problem *models.Problem) int {
	max := 0
	for _, item := range problem.Versions {
		if item.VersionNo > max {
			max = item.VersionNo
		}
	}
	return max
}

func publicProblemView(problem *models.Problem, version *models.ProblemVersion, accepted bool) gin.H {
	view := baseProblemView(problem, version)
	view["accepted"] = accepted
	view["acceptRate"] = problem.AcceptRate()
	view["submitCount"] = problem.SubmitCount
	return view
}

func adminProblemView(problem *models.Problem, version *models.ProblemVersion) gin.H {
	view := baseProblemView(problem, version)
	view["status"] = problem.Status
	view["currentVersionId"] = problem.CurrentVersionID
	view["publishedVersionId"] = problem.PublishedVersionID
	view["reviewComment"] = problem.ReviewComment
	view["publishedAt"] = timePtrToString(problem.PublishedAt)
	view["publishedBy"] = problem.PublishedBy
	view["lastEditedBy"] = problem.LastEditedBy

	versions := make([]gin.H, 0, len(problem.Versions))
	for _, item := range problem.Versions {
		versions = append(versions, gin.H{
			"id":          item.ID,
			"versionNo":   item.VersionNo,
			"title":       item.Title,
			"difficulty":  item.Difficulty,
			"createdAt":   item.CreatedAt.UTC().Format(time.RFC3339),
			"publishedAt": timePtrToString(item.PublishedAt),
		})
	}
	view["versions"] = versions
	return view
}

func baseProblemView(problem *models.Problem, version *models.ProblemVersion) gin.H {
	view := gin.H{
		"id":              problem.ID,
		"title":           problem.Title,
		"difficulty":      problem.Difficulty,
		"difficultyScore": problem.DifficultyScore,
		"tags":            problem.Tags,
		"source":          problem.Source,
		"outputLimitKb":   problem.OutputLimitKBOrDefault(),
	}
	if version == nil {
		return view
	}

	samples := make([]gin.H, 0, len(version.Samples))
	for _, item := range version.Samples {
		samples = append(samples, gin.H{
			"caseNo":      item.CaseNo,
			"input":       item.Input,
			"expected":    item.Expected,
			"explanation": item.Explanation,
		})
	}
	tests := make([]gin.H, 0, len(version.TestCases))
	for _, item := range version.TestCases {
		tests = append(tests, gin.H{
			"caseNo":   item.CaseNo,
			"input":    item.Input,
			"expected": item.Expected,
			"isHidden": item.IsHidden,
		})
	}
	templates := make([]gin.H, 0, len(version.Templates))
	for _, item := range version.Templates {
		templates = append(templates, gin.H{
			"language": item.Language,
			"code":     item.Code,
		})
	}

	view["title"] = version.Title
	view["difficulty"] = version.Difficulty
	view["difficultyScore"] = version.DifficultyScore
	view["tags"] = version.Tags
	view["content"] = version.Content
	view["constraints"] = version.Constraints
	view["source"] = version.Source
	view["timeLimit"] = version.TimeLimit
	view["memoryLimit"] = version.MemoryLimit
	view["outputLimitKb"] = version.OutputLimitKB
	view["editorial"] = version.Editorial
	view["samples"] = samples
	view["testCases"] = tests
	view["templates"] = templates
	view["versionId"] = version.ID
	view["versionNo"] = version.VersionNo
	return view
}

func defaultProblemStatus(status string) string {
	switch status {
	case models.ProblemStatusDraft, models.ProblemStatusReview, models.ProblemStatusPublished, models.ProblemStatusArchived:
		return status
	default:
		return models.ProblemStatusDraft
	}
}

func timePtrToString(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func uint64Ptr(v uint64) *uint64 {
	if v == 0 {
		return nil
	}
	return &v
}

func nilIfZero(v uint64) *uint64 {
	if v == 0 {
		return nil
	}
	return &v
}

func (h *ProblemHandler) publishedSolutions(problemID uint64) []gin.H {
	var rows []models.ProblemSolution
	if err := h.DB.Where("problem_id = ? AND is_published = ?", problemID, true).
		Order("is_official DESC, like_count DESC, updated_at DESC").Find(&rows).Error; err != nil {
		return []gin.H{}
	}
	items := make([]gin.H, 0, len(rows))
	for _, item := range rows {
		items = append(items, gin.H{
			"id":          item.ID,
			"userId":      item.UserID,
			"username":    item.Username,
			"title":       item.Title,
			"content":     item.Content,
			"language":    item.Language,
			"isPublished": item.IsPublished,
			"isOfficial":  item.IsOfficial,
			"likeCount":   item.LikeCount,
			"updatedAt":   item.UpdatedAt.Format("2006-01-02 15:04"),
		})
	}
	return items
}

func (h *ProblemHandler) userSolution(problemID, userID uint64) gin.H {
	var item models.ProblemSolution
	if err := h.DB.Where("problem_id = ? AND user_id = ?", problemID, userID).Take(&item).Error; err != nil {
		return gin.H{}
	}
	return gin.H{
		"id":          item.ID,
		"userId":      item.UserID,
		"username":    item.Username,
		"title":       item.Title,
		"content":     item.Content,
		"language":    item.Language,
		"isPublished": item.IsPublished,
		"updatedAt":   item.UpdatedAt.Format("2006-01-02 15:04"),
	}
}

func matchesStatusFilter(filter string, item problemListItem) bool {
	switch filter {
	case "", "all":
		return true
	case "accepted":
		return item.Accepted
	case "attempted":
		return item.Attempted
	case "favorite":
		return item.Favorite
	case "unattempted":
		return !item.Attempted
	default:
		return true
	}
}

func (h *ProblemHandler) relatedProblems(problem *models.Problem, userID uint64, limit int) []gin.H {
	if problem == nil || limit <= 0 || len(problem.Tags) == 0 {
		return []gin.H{}
	}
	var rows []models.Problem
	if err := h.DB.Where("status = ? AND id <> ?", models.ProblemStatusPublished, problem.ID).Limit(limit * 3).Find(&rows).Error; err != nil {
		return []gin.H{}
	}
	favoriteTags := map[string]struct{}{}
	attemptedSet := map[uint64]bool{}
	if userID > 0 {
		type favRow struct{ Tags models.StringSlice }
		var favRows []favRow
		h.DB.Raw(`SELECT p.tags
			FROM favorites f
			JOIN problems p ON p.id = f.problem_id
			WHERE f.user_id = ?`, userID).Scan(&favRows)
		for _, row := range favRows {
			for _, tag := range row.Tags {
				favoriteTags[tag] = struct{}{}
			}
		}
		var attempted []uint64
		h.DB.Model(&models.Submission{}).Where("user_id = ? AND source = ?", userID, "submit").Distinct("problem_id").Pluck("problem_id", &attempted)
		for _, id := range attempted {
			attemptedSet[id] = true
		}
	}
	type ranked struct {
		item  models.Problem
		score int
	}
	rankedRows := make([]ranked, 0, len(rows))
	for _, item := range rows {
		candidate := item
		score := scoreRelatedProblem(problem, &candidate, favoriteTags, attemptedSet[item.ID])
		if score <= 0 {
			continue
		}
		rankedRows = append(rankedRows, ranked{item: item, score: score})
	}
	slices.SortFunc(rankedRows, func(a, b ranked) int {
		if a.score == b.score {
			if a.item.ID < b.item.ID {
				return -1
			}
			if a.item.ID > b.item.ID {
				return 1
			}
			return 0
		}
		if a.score > b.score {
			return -1
		}
		return 1
	})
	items := make([]gin.H, 0, min(limit, len(rankedRows)))
	for _, row := range rankedRows {
		if row.item.ID == problem.ID {
			continue
		}
		items = append(items, gin.H{
			"id":         row.item.ID,
			"title":      row.item.Title,
			"difficulty": row.item.Difficulty,
			"tags":       row.item.Tags,
			"score":      row.score,
		})
		if len(items) >= limit {
			break
		}
	}
	return items
}

func sharesAnyTag(a, b models.StringSlice) bool {
	set := map[string]struct{}{}
	for _, item := range a {
		set[item] = struct{}{}
	}
	for _, item := range b {
		if _, ok := set[item]; ok {
			return true
		}
	}
	return false
}

func scoreRelatedProblem(base, candidate *models.Problem, favoriteTags map[string]struct{}, attempted bool) int {
	score := 0
	for _, tag := range candidate.Tags {
		for _, baseTag := range base.Tags {
			if tag == baseTag {
				score += 3
			}
		}
		if _, ok := favoriteTags[tag]; ok {
			score += 2
		}
	}
	if candidate.Difficulty == base.Difficulty {
		score += 2
	}
	if attempted {
		score += 1
	}
	return score
}

func (h *ProblemHandler) writeAuditLog(c *gin.Context, resourceType, resourceID, action, detail string) {
	uid, ok := middleware.CurrentUserID(c)
	var userID *uint64
	if ok {
		userID = &uid
	}
	log := models.AuditLog{
		UserID:       userID,
		Username:     middleware.CurrentUsername(c),
		UserRole:     middleware.CurrentUserRole(c),
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Action:       action,
		Detail:       detail,
	}
	if err := h.DB.Create(&log).Error; err != nil {
		// audit failure should not break the main mutation path
		_ = err
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
