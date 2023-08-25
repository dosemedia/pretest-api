package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"aaronblondeau.com/hasura-base-go/controllers"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "hasura-base-go!")
	})

	e.GET("/readycheck", func(c echo.Context) error {
		return c.String(http.StatusOK, "I'm ready!")
	})

	e.GET("/healthcheck", func(c echo.Context) error {
		hasuraBaseUrl := os.Getenv("HASURA_BASE_URL")
		if hasuraBaseUrl == "" {
			hasuraBaseUrl = "http://localhost:8080"
		}
		url := hasuraBaseUrl + "/v1/version"
		resp, requestErr := http.Get(url)
		if requestErr != nil {
			return c.String(http.StatusInternalServerError, "Failed to fetch hasura status!")
		}
		defer resp.Body.Close()

		var response interface{}
		jsonErr := json.NewDecoder(resp.Body).Decode(&response)
		if jsonErr != nil {
			return c.String(http.StatusInternalServerError, "Failed to parse hasura status!")
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"healthy": true,
			"hasura":  response,
		})
	})

	actionsController := controllers.NewActionsController(e)
	if err := actionsController.Run(e); err != nil {
		panic(err)
	}

	authController := controllers.NewAuthController(e)
	if err := authController.Run(e); err != nil {
		panic(err)
	}

	host := os.Getenv("HOST")
	if host == "" {
		// use localhost on windows, 0.0.0.0 elsewhere
		if runtime.GOOS == "windows" {
			host = "localhost"
		} else {
			host = "0.0.0.0"
		}
	}

	// Hook into the shutdown signal
	cleanupCompleteWg := sync.WaitGroup{}
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-done
		// sigint caught, start graceful shutdown
		log.Print("Process Terminating...")

		// Call shutdown on each controller
		actionsController.Shutdown()
		authController.Shutdown()

		e.Close()

		cleanupCompleteWg.Done()
	}()
	cleanupCompleteWg.Add(1)

	port := os.Getenv("PORT")
	if port == "" {
		port = "1323"
	}
	e.Logger.Error(e.Start(host + ":" + port))

	cleanupCompleteWg.Wait()
}
