package storage

import (
	"errors"
)

var ErrLoginDuplicate = errors.New("login already exists")
var ErrNoAuthData = errors.New("no auth data")
