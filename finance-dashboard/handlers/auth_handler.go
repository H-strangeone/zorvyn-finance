package handlers

import (
	"finance-dashboard/services"
	"finance-dashboard/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var input services.RegisterInput




	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendError(c, utils.NewValidationError("invalid request body"))
		return
	}

	response, appErr := h.authService.Register(input)
	if appErr != nil {
		utils.SendError(c, appErr)
		return
	}

	utils.SendSuccess(c, http.StatusCreated, "registration successful", response)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input services.LoginInput

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendError(c, utils.NewValidationError("invalid request body"))
		return
	}

	response, appErr := h.authService.Login(input)
	if appErr != nil {
		utils.SendError(c, appErr)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "login successful", response)
}
