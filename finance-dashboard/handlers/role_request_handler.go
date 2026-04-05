package handlers

import (
	"finance-dashboard/middleware"
	"finance-dashboard/services"
	"finance-dashboard/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RoleRequestHandler struct {
	roleRequestService *services.RoleRequestService
}

func NewRoleRequestHandler(roleRequestService *services.RoleRequestService) *RoleRequestHandler {
	return &RoleRequestHandler{roleRequestService: roleRequestService}
}

func (h *RoleRequestHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var input services.CreateRoleRequestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendError(c, utils.NewValidationError("invalid request body"))
		return
	}

	req, appErr := h.roleRequestService.Create(userID, input)
	if appErr != nil {
		utils.SendError(c, appErr)
		return
	}

	utils.SendSuccess(c, http.StatusCreated, "role request submitted successfully", req)
}

func (h *RoleRequestHandler) GetAll(c *gin.Context) {

	status := c.Query("status")

	requests, appErr := h.roleRequestService.GetAll(status)
	if appErr != nil {
		utils.SendError(c, appErr)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "role requests fetched successfully", requests)
}

func (h *RoleRequestHandler) GetMine(c *gin.Context) {
	userID := middleware.GetUserID(c)

	requests, appErr := h.roleRequestService.GetMine(userID)
	if appErr != nil {
		utils.SendError(c, appErr)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "your role requests fetched successfully", requests)
}

func (h *RoleRequestHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	if !utils.IsValidUUID(id) {
		utils.SendError(c, utils.NewValidationError("invalid request ID format"))
		return
	}

	req, appErr := h.roleRequestService.GetByID(id)
	if appErr != nil {
		utils.SendError(c, appErr)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "role request fetched successfully", req)
}

func (h *RoleRequestHandler) Process(c *gin.Context) {
	id := c.Param("id")
	adminID := middleware.GetUserID(c)

	if !utils.IsValidUUID(id) {
		utils.SendError(c, utils.NewValidationError("invalid request ID format"))
		return
	}

	var input services.ProcessRoleRequestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendError(c, utils.NewValidationError("invalid request body"))
		return
	}

	req, appErr := h.roleRequestService.Process(id, adminID, input)
	if appErr != nil {
		utils.SendError(c, appErr)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "role request processed successfully", req)
}
