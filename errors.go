package gotcha

import (
	"errors"
)

var MaxRetriesExceededError = errors.New("Maximum amount of retries exceeded.")
