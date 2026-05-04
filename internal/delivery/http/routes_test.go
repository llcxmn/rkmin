package http

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterRoutesDoesNotPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHandler(nil, nil, nil)
	h.RegisterRoutes(r, "/api/v1")
}
