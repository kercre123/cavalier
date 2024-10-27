package vars

import "errors"

const CodeUserNotFound string = "user_not_found"
const CodeUserAlreadyExists string = "user_already_exists"
const CodeInvalidEmail string = "invalid_email"
const CodeBadCredentials string = "bad_credentials"
const CodeServerError string = "general_server_error"
const CodeShortPW string = "too_short_pw"
const CodeBadEmail string = "bad_email"

var ErrUserNotFound error = errors.New(CodeUserNotFound)
var ErrUserAlreadyExists error = errors.New(CodeUserAlreadyExists)
var ErrInvalidEmail error = errors.New(CodeInvalidEmail)
var ErrBadCredentials error = errors.New(CodeBadCredentials)
var ErrServerError error = errors.New(CodeServerError)
var ErrShortPW error = errors.New(CodeShortPW)
var ErrBadEmail error = errors.New(CodeBadEmail)
