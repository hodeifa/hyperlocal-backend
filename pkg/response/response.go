// Package response provides helper functions to format standard JSON HTTP responses.
package response

import (
	"math"

	"github.com/gin-gonic/gin"
)

// Response defines the standard structure for JSON responses.
//
//nolint:govet // fieldalignment is ignored for readability
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// Meta berisi informasi pagination untuk endpoint list.
type Meta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalData  int64 `json:"total_data"`
	TotalPages int   `json:"total_pages"`
}

// Success mengirimkan respons sukses standar.
func Success(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Error mengirimkan respons error standar.
func Error(c *gin.Context, statusCode int, message, err string) {
	c.JSON(statusCode, Response{
		Success: false,
		Message: message,
		Error:   err,
	})
}

// Paginated mengirimkan respons sukses dengan metadata pagination.
func Paginated(c *gin.Context, statusCode int, message string, data interface{}, page, limit int, totalData int64) {
	totalPages := 0
	if limit > 0 {
		totalPages = int(math.Ceil(float64(totalData) / float64(limit)))
	}

	c.JSON(statusCode, Response{
		Success: true,
		Message: message,
		Data:    data,
		Meta: &Meta{
			Page:       page,
			Limit:      limit,
			TotalData:  totalData,
			TotalPages: totalPages,
		},
	})
}
