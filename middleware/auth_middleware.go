package middleware

import (
	"fmt"
	"github.com/dastardlyjockey/restaurant-management-backend/helpers"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Authentication(c *gin.Context) {
	token := c.Request.Header.Get("token")

	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("no authorization provided")})
		c.Abort()
		return
	}

	//validate the token
	claims, err := helpers.ValidateToken(token)
	if err != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		c.Abort()
		return
	}

	c.Set("email", claims.Email)
	c.Set("first_name", claims.FirstName)
	c.Set("last_name", claims.LastName)
	c.Set("uid", claims.Uid)

	c.Next()
}
