package handlers

import (
	"finance-dashboard/middleware"
	"finance-dashboard/models"
	"finance-dashboard/services"
	"finance-dashboard/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type TransactionHandler struct {
	transactionService *services.TransactionService
}

func NewTransactionHandler(transactionService *services.TransactionService) *TransactionHandler {
	return &TransactionHandler{transactionService: transactionService}
}

func (h *TransactionHandler) GetAll(c *gin.Context) {

	filter := models.TransactionFilter{}

	filter.Type = c.Query("type")
	filter.Category = c.Query("category")
	filter.Search = c.Query("search")
	filter.Sort = c.Query("sort")

	fromStr := c.Query("from")
	if fromStr != "" {
		if t, err := time.Parse("2006-01-02", fromStr); err == nil {
			filter.From = t.UTC()
		} else {
			utils.SendError(c, utils.NewValidationError("invalid from date, use YYYY-MM-DD"))
			return
		}
	}
	toStr := c.Query("to")
	if toStr != "" {
		if t, err := time.Parse("2006-01-02", toStr); err == nil {
			filter.To = t.UTC()
		} else {
			utils.SendError(c, utils.NewValidationError("invalid to date, use YYYY-MM-DD"))
			return
		}
	}
	if fromStr != "" && toStr != "" {
		if filter.From.After(filter.To) {
			utils.SendError(c, utils.NewValidationError("from date must be before to date"))
			return
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}

	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}

	result, appErr := h.transactionService.GetAll(filter)
	if appErr != nil {
		utils.SendError(c, appErr)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "transactions fetched successfully", result)
}

func (h *TransactionHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	if !utils.IsValidUUID(id) {
		utils.SendError(c, utils.NewValidationError("invalid transaction ID format"))
		return
	}

	tx, appErr := h.transactionService.GetByID(id)
	if appErr != nil {
		utils.SendError(c, appErr)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "transaction fetched successfully", tx)
}

func (h *TransactionHandler) Create(c *gin.Context) {

	userID := middleware.GetUserID(c)

	var input services.CreateTransactionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendError(c, utils.NewValidationError("invalid request body"))
		return
	}

	tx, appErr := h.transactionService.Create(userID, input)
	if appErr != nil {
		utils.SendError(c, appErr)
		return
	}

	utils.SendSuccess(c, http.StatusCreated, "transaction created successfully", tx)
}

func (h *TransactionHandler) Update(c *gin.Context) {
	id := c.Param("id")

	if !utils.IsValidUUID(id) {
		utils.SendError(c, utils.NewValidationError("invalid transaction ID format"))
		return
	}

	var input services.UpdateTransactionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendError(c, utils.NewValidationError("invalid request body"))
		return
	}

	tx, appErr := h.transactionService.Update(id, input)
	if appErr != nil {
		utils.SendError(c, appErr)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "transaction updated successfully", tx)
}

func (h *TransactionHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if !utils.IsValidUUID(id) {
		utils.SendError(c, utils.NewValidationError("invalid transaction ID format"))
		return
	}

	if appErr := h.transactionService.Delete(id); appErr != nil {
		utils.SendError(c, appErr)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "transaction deleted successfully", nil)
}
