package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Envelope is the unified {code, message, data} response body.
type Envelope struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// OK writes a successful response.
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Envelope{Code: 0, Message: "ok", Data: data})
}

// Fail writes a non-success response with a custom HTTP status code.
func Fail(c *gin.Context, httpStatus int, code int, msg string) {
	c.AbortWithStatusJSON(httpStatus, Envelope{Code: code, Message: msg, Data: nil})
}

// BadRequest is a shortcut for 400.
func BadRequest(c *gin.Context, msg string) { Fail(c, http.StatusBadRequest, -1, msg) }

// Unauthorized is a shortcut for 401.
func Unauthorized(c *gin.Context, msg string) { Fail(c, http.StatusUnauthorized, -1, msg) }

// Forbidden is a shortcut for 403.
func Forbidden(c *gin.Context, msg string) { Fail(c, http.StatusForbidden, -1, msg) }

// NotFound is a shortcut for 404.
func NotFound(c *gin.Context, msg string) { Fail(c, http.StatusNotFound, -1, msg) }

// Server is a shortcut for 500.
func Server(c *gin.Context, msg string) { Fail(c, http.StatusInternalServerError, -1, msg) }

// TooManyRequests is a shortcut for 429.
func TooManyRequests(c *gin.Context, msg string) { Fail(c, http.StatusTooManyRequests, -1, msg) }
