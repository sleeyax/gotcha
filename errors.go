package gotcha

import "fmt"

type MaxRetriesExceededError struct{}

func (e *MaxRetriesExceededError) Error() string {
	return "Request failed and maximum amount of retries exceeded."
}

func NewMaxRetriesExceededError() error {
	return &MaxRetriesExceededError{}
}

type MaxRedirectsExceededError struct {
	redirects int
}

func (e *MaxRedirectsExceededError) Error() string {
	return fmt.Sprintf("Maximum amount of redirects reached. Redirected %d time(s)", e.redirects)
}

func NewMaxRedirectsExceededError(redirects int) error {
	return &MaxRedirectsExceededError{redirects}
}
