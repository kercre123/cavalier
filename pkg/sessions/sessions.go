package sessions

import (
	"cavalier/pkg/vars"
	"fmt"
	"sync"
	"time"
)

var validSessions []vars.Session
var timeFormat string = "2006-01-02T15:04:05.999999999Z"
var sMu sync.Mutex

var ExpirererRunning bool

func getCurrentAndExpireTime() (string, string) {
	currentTime := time.Now().UTC()
	nextYearTime := currentTime.Add(time.Minute * 10)
	currentTimeFormatted := currentTime.Format(timeFormat)
	nextYearTimeFormatted := nextYearTime.Format(timeFormat)
	return currentTimeFormatted, nextYearTimeFormatted
}

func IsExpired(currentTimeStr, expiryTimeStr string) bool {
	currentTime, err := time.Parse(timeFormat, currentTimeStr)
	if err != nil {
		return false
	}
	expiryTime, err := time.Parse(timeFormat, expiryTimeStr)
	if err != nil {
		return false
	}
	return expiryTime.Before(currentTime)
}

func removeToken(token string) {
	var newValidSessions []vars.Session
	for _, session := range validSessions {
		if session.SessionToken != token {
			newValidSessions = append(newValidSessions, session)
		}
	}
	validSessions = newValidSessions
}

func Expirerer() {
	var removeList []string
	for {
		sMu.Lock()
		for _, session := range validSessions {
			currentTime, _ := getCurrentAndExpireTime()
			if IsExpired(currentTime, session.TimeExpires) {
				removeList = append(removeList, session.SessionToken)
			}
		}
		for _, tok := range removeList {
			fmt.Println("expiring token: " + tok)
			removeToken(tok)
		}
		sMu.Unlock()
		time.Sleep(time.Minute)
	}
}

func NewSession(userID string) vars.Session {
	sMu.Lock()
	defer sMu.Unlock()
	cTime, aTime := getCurrentAndExpireTime()
	session := vars.Session{
		SessionToken: vars.GenerateID(),
		UserID:       userID,
		Scope:        "user",
		TimeCreated:  cTime,
		TimeExpires:  aTime,
	}
	validSessions = append(validSessions, session)
	return session
}

func GetUserIDFromSession(sessionToken string) string {
	sMu.Lock()
	defer sMu.Unlock()
	for _, session := range validSessions {
		if session.SessionToken == sessionToken {
			return session.UserID
		}
	}
	return ""
}

func IsSessionGood(sessionToken string) bool {
	sMu.Lock()
	defer sMu.Unlock()
	for _, session := range validSessions {
		if session.SessionToken == sessionToken {
			return true
		}
	}
	return false
}

func Init() {
	go Expirerer()
}
