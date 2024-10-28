package vars

import "errors"

const CodeUserNotFound string = "user_not_found"
const CodeUserAlreadyExists string = "user_already_exists"
const CodeInvalidEmail string = "invalid_email"
const CodeBadCredentials string = "bad_credentials"
const CodeServerError string = "general_server_error"
const CodeShortPW string = "too_short_pw"
const CodeBadEmail string = "bad_email"
const CodeBadDOB string = "bad_dob"
const CodeTooManyRequests string = "too_many_requests"
const CodeSessionCertNotFound string = "session_cert_not_found"
const CodeSessionExpired string = "session_expired"

var ErrUserNotFound error = errors.New(CodeUserNotFound)
var ErrUserAlreadyExists error = errors.New(CodeUserAlreadyExists)
var ErrInvalidEmail error = errors.New(CodeInvalidEmail)
var ErrBadCredentials error = errors.New(CodeBadCredentials)
var ErrServerError error = errors.New(CodeServerError)
var ErrShortPW error = errors.New(CodeShortPW)
var ErrBadEmail error = errors.New(CodeBadEmail)
var ErrBadDOB error = errors.New(CodeBadDOB)
var ErrSessionCertNotFound error = errors.New(CodeSessionCertNotFound)
var ErrSessionExpired error = errors.New(CodeSessionExpired)
