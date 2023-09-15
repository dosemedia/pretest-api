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
	goredislib "github.com/redis/go-redis/v9"
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

type PasswordResetRequestBody struct {
	Input struct {
		Email string `json:"email"`
	} `json:"input"`
}

type PasswordResetBody struct {
	Input struct {
		Email       string `json:"email"`
		NewPassword string `json:"newPassword"`
		Code        string `json:"code"`
	} `json:"input"`
}

type ActionsController struct {
	client        *db.PrismaClient
	crewContoller *crew.TaskController
	cache         *goredislib.Client
}

func NewActionsController(e *echo.Echo, crewController *crew.TaskController) *ActionsController {
	redisAddress := os.Getenv("REDIS_ADDRESS")
	if redisAddress == "" {
		redisAddress = "localhost:6379"
	}
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB := 0

	controller := ActionsController{
		client:        db.NewClient(),
		crewContoller: crewController,
		cache: goredislib.NewClient(&goredislib.Options{
			Addr:     redisAddress,
			Password: redisPassword,
			DB:       redisDB,
		}),
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

		taskCreateError := CreateCrewTask(controller.crewContoller, "Send initial verification email to user "+createdUser.ID, "verify-email", VerifyEmailJobInput{
			UserId: createdUser.ID,
		})
		if taskCreateError != nil {
			// Only log error here and do not blow up entire call if task create fails. Users can request re-send.
			log.Print(taskCreateError)
		}

		// create auth token for user
		token, tokenErr := generateTokenForUser(*createdUser)
		if tokenErr != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"message": tokenErr.Error(),
			})
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
		taskCreateError := CreateCrewTask(controller.crewContoller, "Send verification email to user "+user.ID, "verify-email", VerifyEmailJobInput{
			UserId: user.ID,
		})
		if taskCreateError != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": taskCreateError.Error(),
			})
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

	e.POST("/hasura/actions/sendPasswordResetEmail", func(c echo.Context) error {
		// Body should contain email
		body := PasswordResetRequestBody{}
		bodyErr := json.NewDecoder(c.Request().Body).Decode(&body)
		if bodyErr != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": "email is required.",
			})
		}

		user, userError := controller.client.Users.FindUnique(db.Users.Email.Equals(body.Input.Email)).Exec(c.Request().Context())
		if userError != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": userError.Error(),
			})
		}

		// Send verification/welcome email via backend task
		taskCreateError := CreateCrewTask(controller.crewContoller, "Password reset email for user "+user.ID, "password-reset-email", ResetPasswordEmailJobInput{
			UserId: user.ID,
		})
		// We should send an error back if task fails to create as user will never get email
		if taskCreateError != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": taskCreateError.Error(),
			})
		}

		return c.JSON(http.StatusOK, true)
	})

	e.POST("/hasura/actions/resetPassword", func(c echo.Context) error {
		// Body should contain email
		body := PasswordResetBody{}
		bodyErr := json.NewDecoder(c.Request().Body).Decode(&body)
		if bodyErr != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": "email, newPassword, and code are all required.",
			})
		}

		user, userError := controller.client.Users.FindFirst(db.Users.Email.Equals(body.Input.Email), db.Users.PasswordResetCode.Equals(body.Input.Code)).Exec(c.Request().Context())
		if userError != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": userError.Error(),
			})
		}

		hashedPassword, hashError := bcrypt.GenerateFromPassword([]byte(body.Input.NewPassword), 10)
		if hashError != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"message": hashError.Error(),
			})
		}

		_, updateError := controller.client.Users.FindUnique(db.Users.ID.Equals(user.ID)).Update(
			db.Users.PasswordResetCode.SetOptional(nil),
			db.Users.HashedPassword.Set(string(hashedPassword))).Exec(c.Request().Context())

		if updateError != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": updateError.Error(),
			})
		}

		// Clear cached auth tokens for this user. Tokens are cached with prefix "userId:"
		FlushRedisPrefix(controller.cache, user.ID+":")

		// Send password changed email
		taskCreateError := CreateCrewTask(controller.crewContoller, "Password changed email for user "+user.ID, "password-changed-email", PasswordChangedEmailJobInput{
			UserId: user.ID,
		})
		if taskCreateError != nil {
			// Only log error here and do not blow up entire call if task create fails since password was already reset.
			log.Print(taskCreateError)
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
