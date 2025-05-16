package api

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"context"
	"net/http/httptest"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joekingsleyMukundi/Gatekeeper/internals"
	"github.com/joekingsleyMukundi/Gatekeeper/middlewares"
	"github.com/redis/go-redis/v9"
)

func TestForgotPasswordRateLimitMiddleware(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})
	ctx := context.Background()
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		t.Fatalf("Failed to connect to Redis: %v", err)
	}
	if err := rdb.FlushDB(ctx).Err(); err != nil {
		t.Fatalf("Failed to flush Redis DB: %v", err)
	}
	ipLimiter := internals.NewSlidingWindowLimiter(rdb, 10, 15*time.Minute)
	emailLimiter := internals.NewSlidingWindowLimiter(rdb, 3, 30*time.Minute)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Request.RemoteAddr = "192.168.1.1:12345"
		c.Next()
	})

	router.POST("/auth/forgot-password", middlewares.ForgotPasswordRateLimitMiddleware(ipLimiter, emailLimiter), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Password reset request accepted"})
	})
	server := httptest.NewServer(router)
	defer server.Close()
	email := "test@example.com"
	reqBody := `{"email": "` + email + `"}`
	emailKey := "forgot_password:email:" + email
	for i := 1; i <= 3; i++ {
		resp, err := http.Post(
			server.URL+"/auth/forgot-password",
			"application/json",
			strings.NewReader(reqBody),
		)
		if err != nil {
			t.Fatalf("Failed to make request %d: %v", i, err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Request %d should succeed, got status %d", i, resp.StatusCode)
		}
		resp.Body.Close()
		count, err := rdb.ZCard(ctx, emailKey).Result()
		if err != nil {
			t.Errorf("Failed to check Redis key: %v", err)
		}
		t.Logf("After request %d, count for key %s: %d", i, emailKey, count)
		time.Sleep(50 * time.Millisecond)
	}
	resp, err := http.Post(
		server.URL+"/auth/forgot-password",
		"application/json",
		strings.NewReader(reqBody),
	)
	if err != nil {
		t.Fatalf("Failed to make rate-limited request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", resp.StatusCode)
	}
	var respData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		t.Errorf("Failed to decode response JSON: %v", err)
	}
	if _, ok := respData["error"]; !ok {
		t.Errorf("Response should contain error field: %v", respData)
	}
	if _, ok := respData["retry_after"]; !ok {
		t.Errorf("Response should contain retry_after field: %v", respData)
	}
	count, err := rdb.ZCard(ctx, emailKey).Result()
	if err != nil {
		t.Errorf("Failed to check Redis key: %v", err)
	}
	t.Logf("After rate-limited request, count for key %s: %d", emailKey, count)
	if err := rdb.FlushDB(ctx).Err(); err != nil {
		t.Fatalf("Failed to flush Redis DB after test: %v", err)
	}
}
