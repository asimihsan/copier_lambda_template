package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type TokenHandler struct{}

func NewTokenHandler() *TokenHandler {
	return &TokenHandler{}
}

func (h *TokenHandler) IssueTokenHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"token": "fake-jwt-token",
	})
}
