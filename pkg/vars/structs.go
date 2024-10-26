package vars

import (
	"encoding/json"
	"net/http"
	"time"
)

// -- ACCOUNTS --

type Session struct {
	SessionToken string    `json:"session_token"`
	UserID       string    `json:"user_id"`
	Scope        string    `json:"scope"`
	TimeCreated  time.Time `json:"time_created"`
	TimeExpires  time.Time `json:"time_expires"`
}

type User struct {
	User struct {
		UserID               string    `json:"user_id"`
		DriveGuestID         string    `json:"drive_guest_id"`
		PlayerID             string    `json:"player_id"`
		CreatedByAppName     string    `json:"created_by_app_name"`
		CreatedByAppVersion  string    `json:"created_by_app_version"`
		CreatedByAppPlatform string    `json:"created_by_app_platform"`
		Dob                  string    `json:"dob"`
		Email                string    `json:"email"`
		FamilyName           string    `json:"family_name"`
		Gender               string    `json:"gender"`
		GivenName            string    `json:"given_name"`
		Username             string    `json:"username"`
		EmailIsVerified      bool      `json:"email_is_verified"`
		EmailFailureCode     string    `json:"email_failure_code"`
		EmailLang            string    `json:"email_lang"`
		PasswordIsComplex    bool      `json:"password_is_complex"`
		Status               string    `json:"status"`
		TimeCreated          time.Time `json:"time_created"`
		DeactivationReason   string    `json:"deactivation_reason"`
		PurgeReason          string    `json:"purge_reason"`
		EmailIsBlocked       bool      `json:"email_is_blocked"`
		NoAutodelete         bool      `json:"no_autodelete"`
		IsEmailAccount       bool      `json:"is_email_account"`
	}
}

type Sessions struct {
	Session `json:"session"`
	User    `json:"user"`
}

type UserAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// -- GENERAL HTTP --

type HTTPStatus struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

func HTTPSuccess(w http.ResponseWriter, msg string) {
	status := HTTPStatus{
		Code:    "success",
		Message: msg,
		Status:  "success",
	}
	out, _ := json.Marshal(status)
	w.Write(out)
}

func HTTPError(w http.ResponseWriter, msg string, code int) {
	status := HTTPStatus{
		Code:    "error",
		Message: msg,
		Status:  "error",
	}
	out, _ := json.Marshal(status)
	http.Error(w, string(out), code)
}
