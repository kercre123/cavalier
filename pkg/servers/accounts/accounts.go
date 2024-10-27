package accounts

import (
	"cavalier/pkg/users"
	"cavalier/pkg/vars"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// - sends:

// ```
// {
//     "username": "redacted@gmail.com",
//     "password": "redacted"
// }
// ```

// - receives:
// ```
// {
//     "session": {
//         "session_token": "redacted",
//         "user_id": "redacted",
//         "scope": "user",
//         "time_created": "2024-10-26T00:26:57.174620948Z",
//         "time_expires": "2025-10-26T00:26:57.174600148Z"
//     },
//     "user": {
//         "user_id": "redacted",
//         "drive_guest_id": "b80a7379-211a-4d7c-8440-01aa954635e1",
//         "player_id": "b80a7379-211a-4d7c-8440-01aa954635e1",
//         "created_by_app_name": null,
//         "created_by_app_version": null,
//         "created_by_app_platform": null,
//         "dob": "1970-01-01",
//         "email": "redacted@gmail.com",
//         "family_name": null,
//         "gender": null,
//         "given_name": null,
//         "username": "redacted@gmail.com",
//         "email_is_verified": true,
//         "email_failure_code": null,
//         "email_lang": null,
//         "password_is_complex": true,
//         "status": "active",
//         "time_created": "2024-10-24T18:42:56Z",
//         "deactivation_reason": null,
//         "purge_reason": null,
//         "email_is_blocked": false,
//         "no_autodelete": false,
//         "is_email_account": true
//     }
// }
// ```

// - if bad creds:
// ```
// {
//   "code": "server_failure",
//   "message": "An unexpected server error occurred",
//   "status": "error"
// }
// ```

// let's define a REST API

func AccountsAPI(w http.ResponseWriter, r *http.Request) {
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
			Dob:               "1970-01-01",
		}
		fmt.Println(fullUser)
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
	}
}
