package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/aaronblondau/hasura-base-go/prisma/db"
	"github.com/aaronblondeau/crew-go/crew"
	"github.com/labstack/echo/v4"
)

type EventBody struct {
	Event struct {
		Data struct {
			New struct {
				ID string `json:"id"`
			} `json:"new"`
			Old struct {
				ID string `json:"id"`
			} `json:"old"`
		} `json:"data"`
	} `json:"event"`
	Trigger struct {
		Name string `json:"name"`
	} `json:"trigger"`
}

type EventsController struct {
	client        *db.PrismaClient
	crewContoller *crew.TaskController
}

func NewEventsController(e *echo.Echo, crewController *crew.TaskController) *EventsController {
	controller := EventsController{
		client:        db.NewClient(),
		crewContoller: crewController,
	}
	return &controller
}

func (controller *EventsController) Run(e *echo.Echo) error {

	if err := controller.client.Prisma.Connect(); err != nil {
		return err
	}

	e.POST("/hasura/events", func(c echo.Context) error {
		body := EventBody{}
		bodyErr := json.NewDecoder(c.Request().Body).Decode(&body)
		if bodyErr != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": "unable to decode event payload.",
			})
		}

		if body.Trigger.Name == "insert_contact_form_submission" {
			submissionId := body.Event.Data.New.ID
			// Convert to a crew task
			notifyContactFOrmSubmissionTaskError := CreateCrewTask(controller.crewContoller, "Notify admins of contact form submission "+submissionId, "contact-form-notify-admins", ContactFormNotifyAdminsJobInput{
				SubmissionId: submissionId,
			})
			if notifyContactFOrmSubmissionTaskError != nil {
				// Only log error here and do not blow up entire call if task create fails. Users can request re-send.
				log.Print(notifyContactFOrmSubmissionTaskError)
			}
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"success": "Thanks for the " + body.Trigger.Name + " Hasura!",
			"at":      time.Now().String(),
		})
	})

	return nil
}

func (controller *EventsController) Shutdown() {
	log.Print("Shutting down events controller...")
	if controller.client != nil {
		controller.client.Disconnect()
	}
}
