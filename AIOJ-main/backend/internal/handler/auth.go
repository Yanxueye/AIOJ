package handler

import (
	"errors"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/terminaloj/backend/internal/models"
	"github.com/terminaloj/backend/internal/utils"
	"gorm.io/gorm"
)

type AuthHandler struct {
	DB  *gorm.DB
	JWT *utils.JWTManager
}

type loginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type registerReq struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

var emailRe = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "用户名和密码不能为空")
		return
	}
	var user models.User
	if err := h.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.BadRequest(c, "用户名或密码错误")
			return
		}
		utils.Server(c, err.Error())
		return
	}
	if !utils.CheckPassword(user.PasswordHash, req.Password) {
		utils.BadRequest(c, "用户名或密码错误")
		return
	}
	token, err := h.JWT.Sign(user.ID, user.Username)
	if err != nil {
		utils.Server(c, err.Error())
		return
	}
	profile := buildProfile(h.DB, &user, false)
	utils.OK(c, gin.H{"token": token, "user": profile})
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数不合法")
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	if n := len(req.Username); n < 3 || n > 20 {
		utils.BadRequest(c, "用户名长度需在 3-20 字符之间")
		return
	}
	if !emailRe.MatchString(req.Email) {
		utils.BadRequest(c, "邮箱格式不合法")
		return
	}
	if len(req.Password) < 6 {
		utils.BadRequest(c, "密码至少 6 位")
		return
	}

	var cnt int64
	h.DB.Model(&models.User{}).
		Where("username = ? OR email = ?", req.Username, req.Email).Count(&cnt)
	if cnt > 0 {
		utils.BadRequest(c, "用户名或邮箱已被占用")
		return
	}
	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.Server(c, err.Error())
		return
	}
	user := models.User{Username: req.Username, Email: req.Email, PasswordHash: hash, Rating: 1200}
	if err := h.DB.Create(&user).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}
	utils.OK(c, gin.H{"message": "注册成功"})
}
