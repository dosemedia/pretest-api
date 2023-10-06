package controllers

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aaronblondau/hasura-base-go/controllers/lib"
	"github.com/aaronblondau/hasura-base-go/prisma/db"
	"github.com/aaronblondeau/crew-go/crew"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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

type ResetPasswordEmailJobInput struct {
	UserId string `json:"userId"`
}

type PasswordChangedEmailJobInput struct {
	UserId string `json:"userId"`
}

type UserDestroyedEmailJobInput struct {
	Email string `json:"email"`
}

type UserDestroyedCleanupFilesJobInput struct {
	UserId string `json:"userId"`
}

type ContactFormNotifyAdminsJobInput struct {
	SubmissionId string `json:"submissionId"`
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

func CreateCrewTask(crewController *crew.TaskController, taskName string, worker string, input interface{}) error {
	// Create task group for task
	group := crew.NewTaskGroup("", taskName)
	taskGroupCreateError := crewController.CreateTaskGroup(group)
	if taskGroupCreateError != nil {
		return taskGroupCreateError
	}

	// Create the task
	task := crew.NewTask()
	task.TaskGroupId = group.Id
	task.Name = taskName
	task.Worker = worker
	task.Input = input
	taskCreateError := crewController.CreateTask(task)
	return taskCreateError
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
		// t, _ := template.ParseFiles("./emails/build_production/email-verify.html")

		rawHtml, rawHtmlError := controller.emailTemplates.ReadFile("emails/build_production/email-verify.html")
		if rawHtmlError != nil {
			return c.JSON(http.StatusBadRequest, rawHtmlError.Error())
		}
		// Note, delims are changed to make to templates easier to use with Maizzle
		t, tError := template.New("").Delims("[[", "]]").Parse(string(rawHtml))
		if tError != nil {
			return c.JSON(http.StatusBadRequest, tError.Error())
		}

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

	e.POST("/crew/worker/password-reset-email", func(c echo.Context) error {
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
		log.Println("~~ password-reset-email worker", userId)

		// Load the user
		user, userErr := controller.client.Users.FindUnique(db.Users.ID.Equals(userId)).Exec(c.Request().Context())
		if userErr != nil {
			return c.JSON(http.StatusBadRequest, userErr.Error())
		}

		// Generate and store password reset code
		resetCode := RandStringRunes(8)
		_, saveCodeError := controller.client.Users.FindUnique(
			db.Users.ID.Equals(userId),
		).Update(
			db.Users.PasswordResetCode.Set(resetCode),
		).Exec(c.Request().Context())
		if saveCodeError != nil {
			return c.JSON(http.StatusBadRequest, saveCodeError.Error())
		}

		// Build the email content
		// Note, email templates are embedded : https://pkg.go.dev/embed
		// To debug template with local fs, use this:
		// t, _ := template.ParseFiles("./emails/build_production/password-reset.html")

		rawHtml, rawHtmlError := controller.emailTemplates.ReadFile("emails/build_production/password-reset.html")
		if rawHtmlError != nil {
			return c.JSON(http.StatusBadRequest, rawHtmlError.Error())
		}
		// Note, delims are changed to make to templates easier to use with Maizzle
		t, tError := template.New("").Delims("[[", "]]").Parse(string(rawHtml))
		if tError != nil {
			return c.JSON(http.StatusBadRequest, tError.Error())
		}

		var htmlBody bytes.Buffer

		resetPasswordUrl := os.Getenv("WEB_BASE_URL") + "/reset-password/" + resetCode

		t.Execute(&htmlBody, struct {
			ResetPasswordUrl string
		}{
			ResetPasswordUrl: resetPasswordUrl,
		})

		// Send the email
		from := os.Getenv("EMAIL_SENDER")
		emailSendError := lib.SendEmail(from, user.Email, "AppName password reset request", htmlBody.String())
		if emailSendError != nil {
			return c.JSON(http.StatusBadRequest, emailSendError.Error())
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"success": true,
		})
	})

	e.POST("/crew/worker/password-changed-email", func(c echo.Context) error {
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
		log.Println("~~ password-changed-email worker", userId)

		// Load the user
		user, userErr := controller.client.Users.FindUnique(db.Users.ID.Equals(userId)).Exec(c.Request().Context())
		if userErr != nil {
			return c.JSON(http.StatusBadRequest, userErr.Error())
		}

		// Build the email content
		// Note, email templates are embedded : https://pkg.go.dev/embed
		// To debug template with local fs, use this:
		// t, _ := template.ParseFiles("./emails/build_production/password-changed.html")

		rawHtml, rawHtmlError := controller.emailTemplates.ReadFile("emails/build_production/password-changed.html")
		if rawHtmlError != nil {
			return c.JSON(http.StatusBadRequest, rawHtmlError.Error())
		}
		// Note, delims are changed to make to templates easier to use with Maizzle
		t, tError := template.New("").Delims("[[", "]]").Parse(string(rawHtml))
		if tError != nil {
			return c.JSON(http.StatusBadRequest, tError.Error())
		}

		var htmlBody bytes.Buffer

		t.Execute(&htmlBody, nil)

		// Send the email
		from := os.Getenv("EMAIL_SENDER")
		emailSendError := lib.SendEmail(from, user.Email, "AppName password changed", htmlBody.String())
		if emailSendError != nil {
			return c.JSON(http.StatusBadRequest, emailSendError.Error())
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"success": true,
		})
	})

	e.POST("/crew/worker/user-destroyed-email", func(c echo.Context) error {
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
		email := payloadInput["email"].(string)
		log.Println("~~ user-destroyed-email worker", email)

		// Build the email content
		// Note, email templates are embedded : https://pkg.go.dev/embed
		// To debug template with local fs, use this:
		// t, _ := template.ParseFiles("./emails/build_production/user-destroyed.html")

		rawHtml, rawHtmlError := controller.emailTemplates.ReadFile("emails/build_production/user-destroyed.html")
		if rawHtmlError != nil {
			return c.JSON(http.StatusBadRequest, rawHtmlError.Error())
		}
		// Note, delims are changed to make to templates easier to use with Maizzle
		t, tError := template.New("").Delims("[[", "]]").Parse(string(rawHtml))
		if tError != nil {
			return c.JSON(http.StatusBadRequest, tError.Error())
		}

		var htmlBody bytes.Buffer

		t.Execute(&htmlBody, nil)

		// Send the email
		from := os.Getenv("EMAIL_SENDER")
		emailSendError := lib.SendEmail(from, email, "AppName account destroyed", htmlBody.String())
		if emailSendError != nil {
			return c.JSON(http.StatusBadRequest, emailSendError.Error())
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"success": true,
		})
	})

	e.POST("/crew/worker/user-destroyed-cleanup-files", func(c echo.Context) error {
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
		log.Println("~~ user-destroyed-cleanup-files worker", userId)

		// cleanup all files in user-public bucket with prefix userId/
		userPublicS3Client, userPublicBucket, _, _, userPublicS3ClientError := getUserPublicS3Client()
		if userPublicS3ClientError != nil {
			return c.String(http.StatusInternalServerError, userPublicS3ClientError.Error())
		}

		listResult, listResultError := userPublicS3Client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
			Bucket: aws.String(userPublicBucket),
			Prefix: aws.String(userId + "/"),
		})
		if listResultError != nil {
			return c.String(http.StatusInternalServerError, listResultError.Error())
		}

		for _, object := range listResult.Contents {
			log.Println("~~ user-destroyed-cleanup-files worker deleting", *object.Key)
			_, deleteError := userPublicS3Client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
				Bucket: aws.String(userPublicBucket),
				Key:    object.Key,
			})
			if deleteError != nil {
				return c.String(http.StatusInternalServerError, deleteError.Error())
			}
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"success": true,
		})
	})

	e.POST("/crew/worker/contact-form-notify-admins", func(c echo.Context) error {
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
		submissionId := payloadInput["submissionId"].(string)
		log.Println("~~ contact-form-notify-admins worker", submissionId)

		// Fetch the submission
		submission, submissionErr := controller.client.ContactFormSubmissions.FindUnique(db.ContactFormSubmissions.ID.Equals(submissionId)).Exec(c.Request().Context())
		if submissionErr != nil {
			return c.JSON(http.StatusBadRequest, submissionErr.Error())
		}

		// Build the email content
		// Note, email templates are embedded : https://pkg.go.dev/embed
		// To debug template with local fs, use this:
		// t, _ := template.ParseFiles("./emails/build_production/contact-form-submission.html")

		rawHtml, rawHtmlError := controller.emailTemplates.ReadFile("emails/build_production/contact-form-submission.html")
		if rawHtmlError != nil {
			return c.JSON(http.StatusBadRequest, rawHtmlError.Error())
		}
		// Note, delims are changed to make to templates easier to use with Maizzle
		t, tError := template.New("").Delims("[[", "]]").Parse(string(rawHtml))
		if tError != nil {
			return c.JSON(http.StatusBadRequest, tError.Error())
		}

		var htmlBody bytes.Buffer

		submissionName, nameOk := submission.Name()
		if !nameOk {
			submissionName = "?"
		}

		t.Execute(&htmlBody, struct {
			Name    string
			Email   string
			Message string
		}{
			Name:    submissionName,
			Email:   submission.Email,
			Message: submission.Message,
		})

		recipientsEnv := os.Getenv("CONTACT_FORM_RECIPIENT_EMAILS")
		if recipientsEnv == "" {
			recipientsEnv = "admin@AppName.com"
		}
		recipients := strings.Split(recipientsEnv, ",")

		// Send the email(s)
		for _, recipient := range recipients {
			from := os.Getenv("EMAIL_SENDER")
			emailSendError := lib.SendEmail(from, recipient, "AppName Contact Form Submission", htmlBody.String())
			if emailSendError != nil {
				return c.JSON(http.StatusBadRequest, emailSendError.Error())
			}
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
