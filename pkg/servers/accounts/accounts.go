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
	"sync"

	"golang.org/x/time/rate"
)

var visitors = make(map[string]*rate.Limiter)
var mu sync.Mutex

func getLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()
	limiter, exists := visitors[ip]
	if !exists {
		limiter = rate.NewLimiter(1, 5)
		visitors[ip] = limiter
	}
	return limiter
}

func rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		limiter := getLimiter(ip)
		if !limiter.Allow() {
			vars.HTTPError(w, "rate limit exceeded", vars.CodeTooManyRequests, http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func maxRequestSizeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB = 2^20 bytes
		next.ServeHTTP(w, r)
	})
}

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
				vars.HTTPError(w, vars.CodeSessionCertNotFound, "cert not found", 500)
			}
		} else {
			vars.HTTPError(w, vars.CodeSessionCertNotFound, "cert not found", 500)
		}
	}
}

func main() {
	http.Handle("/v1/", maxRequestSizeMiddleware(rateLimitMiddleware(http.HandlerFunc(AccountsAPI))))
	http.ListenAndServe(":8080", nil)
}
