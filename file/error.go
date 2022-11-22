package file

import (
	"context"

	"github.com/ONSdigital/log.go/v2/log"
)

type CommitError struct {
	err error
}

func (n CommitError) Commit() bool {
	return true
}

func (n CommitError) Error() string {
	return n.err.Error()
}

func NewCommitError(ctx context.Context, err error, event string, logData log.Data) CommitError {
	log.Error(ctx, event, err, logData)
	return CommitError{err}
}
