package api

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/echosonusharma/image-uploader-service/logger"
	"github.com/echosonusharma/image-uploader-service/utils"

	"github.com/go-chi/httprate"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type ApiServer struct {
	listenAddr   string
	writeTimeout time.Duration
	readTimeout  time.Duration
	cors         *cors.Cors
}

func NewApiServer(addr string, cors *cors.Cors) *ApiServer {
	return &ApiServer{
		listenAddr:   addr,
		writeTimeout: 15 * time.Second,
		readTimeout:  15 * time.Second,
		cors:         cors,
	}
}

func (s *ApiServer) Run() {
	router := mux.NewRouter()

	router.Use(s.cors.Handler)

	// per ip & per path
	router.Use(httprate.Limit(
		100,            // requests
		10*time.Second, // per duration
		httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint),
	))

	router.NotFoundHandler = http.HandlerFunc(makeHTTPHandlerFunc(s.NotFound))

	router.HandleFunc("/", makeHTTPHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		if err := utils.WithJSON(w, 200, Msg{Msg: "Welcome to the image-uploader-service 😄!"}); err != nil {
			return err
		}

		return nil
	})).Methods(http.MethodGet)

	mainRouter := router.PathPrefix("/api/v1").Subrouter()

	mainRouter.HandleFunc("/ping", makeHTTPHandlerFunc(s.PingHandler)).Methods(http.MethodGet)

	// user handlers
	mainRouter.HandleFunc("/user/create", makeHTTPHandlerFunc(s.HandleCreateUser)).Methods(http.MethodPost)

	mainRouter.HandleFunc("/user/update/{userId}", makeHTTPHandlerFunc(s.HandleUpdateUser)).Methods(http.MethodPut)

	mainRouter.HandleFunc("/user/delete/{userId}", makeHTTPHandlerFunc(s.HandleDeleteUser)).Methods(http.MethodDelete)

	mainRouter.HandleFunc("/user/{userId}", makeHTTPHandlerFunc(s.HandleGetUser)).Methods(http.MethodGet)

	mainRouter.HandleFunc("/user", makeHTTPHandlerFunc(s.HandleGetAllUser)).Methods(http.MethodGet)

	mainRouter.HandleFunc("/user/profilePic", makeHTTPHandlerFunc(s.HandleUploadUserProfilePic)).Methods(http.MethodPost)
	mainRouter.HandleFunc("/user/profilePic/{userId}", makeHTTPHandlerFunc(s.HandleUploadUserProfilePic)).Methods(http.MethodPost)

	logger.Log.Info(fmt.Sprintf("🚀 Server is running on http://127.0.0.1:%s", s.listenAddr))

	srv := &http.Server{
		Handler:      utils.StripSlashes(requestLog(router)),
		Addr:         fmt.Sprintf("127.0.0.1:%s", s.listenAddr),
		WriteTimeout: s.writeTimeout,
		ReadTimeout:  s.readTimeout,
	}

	log.Fatal(srv.ListenAndServe())
}

type Msg struct {
	Msg string `json:"msg"`
}

type ApiErr struct {
	Err        string `json:"err"`
	StatusCode int
}

type ApiGenericRes struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
}

func (e *ApiErr) Error() string {
	return e.Err
}

type apiHandlerFunc func(http.ResponseWriter, *http.Request) error

var genericApiErr *ApiErr = &ApiErr{
	Err:        "something went wrong!",
	StatusCode: http.StatusInternalServerError,
}

// wrapper for error handling
func makeHTTPHandlerFunc(f apiHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {

			// if api err then send err
			if apiErr, ok := err.(*ApiErr); ok {
				if apiErr.StatusCode == 0 {
					apiErr.StatusCode = http.StatusBadRequest
				}
				utils.WithJSON(w, apiErr.StatusCode, apiErr)
			} else {
				utils.WithJSON(w, genericApiErr.StatusCode, genericApiErr)
			}
		}
	}
}

// middleware for request logging
func requestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(startTime)

		logger.Log.Info("success",
			slog.Group("request",
				slog.String("method", r.Method),
				slog.String("path", r.RequestURI),
				slog.String("duration", duration.String())),
		)
	})
}
