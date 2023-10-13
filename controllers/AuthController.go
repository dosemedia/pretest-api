package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dosemedia/pretest-api/prisma/db"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	goredislib "github.com/redis/go-redis/v9"
)

type AuthController struct {
	client *db.PrismaClient
	cache  *goredislib.Client
}

func NewAuthController(e *echo.Echo) *AuthController {
	redisAddress := os.Getenv("REDIS_ADDRESS")
	if redisAddress == "" {
		redisAddress = "localhost:6379"
	}
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB := 0

	controller := AuthController{
		client: db.NewClient(),
		cache: goredislib.NewClient(&goredislib.Options{
			Addr:     redisAddress,
			Password: redisPassword,
			DB:       redisDB,
		}),
	}
	return &controller
}

func FlushRedisPrefix(client *goredislib.Client, prefix string) error {
	keys, err := client.Keys(context.Background(), prefix+"*").Result()
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		_, err := client.Del(context.Background(), keys...).Result()
		if err != nil {
			return err
		}
	}
	return nil
}

func authToken(controller *AuthController, token string, c echo.Context) error {
	if token == "" {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"X-Hasura-Role": "public",
		})
	}

	// Parse the token
	parsed, tokenError := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		keyStr := os.Getenv("JWT_TOKEN_KEY")
		if keyStr == "" {
			return nil, fmt.Errorf("backend not configured - missing jwt token key")
		}
		key := []byte(keyStr)
		return key, nil
	})
	if tokenError != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": tokenError.Error(),
		})
	}

	if !parsed.Valid {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"message": "Invalid token.",
		})
	}

	claims := parsed.Claims.(jwt.MapClaims)
	userId := claims["user_id"].(string)

	if userId == "" {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"message": "Invalid token.",
		})
	}

	cacheKey := userId + ":" + token

	// Check for cached value in redis
	cachedResponseStr, cachedResponseError := controller.cache.Get(context.Background(), cacheKey).Bytes()
	if cachedResponseError == nil {
		cachedResponse := map[string]interface{}{}
		cachedResponseUnmarshalError := json.Unmarshal(cachedResponseStr, &cachedResponse)
		if cachedResponseUnmarshalError == nil {
			return c.JSON(http.StatusOK, cachedResponse)
		}
	}

	// No cached response for token, check if user exists
	user, userErr := controller.client.Users.FindUnique(db.Users.ID.Equals(userId)).Exec(context.Background())
	if userErr != nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"message": "User not found.",
		})
	}

	// Use password timestamp to invalidate tokens when password changes
	// Note, this depends on user's cached tokens getting removed on password changes
	passwordAt := claims["password_at"].(string)
	if user.PasswordAt.String() != passwordAt {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"message": "Token expired.",
		})
	}

	role := "user"
	roles := []string{role}

	if strings.Contains(user.Email, "@orchard-insights.com") {
		role = "superuser"
		roles = append(roles, "superuser")
	}

	response := map[string]interface{}{
		"X-Hasura-Role":          role,
		"X-Hasura-User-Id":       userId,
		"X-Hasura-Allowed-Roles": roles,
	}

	// Cache response
	responseJson, _ := json.Marshal(response)
	redisErr := controller.cache.Set(context.Background(), cacheKey, responseJson, time.Duration(time.Minute*60)).Err()
	if redisErr != nil {
		log.Print(redisErr)
	}

	return c.JSON(http.StatusOK, response)
}

func (controller *AuthController) Run(e *echo.Echo) error {
	if err := controller.client.Prisma.Connect(); err != nil {
		return err
	}

	// Check cache connection
	pingCmd := controller.cache.Ping(context.Background())
	pingErr := pingCmd.Err()
	if pingErr != nil {
		return pingErr
	}

	e.POST("/hasura/auth", func(c echo.Context) error {
		token := c.Request().Header.Get("Authorization")

		if token == "" {
			body := map[string]interface{}{}
			bodyErr := json.NewDecoder(c.Request().Body).Decode(&body)
			if bodyErr == nil {
				token = body["token"].(string)
			}
		}

		if strings.HasPrefix(token, "Bearer ") {
			token = strings.Replace(token, "Bearer ", "", 1)
		}

		return authToken(controller, token, c)
	})

	e.GET("/hasura/auth", func(c echo.Context) error {
		token := c.Request().Header.Get("Authorization")

		if token == "" {
			token = c.QueryParam("token")
		}

		if strings.HasPrefix(token, "Bearer ") {
			token = strings.Replace(token, "Bearer ", "", 1)
		}

		return authToken(controller, token, c)
	})

	return nil
}

func (controller *AuthController) Shutdown() {
	log.Print("Shutting down auth controller...")
	if controller.client != nil {
		controller.client.Disconnect()
	}
}
