package accounts

import (
	"cavalier/pkg/sessions"
	"cavalier/pkg/users"
	"cavalier/pkg/vars"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func AccountsAPI(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Path)
	switch r.URL.Path {
	case "/v1/sessions":
		body, err := io.ReadAll(r.Body)
		if err != nil {
			vars.HTTPError(w, "failed to read request body: "+err.Error(), vars.CodeServerError, 500)
			return
		}
		var creds vars.UserAuth
		err = json.Unmarshal(body, &creds)
		if err != nil {
			vars.HTTPError(w, "failed to unmarshal json: "+err.Error(), vars.CodeServerError, 500)
			return
		}
		user, err := users.AuthUser(creds.Username, creds.Password)
		if err != nil {
			vars.HTTPError(w, err.Error(), err.Error(), 400)
			return
		}
		fullUser := vars.User{
			UserID:            user.UserID,
			PlayerID:          user.UUID,
			DriveGuestID:      user.UUID,
			Email:             user.Email,
			Username:          user.Email,
			EmailIsVerified:   true,
			PasswordIsComplex: true,
			Status:            "active",
			EmailIsBlocked:    false,
			NoAutodelete:      false,
			IsEmailAccount:    true,
			Dob:               user.DOB,
		}
		session := sessions.NewSession(user.UserID)
		var fullSession vars.Sessions
		fullSession.Session = session
		fullSession.User = fullUser
		writeBytes, err := json.Marshal(fullSession)
		if err != nil {
			vars.HTTPError(w, "failed to marshal json: "+err.Error(), vars.CodeServerError, 500)
			return
		}
		w.Write(writeBytes)
	case "/v1/create_user":
		body, err := io.ReadAll(r.Body)
		if err != nil {
			vars.HTTPError(w, "failed to read request body: "+err.Error(), vars.CodeServerError, 500)
		}
		var creds vars.CreateUser
		err = json.Unmarshal(body, &creds)
		if err != nil {
			vars.HTTPError(w, "failed to unmarshal json: "+err.Error(), vars.CodeServerError, 500)
		}
		err = users.CreateUser(creds.Username, creds.Password, creds.DOB)
		if err != nil {
			vars.HTTPError(w, err.Error(), err.Error(), 400)
			return
		}
		vars.HTTPSuccess(w, "account created")
	}

	if strings.HasPrefix(r.URL.Path, "/v1/session_cert/") {
		urlSplit := strings.Split(r.URL.Path, "/")
		if len(urlSplit) == 4 {
			cert, err := os.ReadFile(filepath.Join(vars.SessionCertsStorage, urlSplit[3]))
			if err == nil {
				w.Write(cert)
			} else {
				vars.HTTPError(w, "missing_cert", "cert not found", 500)
			}
		} else {
			vars.HTTPError(w, "missing_cert", "cert not found", 500)
		}
	}
}
