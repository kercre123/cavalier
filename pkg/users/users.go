package users

import (
	"cavalier/pkg/vars"
	"database/sql"
	"errors"
	"regexp"
	"strings"
	"sync"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB
var dbMutex sync.Mutex

func Init(dbConn *sql.DB) {
	db = dbConn

	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS cavalier_users (
			uuid TEXT PRIMARY KEY,
			userid TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			hashed_pw TEXT NOT NULL,
			date_of_birth TEXT NOT NULL
		);
	`)
	if err != nil {
		panic("failed to initialize users table: " + err.Error())
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS user_robots (
			esn TEXT NOT NULL,
			user_id TEXT NOT NULL,
			PRIMARY KEY (esn, user_id),
			FOREIGN KEY (user_id) REFERENCES cavalier_users(userid)
		);
	`)
	if err != nil {
		panic("failed to initialize user_robots table: " + err.Error())
	}
}

func GetUUIDFromEmail(email string) (string, error) {
	var userUUID string
	err := db.QueryRow("SELECT uuid FROM cavalier_users WHERE email = ?", email).Scan(&userUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", vars.ErrUserNotFound
		}
		return "", err
	}

	return userUUID, nil
}

func GetUserFromUUID(uuid string) (vars.UserInDB, error) {
	var user vars.UserInDB
	err := db.QueryRow("SELECT uuid, userid, email, hashed_pw, date_of_birth FROM cavalier_users WHERE uuid = ?", uuid).Scan(&user.UUID, &user.UserID, &user.Email, &user.HashedPW, &user.DOB)
	if err != nil {
		if err == sql.ErrNoRows {
			return vars.UserInDB{}, vars.ErrUserNotFound
		}
		return vars.UserInDB{}, err
	}

	user.ESNs, _ = GetESNsForUser(user.UserID)

	return user, nil
}

func GetESNsForUser(userID string) ([]string, error) {
	rows, err := db.Query("SELECT esn FROM user_robots WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var esns []string
	for rows.Next() {
		var esn string
		if err := rows.Scan(&esn); err != nil {
			return nil, err
		}
		esns = append(esns, esn)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return esns, nil
}

func AuthUser(email string, password string) (vars.UserInDB, error) {
	if email == "" || password == "" {
		return vars.UserInDB{}, vars.ErrBadCredentials
	}
	user, err := getUser(email)
	if err != nil {
		return vars.UserInDB{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.HashedPW), []byte(password)) == nil {
		return user, nil
	}
	return vars.UserInDB{}, vars.ErrBadCredentials
}

func getUser(email string) (vars.UserInDB, error) {
	dbMutex.Lock()
	defer dbMutex.Unlock()
	uuid, err := GetUUIDFromEmail(email)
	if err != nil {
		return vars.UserInDB{}, err
	}
	user, err := GetUserFromUUID(uuid)
	if err != nil {
		return vars.UserInDB{}, err
	}
	return user, nil
}

func ValidatePassword(pw string) error {
	if len([]rune(pw)) < 8 {
		return vars.ErrShortPW
	}
	return nil
}

func ValidateEmail(email string) error {
	if len([]rune(email)) < 8 {
		return vars.ErrBadEmail
	}
	if !strings.Contains(email, "@") {
		return vars.ErrBadEmail
	}
	return nil
}

func ValidateDOB(dob string) error {
	matched, err := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, dob)
	if err != nil || !matched {
		return vars.ErrBadDOB
	}
	return nil
}

func CreateUser(email, password, dateOfBirth string) error {
	pwErr := ValidatePassword(password)
	if pwErr != nil {
		return pwErr
	}
	emailErr := ValidateEmail(email)
	if emailErr != nil {
		return emailErr
	}
	dobErr := ValidateDOB(dateOfBirth)
	if dobErr != nil {
		return dobErr
	}
	if _, err := getUser(email); err == nil {
		return vars.ErrUserAlreadyExists
	}
	pw, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return errors.New("CreateUser: failed to generate password hash: " + err.Error())
	}
	dbMutex.Lock()
	defer dbMutex.Unlock()

	_, err = db.Exec("INSERT INTO cavalier_users (uuid, userid, email, hashed_pw, date_of_birth) VALUES (?, ?, ?, ?, ?)", uuid.New().String(), vars.GenerateID(), email, string(pw), dateOfBirth)
	if err != nil {
		return errors.New("CreateUser: failed to insert user into db: " + err.Error())
	}

	return nil
}

func ResetPassword(email, oldPassword, newPassword string) error {
	user, err := getUser(email)
	if err != nil {
		return err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.HashedPW), []byte(oldPassword)) != nil {
		return vars.ErrBadCredentials
	}

	pwErr := ValidatePassword(newPassword)
	if pwErr != nil {
		return pwErr
	}

	newHashedPw, err := bcrypt.GenerateFromPassword([]byte(newPassword), 10)
	if err != nil {
		return errors.New("ResetPassword: failed to generate new password hash: " + err.Error())
	}

	dbMutex.Lock()
	defer dbMutex.Unlock()

	_, err = db.Exec("UPDATE cavalier_users SET hashed_pw = ? WHERE email = ?", string(newHashedPw), email)
	if err != nil {
		return errors.New("ResetPassword: failed to update password: " + err.Error())
	}

	return nil
}

func RemoveUser(email string) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	result, err := db.Exec("DELETE FROM cavalier_users WHERE email = ?", email)
	if err != nil {
		return errors.New("RemoveUser: failed to delete user: " + err.Error())
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.New("RemoveUser: failed to check rows affected: " + err.Error())
	}

	if rowsAffected == 0 {
		return vars.ErrUserNotFound
	}

	return nil
}

func AssociateRobotWithAccount(esn string, userID string) error {
	if userID == "notauser" {
		return nil
	}
	dbMutex.Lock()
	defer dbMutex.Unlock()

	// Check if the user exists
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM cavalier_users WHERE userid = ?", userID).Scan(&count)
	if err != nil || count == 0 {
		return errors.New("AssociateRobotWithAccount: user not found")
	}

	_, err = db.Exec("INSERT INTO user_robots (esn, user_id) VALUES (?, ?)", esn, userID)
	if err != nil {
		return errors.New("AssociateRobotWithAccount: failed to associate robot with user: " + err.Error())
	}

	return nil
}

func IsRobotAssociatedWithAccount(esn string, userID string) bool {
	if userID == "notauser" {
		return true
	}
	dbMutex.Lock()
	defer dbMutex.Unlock()

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM user_robots WHERE esn = ? AND user_id = ?", esn, userID).Scan(&count)
	if err != nil {
		return false
	}

	return count > 0
}
