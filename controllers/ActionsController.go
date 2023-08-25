package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"aaronblondeau.com/hasura-base-go/prisma/db"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type RegisterBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ActionsController struct {
	client *db.PrismaClient
}

func NewActionsController(e *echo.Echo) *ActionsController {
	controller := ActionsController{
		client: db.NewClient(),
	}
	return &controller
}

func generateTokenForUser(user db.UsersModel) (string, error) {
	keyStr := os.Getenv("JWT_TOKEN_KEY")
	if keyStr == "" {
		return "", fmt.Errorf("backend not configured - missing jwt token key")
	}
	key := []byte(keyStr)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":     user.ID,
		"email":       user.Email,
		"password_at": user.PasswordAt.String(),
	})
	return token.SignedString(key)
}

func (controller *ActionsController) Run(e *echo.Echo) error {

	if err := controller.client.Prisma.Connect(); err != nil {
		return err
	}

	e.POST("/hasura/actions/register", func(c echo.Context) error {
		// Body should contain email
		body := RegisterBody{}
		bodyErr := json.NewDecoder(c.Request().Body).Decode(&body)
		if bodyErr != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": "email and password are required.",
			})
		}

		// Always handle emails in lowercase on the backend
		email := strings.ToLower(body.Email)
		if email == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": "Email is required.",
			})
		}

		password := body.Password
		if len(password) < 5 {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": "Password must be at least 5 characters long.",
			})
		}

		// Check for existing user so we can send a nicer error message than unique key constraint
		_, existingError := controller.client.Users.FindUnique(db.Users.Email.Equals(email)).Exec(context.Background())
		if existingError == nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": fmt.Sprintf("User with email %v already exists.", email),
			})
		}

		hashedPassword, hashError := bcrypt.GenerateFromPassword([]byte(password), 10)
		if hashError != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"message": hashError.Error(),
			})
		}

		createdUser, createErr := controller.client.Users.CreateOne(db.Users.Email.Set(email), db.Users.HashedPassword.Set(string(hashedPassword))).Exec(context.Background())
		if createErr != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"message": createErr.Error(),
			})
		}

		// create auth token for user
		token, tokenErr := generateTokenForUser(*createdUser)
		if tokenErr != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"message": tokenErr.Error(),
			})
		}

		// TODO - send verification/welcome email

		return c.JSON(http.StatusOK, map[string]interface{}{
			"token": token,
			"id":    createdUser.ID,
		})
	})

	e.POST("/hasura/actions/login", func(c echo.Context) error {
		// Body should contain email
		body := LoginBody{}
		bodyErr := json.NewDecoder(c.Request().Body).Decode(&body)
		if bodyErr != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": "email and password are required.",
			})
		}

		// Always handle emails in lowercase on the backend
		email := strings.ToLower(body.Email)
		if email == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": "Email is required.",
			})
		}

		password := body.Password
		if password == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": "Password is required.",
			})
		}

		user, findUserError := controller.client.Users.FindUnique(db.Users.Email.Equals(email)).Exec(context.Background())
		if findUserError != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": "User not found.",
			})
		}

		// create auth token for user
		token, tokenErr := generateTokenForUser(*user)
		if tokenErr != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"message": tokenErr.Error(),
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"token": token,
			"id":    user.ID,
		})
	})
	return nil
}

func (controller *ActionsController) Shutdown() {
	log.Print("Shutting down actions controller...")
	if controller.client != nil {
		controller.client.Disconnect()
	}
}
