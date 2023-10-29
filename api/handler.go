package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	userHandler "github.com/echosonusharma/image-uploader-service/db"
	"github.com/echosonusharma/image-uploader-service/utils"

	"github.com/gorilla/mux"
)

func (s *ApiServer) HandleCreateUser(w http.ResponseWriter, r *http.Request) error {
	var newUser userHandler.User
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		return &ApiErr{Err: "invalid json or user body provided!"}
	}

	defer r.Body.Close()

	if strings.TrimSpace(newUser.Name) == "" || strings.TrimSpace(newUser.Email) == "" {
		return &ApiErr{Err: "name or email missing!"}
	}

	if _, handlerErr := userHandler.CreateUser(&newUser); handlerErr != nil {
		return &ApiErr{Err: "failed to create user!"}
	}

	jsonEncodeErr := utils.WithJSON(w, http.StatusOK, ApiGenericRes{
		Success: true,
		Msg:     "User created successfully!",
	})

	if jsonEncodeErr != nil {
		return &ApiErr{Err: "failed to encode response to json!"}
	}

	return nil
}

func (s *ApiServer) HandleUpdateUser(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	var newUser userHandler.User

	userId, parseErr := strconv.ParseInt(vars["userId"], 10, 64)
	if parseErr != nil {
		return &ApiErr{Err: "invalid userId provided!"}
	}

	newUser.Id = userId

	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		return &ApiErr{Err: "invalid json or user body provided!"}
	}

	defer r.Body.Close()

	updateErr := userHandler.UpdateUser(&newUser)
	if updateErr != nil {
		return &ApiErr{Err: "failed to update user!"}
	}

	jsonEncodeErr := utils.WithJSON(w, http.StatusOK, ApiGenericRes{
		Success: true,
		Msg:     "User updated successfully!",
	})

	if jsonEncodeErr != nil {
		return &ApiErr{Err: "failed to encode response to json!"}
	}

	return nil
}

func (s *ApiServer) HandleDeleteUser(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)

	userId, parseErr := strconv.ParseInt(vars["userId"], 10, 64)
	if parseErr != nil {
		return &ApiErr{Err: "invalid userId provided!"}
	}

	err := userHandler.DeleteUser(userId)
	if err != nil {
		return &ApiErr{Err: "failed to delete user!"}
	}

	jsonEncodeErr := utils.WithJSON(w, http.StatusOK, ApiGenericRes{
		Success: true,
		Msg:     "User deleted successfully!",
	})

	if jsonEncodeErr != nil {
		return &ApiErr{Err: "failed to encode response to json!"}
	}

	return nil
}

func (s *ApiServer) HandleGetUser(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)

	userId, parseErr := strconv.ParseInt(vars["userId"], 10, 64)
	if parseErr != nil {
		return &ApiErr{Err: "invalid userId provided!"}
	}

	user, err := userHandler.GetUser(userId)
	if err != nil {
		return &ApiErr{Err: "user not found!", StatusCode: http.StatusNotFound}
	}

	jsonEncodeErr := utils.WithJSON(w, http.StatusOK, user)

	if jsonEncodeErr != nil {
		return &ApiErr{Err: "failed to encode response to json!"}
	}

	return nil
}

func (s *ApiServer) HandleGetAllUser(w http.ResponseWriter, r *http.Request) error {
	data, err := userHandler.GetAllUsers()
	if err != nil {
		return err
	}

	jsonEncodeErr := utils.WithJSON(w, http.StatusOK, data)

	if jsonEncodeErr != nil {
		return &ApiErr{Err: "failed to encode response to json!"}
	}

	return nil
}

func (s *ApiServer) HandleUploadUserProfilePic(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	uId := vars["userId"]

	fileName, err := FileUploadHandler(r)
	if err != nil {
		return &ApiErr{
			Err: err.Error(),
		}
	}

	if uId != "" {
		userId, parseErr := strconv.ParseInt(uId, 10, 64)
		if parseErr != nil {
			return &ApiErr{Err: "failed to parse userId!"}
		}

		updateErr := userHandler.UpdateUser(&userHandler.User{Id: userId, ProfilePic: fileName})
		if updateErr != nil {
			return &ApiErr{Err: "failed to update user!"}
		}
	}

	jsonEncodeErr := utils.WithJSON(w, http.StatusOK, &ApiGenericRes{
		Success: true,
		Msg:     "file uploaded successfully!",
	})

	if jsonEncodeErr != nil {
		return &ApiErr{Err: "failed to encode response to json!"}
	}

	return nil
}

// ping
func (s *ApiServer) PingHandler(w http.ResponseWriter, r *http.Request) error {
	if err := utils.WithJSON(w, 200, Msg{Msg: "🚀"}); err != nil {
		return err
	}

	return nil
}

// not found
func (s *ApiServer) NotFound(w http.ResponseWriter, r *http.Request) error {
	if err := utils.WithJSON(w, http.StatusNotFound, &ApiErr{
		Err:        "route not found",
		StatusCode: http.StatusNotFound,
	}); err != nil {
		return err
	}

	return nil
}

// handle file upload
// no compression added so i take in all formats of images
func FileUploadHandler(r *http.Request) (string, error) {
	err := r.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		return "", errors.New("failed to parse form data, make sure it's less than 10Mb")
	}

	file, handler, err := r.FormFile("profilePic")
	if err != nil {
		return "", errors.New("error retrieving the file")
	}

	defer file.Close()

	contentType := handler.Header["Content-Type"]
	if len(contentType) > 0 {
		contentType = strings.Split(contentType[0], "/")
	}

	if len(contentType) == 0 || contentType[0] != "image" {
		return "", errors.New("file is not an image")
	}

	errCh := make(chan error, 1)

	sysFileName := fmt.Sprintf("%d-%s", time.Now().UnixMilli(), handler.Filename)

	// store the file
	go func() {
		f, err := os.Create(fmt.Sprintf("./storage/%s", sysFileName))
		if err != nil {
			errCh <- err
			return
		}

		defer f.Close()

		_, err = io.Copy(f, file)
		if err != nil {
			errCh <- err
			return
		}

		errCh <- nil
	}()

	if err := <-errCh; err != nil {
		log.Println(err)
		return "", errors.New("failed to save the file on the server")
	}

	return sysFileName, nil
}
