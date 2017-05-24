package gus

import "strings"

var (
	ErrNotAuth         = &NotAuthenticatedError{}
	ErrNotFound        = &NotFoundError{}
	ErrCantDeleteSelf  = ErrInvalid("You can't delete yourself.")
	ErrCantSuspendSelf = ErrInvalid("You can't suspend yourself.")
	ErrTokenExpired = ErrInvalid("That access token has expired.")
)

type NotAuthenticatedError struct {
}

func (n *NotAuthenticatedError) Error() string {
	return "Not Authenticated"
}

type RateLimitExceededError struct {
	Messages []string `json:"messages"`
}

func (rl *RateLimitExceededError) Error() string {
	return strings.Join(rl.Messages, "\n- ")
}

type NotFoundError struct {
}

func (n *NotFoundError) Error() string {
	return "Not found"
}

func ErrInvalid(messages ...string) error {
	return &ValidationError{Messages: messages}
}

type ValidationError struct {
	Messages []string `json:"messages"`
}

func (v *ValidationError) Error() string {
	return strings.Join(v.Messages, "\n- ")
}
