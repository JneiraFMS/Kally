package application

import "errors"

// ErrAlreadyExists indicates a create request conflicts with an existing ID.
var ErrAlreadyExists = errors.New("kb: already exists")
