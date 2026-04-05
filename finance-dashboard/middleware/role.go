package middleware

import (
	"finance-dashboard/models"
	"finance-dashboard/utils"

	"github.com/gin-gonic/gin"
)


func RequireRole(allowedRoles ...models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {

		roleStr := GetRole(c)
		if roleStr == "" {
		
			utils.SendError(c, utils.NewUnauthorizedError("authorization required"))
			c.Abort()
			return
		}

		currentRole := models.Role(roleStr)

		for _, allowed := range allowedRoles {
			if currentRole == allowed {
				c.Next() // role matched — continue
				return
			}
		}

		
		utils.SendError(c, utils.NewForbiddenError(
			string(currentRole)+" role does not have access to this resource",
		))
		c.Abort()
	}
}



func RequireAdmin() gin.HandlerFunc {
	return RequireRole(models.RoleAdmin)
}

func RequireAnalystOrAdmin() gin.HandlerFunc {
	return RequireRole(models.RoleAdmin, models.RoleAnalyst)
}

func RequireAnyRole() gin.HandlerFunc {
	return RequireRole(models.RoleAdmin, models.RoleAnalyst, models.RoleViewer)
}