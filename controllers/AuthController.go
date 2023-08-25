package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"aaronblondeau.com/hasura-base-go/prisma/db"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

type AuthController struct {
	client *db.PrismaClient
}

func NewAuthController(e *echo.Echo) *AuthController {
	controller := AuthController{
		client: db.NewClient(),
	}
	return &controller
}

func (controller *AuthController) Run(e *echo.Echo) error {
	if err := controller.client.Prisma.Connect(); err != nil {
		return err
	}

	// e.POST("/hasura/auth", func(c echo.Context) error {

	// }

	e.GET("/hasura/auth", func(c echo.Context) error {
		token := c.Request().Header.Get("Authorization")

		if token == "" {
			token = c.QueryParam("token")
		}

		if strings.HasPrefix(token, "Bearer ") {
			token = strings.Replace(token, "Bearer ", "", 1)
		}

		// TODO check for cached value in redis

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

		// Check if user exists
		user, userErr := controller.client.Users.FindUnique(db.Users.ID.Equals(userId)).Exec(context.Background())
		if userErr != nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"message": "User not found.",
			})
		}

		// Use password timestamp to invalidate tokens when password changes
		// Note, this depends on user's cached tokens getting removed on password changes
		passwordAt := claims["password_at"].(string)
		fmt.Println(user.PasswordAt.String(), passwordAt)
		if user.PasswordAt.String() != passwordAt {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"message": "Token expired.",
			})
		}

		role := "user"

		// TODO - cache the response in redis

		return c.JSON(http.StatusOK, map[string]interface{}{
			"X-Hasura-Role":    role,
			"X-Hasura-User-Id": userId,
		})
	})

	return nil
}

func (controller *AuthController) Shutdown() {
	log.Print("Shutting down auth controller...")
	if controller.client != nil {
		controller.client.Disconnect()
	}
}
