package controllers

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"aaronblondeau.com/hasura-base-go/prisma/db"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/disintegration/imaging"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type FileStorageController struct {
	client *db.PrismaClient
}

func NewFileStorageController(e *echo.Echo) *FileStorageController {
	controller := FileStorageController{
		client: db.NewClient(),
	}
	return &controller
}

func getUserFromAuthHeader(c echo.Context, client *db.PrismaClient) (*db.UsersModel, error) {
	token := c.Request().Header.Get("Authorization")
	if token == "" {
		return nil, fmt.Errorf("no token provided")
	}

	token = strings.Replace(token, "Bearer ", "", 1)
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
		return nil, tokenError
	}
	if !parsed.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims := parsed.Claims.(jwt.MapClaims)
	userId := claims["user_id"].(string)

	return client.Users.FindUnique(db.Users.ID.Equals(userId)).Exec(c.Request().Context())
}

func getUserPublicS3Client() (*s3.Client, string, string, string, error) {
	// Note, only use endpoint when pointing at minio
	userPublicEndpoint := os.Getenv("S3_USER_PUBLIC_ENDPOINT")
	userPublicRegion := os.Getenv("S3_USER_PUBLIC_REGION")

	if userPublicRegion == "" {
		userPublicRegion = "us-west-2"
	}

	userPublicBucket := os.Getenv("S3_USER_PUBLIC_BUCKET")
	if userPublicBucket == "" {
		userPublicBucket = "user-public"
	}

	var userPublicS3Client *s3.Client

	if userPublicEndpoint != "" {
		// MINIO
		resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...any) (aws.Endpoint, error) {
			return aws.Endpoint{
				PartitionID:       "aws",
				SigningRegion:     userPublicRegion,
				URL:               userPublicEndpoint,
				HostnameImmutable: true,
			}, nil
		})
		cfg := aws.Config{
			Region:                      userPublicRegion,
			EndpointResolverWithOptions: resolver,
			Credentials:                 credentials.NewStaticCredentialsProvider(os.Getenv("S3_ACCESS_KEY"), os.Getenv("S3_SECRET_KEY"), ""),
		}
		userPublicS3Client = s3.NewFromConfig(cfg)
	} else {
		// REAL S3 (TODO UNTESTED!)
		cfg, err := config.LoadDefaultConfig(context.Background(), config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(os.Getenv("S3_ACCESS_KEY"), os.Getenv("S3_SECRET_KEY"), "")))
		if err != nil {
			return nil, "", "", "", err
		}
		userPublicS3Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
			if userPublicEndpoint != "" {
				fmt.Println("~~ Using custom S3 endpoint", userPublicEndpoint)
				o.BaseEndpoint = aws.String(userPublicEndpoint)
			}
			o.Region = userPublicRegion
		})
	}
	return userPublicS3Client, userPublicBucket, userPublicEndpoint, userPublicRegion, nil
}

func (controller *FileStorageController) Run(e *echo.Echo) error {

	if err := controller.client.Prisma.Connect(); err != nil {
		return err
	}

	userPublicS3Client, userPublicBucket, userPublicEndpoint, userPublicRegion, userPublicS3ClientError := getUserPublicS3Client()
	if userPublicS3ClientError != nil {
		return userPublicS3ClientError
	}

	e.POST("/files/user-avatar", func(c echo.Context) error {
		user, userError := getUserFromAuthHeader(c, controller.client)
		if userError != nil {
			return c.String(http.StatusUnauthorized, userError.Error())
		}

		// https://echo.labstack.com/docs/cookbook/file-upload
		// File field is named "avatar"
		file, err := c.FormFile("avatar")
		if err != nil {
			return err
		}
		src, err := file.Open()
		if err != nil {
			return err
		}
		defer src.Close()

		// Upload input parameters
		ext := filepath.Ext(file.Filename)
		mtype := mime.TypeByExtension(ext)
		key := user.ID + "/" + uuid.New().String() + ext

		// Perform an upload to S3
		// https://github.com/awsdocs/aws-doc-sdk-examples/blob/main/gov2/s3/actions/bucket_basics.go#L100
		_, uploadErr := userPublicS3Client.PutObject(context.Background(), &s3.PutObjectInput{
			Bucket: aws.String(userPublicBucket),
			Key:    aws.String(key),
			Body:   src,
		})
		if uploadErr != nil {
			return c.String(http.StatusInternalServerError, uploadErr.Error())
		}

		// Save file key to user profile
		_, updateUserError := controller.client.Users.FindUnique(
			db.Users.ID.Equals(user.ID),
		).Update(
			db.Users.AvatarFileKey.Set(key),
		).Exec(c.Request().Context())
		if updateUserError != nil {
			return c.String(http.StatusInternalServerError, updateUserError.Error())
		}

		// This response needs to look like a multer file (node.js)
		return c.JSON(http.StatusOK, map[string]interface{}{
			"bucket":       userPublicBucket,
			"originalname": file.Filename,
			"mimetype":     mtype,
			"key":          key,
			"size":         file.Size,
			"endpoint":     userPublicEndpoint,
			"region":       userPublicRegion,
			"userId":       user.ID,
		})
	})

	e.GET("/files/user-avatar/:userId/:fileId", func(c echo.Context) error {
		userId := c.Param("userId")
		fileId := c.Param("fileId")

		result, getObjectError := userPublicS3Client.GetObject(c.Request().Context(), &s3.GetObjectInput{
			Bucket: aws.String(userPublicBucket),
			Key:    aws.String(userId + "/" + fileId),
		})
		if getObjectError != nil {
			return c.String(http.StatusInternalServerError, getObjectError.Error())
		}
		defer result.Body.Close()

		// // To return raw image from s3:
		// return c.Stream(http.StatusOK, *result.ContentType, result.Body)

		// Resize image:
		src, _, decodeError := image.Decode(result.Body)
		if decodeError != nil {
			return c.String(http.StatusInternalServerError, decodeError.Error())
		}

		// resize input image
		dst := imaging.Resize(src, 200, 200, imaging.Lanczos)

		encoded := &bytes.Buffer{}
		encodedError := png.Encode(encoded, dst)
		if encodedError != nil {
			return c.String(http.StatusInternalServerError, encodedError.Error())
		}

		return c.Stream(http.StatusOK, *result.ContentType, encoded)

		// // Save to a temp file and return (for use with image processing libs that need a file path)

		// file, fileErr := os.CreateTemp("", fileId)
		// if fileErr != nil {
		// 	return c.String(http.StatusInternalServerError, fileErr.Error())
		// }
		// defer file.Close()
		// // Doesn't work on windows - why?
		// defer os.Remove(file.Name())

		// body, readErr := io.ReadAll(result.Body)
		// if readErr != nil {
		// 	return c.String(http.StatusInternalServerError, readErr.Error())
		// }
		// _, fileWriteErr := file.Write(body)
		// if fileWriteErr != nil {
		// 	return c.String(http.StatusInternalServerError, fileWriteErr.Error())
		// }

		// return c.File(file.Name())
	})

	return nil
}

func (controller *FileStorageController) Shutdown() {
	log.Print("Shutting down file storage controller...")
	if controller.client != nil {
		controller.client.Disconnect()
	}
}
