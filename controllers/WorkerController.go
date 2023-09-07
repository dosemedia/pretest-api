package controllers

import (
	"bytes"
	"embed"
	"encoding/json"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"aaronblondeau.com/hasura-base-go/controllers/lib"
	"aaronblondeau.com/hasura-base-go/prisma/db"
	"github.com/aaronblondeau/crew-go/crew"
	"github.com/labstack/echo/v4"
)

type WorkerController struct {
	client         *db.PrismaClient
	emailTemplates embed.FS
}

func NewWorkerController(e *echo.Echo, emailTemplates embed.FS) *WorkerController {
	controller := WorkerController{
		client:         db.NewClient(),
		emailTemplates: emailTemplates,
	}
	return &controller
}

type VerifyEmailJobInput struct {
	UserId string `json:"userId"`
}

// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func (controller *WorkerController) Run(e *echo.Echo) error {
	if err := controller.client.Prisma.Connect(); err != nil {
		return err
	}

	e.POST("/crew/worker/verify-email", func(c echo.Context) error {
		token := c.Request().Header.Get("Authorization")
		expectedToken := os.Getenv("CREW_WORKER_AUTHORIZATION_HEADER")

		if token != expectedToken {
			return c.String(http.StatusUnauthorized, "Unauthorized")
		}

		payload := crew.WorkerPayload{}
		json.NewDecoder(c.Request().Body).Decode(&payload)

		payloadInput, ok := payload.Input.(map[string]interface{})
		if !ok {
			return c.JSON(http.StatusBadRequest, "Invalid payload")
		}
		userId := payloadInput["userId"].(string)
		log.Println("~~ verify-email worker", userId)

		// Load the user
		user, userErr := controller.client.Users.FindUnique(db.Users.ID.Equals(userId)).Exec(c.Request().Context())
		if userErr != nil {
			return c.JSON(http.StatusBadRequest, userErr.Error())
		}

		// Generate and store verification code
		verificationCode := RandStringRunes(8)
		_, saveCodeError := controller.client.Users.FindUnique(
			db.Users.ID.Equals(userId),
		).Update(
			db.Users.EmailVerificationCode.Set(verificationCode),
		).Exec(c.Request().Context())
		if saveCodeError != nil {
			return c.JSON(http.StatusBadRequest, saveCodeError.Error())
		}

		// Build the email content
		// Note, email templates are embedded : https://pkg.go.dev/embed
		// To debug template with local fs, use this:
		// t, _ := template.ParseFiles("./emails/email-verify.html")
		t, _ := template.ParseFS(controller.emailTemplates, "emails/email-verify.html")
		var htmlBody bytes.Buffer

		verificationUrl := os.Getenv("WEB_BASE_URL") + "/verify-email/" + verificationCode

		t.Execute(&htmlBody, struct {
			VerificationUrl string
		}{
			VerificationUrl: verificationUrl,
		})

		// Send the email
		from := os.Getenv("EMAIL_SENDER")
		emailSendError := lib.SendEmail(from, user.Email, "Verify your AppName email", htmlBody.String())
		if emailSendError != nil {
			return c.JSON(http.StatusBadRequest, emailSendError.Error())
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"success": true,
		})
	})

	return nil
}

func (controller *WorkerController) Shutdown() {
	log.Print("Shutting down worker controller...")
}
