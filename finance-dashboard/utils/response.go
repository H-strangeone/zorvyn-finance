package utils

import "github.com/gin-gonic/gin"

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Error   *ErrorBody  `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Details string `json:"details"`
}

func SendSuccess(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
		Error:   nil,
	})
}

func SendError(c *gin.Context, err *AppError) {
	c.JSON(err.StatusCode, APIResponse{
		Success: false,
		Message: err.Message,  // short: "Validation failed"
		Data:    nil,
		Error: &ErrorBody{
			Code:    err.Code,
			Details: err.Details, // specific: "amount must be greater than 0"
		},
	})
}

func SendInternalError(c *gin.Context) {
	SendError(c, NewInternalError("please try again later"))
}