package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"aaronblondeau.com/hasura-base-go/prisma/db"
	"github.com/aaronblondeau/crew-go/crew"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type RegisterBodyInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterBody struct {
	Input RegisterBodyInput `json:"input"`
}

type LoginBodyInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginBody struct {
	Input LoginBodyInput `json:"input"`
}

type VerifyEmailBody struct {
	Input VerifyEmailBodyInput `json:"input"`
}

type VerifyEmailBodyInput struct {
	Code string `json:"code"`
}

type UserSessionBody struct {
	SessionVariables struct {
		XHasuraUserId string `json:"x-hasura-user-id"`
	} `json:"session_variables"`
}

type ActionsController struct {
	client        *db.PrismaClient
	crewContoller *crew.TaskController
}

func NewActionsController(e *echo.Echo, crewController *crew.TaskController) *ActionsController {
	controller := ActionsController{
		client:        db.NewClient(),
		crewContoller: crewController,
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

// NOTE : requests sent with hasura console need x-hasura-admin-secret request header unchecked.
// If x-hasura-admin-secret is checked, only x-hasura-role=admin var is sent in body.
func getUserForRequest(c echo.Context, client *db.PrismaClient) (*db.UsersModel, error) {
	body := UserSessionBody{}
	bodyErr := json.NewDecoder(c.Request().Body).Decode(&body)
	if bodyErr != nil {
		return nil, bodyErr
	}
	user, findUserError := client.Users.FindUnique(db.Users.ID.Equals(body.SessionVariables.XHasuraUserId)).Exec(c.Request().Context())
	if findUserError != nil {
		return nil, findUserError
	}
	return user, nil
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
		email := strings.ToLower(body.Input.Email)
		if email == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": "Email is required.",
			})
		}

		password := body.Input.Password
		if len(password) < 5 {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": "Password must be at least 5 characters long.",
			})
		}

		// Check for existing user so we can send a nicer error message than unique key constraint
		_, existingError := controller.client.Users.FindUnique(db.Users.Email.Equals(email)).Exec(c.Request().Context())
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

		createdUser, createErr := controller.client.Users.CreateOne(db.Users.Email.Set(email), db.Users.HashedPassword.Set(string(hashedPassword))).Exec(c.Request().Context())
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

		// Send verification/welcome email via backend task
		group := crew.NewTaskGroup("", "Welcome User "+createdUser.ID)
		taskGroupCreateError := controller.crewContoller.CreateTaskGroup(group)
		if taskGroupCreateError != nil {
			log.Print(taskGroupCreateError)
		}

		task := crew.NewTask()
		task.TaskGroupId = group.Id
		task.Name = "Send Verification Email"
		task.Worker = "verify-email"
		task.Input = VerifyEmailJobInput{
			UserId: createdUser.ID,
		}
		taskCreateError := controller.crewContoller.CreateTask(task)
		if taskCreateError != nil {
			log.Print(taskCreateError)
		}

		// Return result of registration action
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
		email := strings.ToLower(body.Input.Email)
		if email == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": "Email is required.",
			})
		}

		password := body.Input.Password
		if password == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": "Password is required.",
			})
		}

		user, findUserError := controller.client.Users.FindUnique(db.Users.Email.Equals(email)).Exec(c.Request().Context())
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

	e.POST("/hasura/actions/resendVerificationEmail", func(c echo.Context) error {
		user, userError := getUserForRequest(c, controller.client)
		if userError != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": userError.Error(),
			})
		}
		if user == nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": "User not found.",
			})
		}

		// Send verification email via backend task
		group := crew.NewTaskGroup("", "Re-verify User "+user.ID)
		taskGroupCreateError := controller.crewContoller.CreateTaskGroup(group)
		if taskGroupCreateError != nil {
			log.Print(taskGroupCreateError)
		}

		task := crew.NewTask()
		task.TaskGroupId = group.Id
		task.Name = "Re-Send Verification Email"
		task.Worker = "verify-email"
		task.Input = VerifyEmailJobInput{
			UserId: user.ID,
		}
		taskCreateError := controller.crewContoller.CreateTask(task)
		if taskCreateError != nil {
			log.Print(taskCreateError)
		}

		return c.JSON(http.StatusOK, true)
	})

	e.POST("/hasura/actions/verifyEmail", func(c echo.Context) error {
		// Body should contain email
		body := VerifyEmailBody{}
		bodyErr := json.NewDecoder(c.Request().Body).Decode(&body)
		if bodyErr != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": "code is required.",
			})
		}

		_, verifyError := controller.client.Users.FindMany(
			db.Users.EmailVerificationCode.Equals(body.Input.Code),
		).Update(
			db.Users.EmailVerified.Set(true),
			db.Users.EmailVerificationCode.SetOptional(nil),
		).Exec(c.Request().Context())
		if verifyError != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": verifyError.Error(),
			})
		}

		return c.JSON(http.StatusOK, true)
	})

	return nil
}

func (controller *ActionsController) Shutdown() {
	log.Print("Shutting down actions controller...")
	if controller.client != nil {
		controller.client.Disconnect()
	}
}
