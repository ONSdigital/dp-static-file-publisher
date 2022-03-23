package file

import (
	"context"
	"github.com/ONSdigital/log.go/v2/log"
)

type NoCommitError struct {
	err error
}

func (n NoCommitError) Commit() bool {
	return false
}

func (n NoCommitError) Error() string {
	return n.err.Error()
}

func NewNoCommitError(ctx context.Context, err error, event string, logData log.Data) NoCommitError {
	log.Error(ctx, event, err, logData)
	return NoCommitError{err}
}
