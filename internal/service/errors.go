package service

import "errors"

var (
	ErrItemNotFound     = errors.New("item not found")
	ErrItemAlreadyExist = errors.New("item already exists")
	// todo ErrInvalidSteamURL
)
