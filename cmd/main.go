package main

import (
	"log"
	"strings"

	"github.com/echosonusharma/image-uploader-service/api"
	"github.com/echosonusharma/image-uploader-service/config"
	"github.com/echosonusharma/image-uploader-service/db"
	"github.com/rs/cors"
)

func main() {
	if err := config.LoadEnvs(); err != nil {
		log.Fatal(err)
	}

	//db
	if err := db.InitDB(config.Cfg.SQL_DATABASE_URL); err != nil {
		log.Fatalf("couldn't initialize db: %v", err)
	}

	defer db.Db.Close()

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
