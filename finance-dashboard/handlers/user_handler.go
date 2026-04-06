package handlers

import (
	"finance-dashboard/middleware"
	"finance-dashboard/services"
	"finance-dashboard/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetAll(c *gin.Context) {
	users := h.userService.GetAll()
	utils.SendSuccess(c, http.StatusOK, "users fetched successfully", users)
}

func (h *UserHandler) GetByID(c *gin.Context) {
	id := c.Param("id")





	if !utils.IsValidUUID(id) {
		utils.SendError(c, utils.NewValidationError("invalid user ID format"))
		return
	}

	user, appErr := h.userService.GetByID(id)
	if appErr != nil {
		utils.SendError(c, appErr)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "user fetched successfully", user)
}

func (h *UserHandler) GetMe(c *gin.Context) {

	userID := middleware.GetUserID(c)

	user, appErr := h.userService.GetByID(userID)
	if appErr != nil {
		utils.SendError(c, appErr)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "profile fetched successfully", user)
}

func (h *UserHandler) Create(c *gin.Context) {
	var input services.CreateUserInput

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendError(c, utils.NewValidationError("invalid request body"))
		return
	}

	user, appErr := h.userService.Create(input)
	if appErr != nil {
		utils.SendError(c, appErr)
		return
	}

	utils.SendSuccess(c, http.StatusCreated, "user created successfully", user)
}

func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")

	if !utils.IsValidUUID(id) {
		utils.SendError(c, utils.NewValidationError("invalid user ID format"))
		return
	}

	var input services.UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendError(c, utils.NewValidationError("invalid request body"))
		return
	}

	user, appErr := h.userService.Update(id, input)
	if appErr != nil {
		utils.SendError(c, appErr)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "user updated successfully", user)
}

func (h *UserHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	requestingAdminID := middleware.GetUserID(c)
	if !utils.IsValidUUID(id) {
		utils.SendError(c, utils.NewValidationError("invalid user ID format"))
		return
	}




	var body struct {
		IsActive bool `json:"isActive"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.SendError(c, utils.NewValidationError("invalid request body"))
		return
	}

	user, appErr := h.userService.UpdateStatus(id, requestingAdminID, body.IsActive)
	if appErr != nil {
		utils.SendError(c, appErr)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "user status updated successfully", user)
}
