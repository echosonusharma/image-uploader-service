package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/echosonusharma/image-uploader-service/api"
	"github.com/echosonusharma/image-uploader-service/config"
	"github.com/echosonusharma/image-uploader-service/db"
	"github.com/echosonusharma/image-uploader-service/logger"
	"github.com/robfig/cron"
	"github.com/rs/cors"
)

func main() {
	if err := config.LoadEnvs(); err != nil {
		log.Fatalf("failed to load envs: %v", err)
	}

	if err := logger.Init(); err != nil {
		log.Fatalf("failed to initialize the logger: %v", err)
	}

	if err := db.InitDB(config.Cfg.SQL_DATABASE_URL); err != nil {
		log.Fatalf("couldn't initialize db: %v", err)
	}

	c := cron.New()

	c.AddFunc("@every 5m", func() {
		jobName := "clear-storage"

		files, err := os.ReadDir("./storage")
		if err != nil {
			logger.Log.Error("cron job - failed",
				slog.Group("error",
					slog.String("job name", jobName),
					slog.String("status", "failed"),
					slog.String("reason", err.Error()),
				))
		}

		fileNames := []string{}

		for _, file := range files {
			fileNames = append(fileNames, file.Name())
		}

		limit := 10
		count := 0

		for {
			offset := limit * count

			users, getUserErr := db.GetAllUsers(int64(limit), int64(offset))
			if getUserErr != nil {
				logger.Log.Error("cron job - failed",
					slog.Group("error",
						slog.String("job name", jobName),
						slog.String("status", "failed"),
						slog.String("reason", err.Error()),
					))

				return
			}

			if len(users) == 0 {
				break
			}

			for _, user := range users {
				fileIndex := slices.Index(fileNames, user.ProfilePic)
				if fileIndex != -1 {
					fileNames = slices.Delete(fileNames, fileIndex, fileIndex+1)
				}
			}

			count++
		}

		for _, fileName := range fileNames {
			if err := os.Remove(filepath.Join("./storage", fileName)); err != nil {
				logger.Log.Error("cron job - failed to delete file",
					slog.Group("error",
						slog.String("job name", jobName),
						slog.String("status", "processing"),
						slog.String("reason", err.Error()),
					))
			} else {
				logger.Log.Info(fmt.Sprintf("cron job - %s - %s file deleted!", jobName, fileName))
			}
		}

		logger.Log.Info(fmt.Sprintf("cron job - %s completed", jobName))
	})

	c.Start()

	defer func() {
		db.Db.Close()
		logger.LogFile.Close()
		c.Stop()
	}()

	// CORS configuration
	corsOptions := cors.New(cors.Options{
		AllowedOrigins:   strings.Split(config.Cfg.CORS, ","),
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		Debug:            false,
	})

	server := api.NewApiServer(config.Cfg.PORT, corsOptions)
	server.Run()
}
