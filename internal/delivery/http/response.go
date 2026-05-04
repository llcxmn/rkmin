package http

import (
	"errors"
	nethttp "net/http"

	"rkmin/internal/repository"
	"rkmin/internal/usecase"

	"github.com/gin-gonic/gin"
)

type response struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Errors  any    `json:"errors"`
	Data    any    `json:"data"`
}

func ok(c *gin.Context, method string, data any) {
	c.JSON(nethttp.StatusOK, response{
		Status:  true,
		Message: "Succeed to " + method + " data",
		Errors:  nil,
		Data:    data,
	})
}

func okMessage(c *gin.Context, message string, data any) {
	c.JSON(nethttp.StatusOK, response{Status: true, Message: message, Errors: nil, Data: data})
}

func fail(c *gin.Context, status int, method string, err error) {
	msg := "Failed to " + method + " data"
	if errors.Is(err, repository.ErrNotFound) {
		status = nethttp.StatusNotFound
	}
	if errors.Is(err, usecase.ErrUnauthorized) {
		status = nethttp.StatusUnauthorized
	}
	if errors.Is(err, usecase.ErrForbidden) {
		status = nethttp.StatusForbidden
	}
	c.JSON(status, response{
		Status:  false,
		Message: msg,
		Errors:  []string{err.Error()},
		Data:    nil,
	})
}
