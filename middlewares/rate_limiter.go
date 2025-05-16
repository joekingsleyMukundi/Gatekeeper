package middlewares

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joekingsleyMukundi/Gatekeeper/internals"
)

type forgortPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

func ForgotPasswordRateLimitMiddleware(ipLimiter, emailLimiter internals.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		ip := c.ClientIP()
		ipKey := "forgot_password:ip:" + ip
		ipAllowed, ipRetry, err := ipLimiter.Allow(ctx, ipKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Rate limiting error (IP)",
			})
			return
		}
		fmt.Println("IP Rate Limit Allowed:", ipAllowed)
		if !ipAllowed {
			c.Header("Retry-After", ipRetry.String())
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many requests from this IP",
				"retry_after": ipRetry.Seconds(),
			})
			return
		}
		var req forgortPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			c.Abort()
			return
		}
		emailKey := "forgot_password:email:" + req.Email
		emailAllowed, emailRetry, err := emailLimiter.Allow(ctx, emailKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Rate limiting error (email)",
			})
			return
		}
		fmt.Println("Email Rate Limit Allowed:", emailAllowed)
		if !emailAllowed {
			c.Header("Retry-After", emailRetry.String())
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many password reset requests for this email",
				"retry_after": emailRetry.Seconds(),
			})
			return
		}
		c.Next()
	}
}
