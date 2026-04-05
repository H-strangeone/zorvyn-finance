package handlers

import (
	"finance-dashboard/services"
	"finance-dashboard/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	dashboardService *services.DashboardService
}

func NewDashboardHandler(dashboardService *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboardService: dashboardService}
}

func (h *DashboardHandler) GetSummary(c *gin.Context) {
	summary := h.dashboardService.GetSummary()
	utils.SendSuccess(c, http.StatusOK, "summary fetched successfully", summary)
}

func (h *DashboardHandler) GetByCategory(c *gin.Context) {
	breakdown := h.dashboardService.GetByCategory()
	utils.SendSuccess(c, http.StatusOK, "category breakdown fetched successfully", breakdown)
}

func (h *DashboardHandler) GetTrends(c *gin.Context) {
	trends := h.dashboardService.GetTrends()
	utils.SendSuccess(c, http.StatusOK, "trends fetched successfully", trends)
}

func (h *DashboardHandler) GetRecent(c *gin.Context) {

	limit := 5
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	recent := h.dashboardService.GetRecent(limit)
	utils.SendSuccess(c, http.StatusOK, "recent transactions fetched successfully", recent)
}
