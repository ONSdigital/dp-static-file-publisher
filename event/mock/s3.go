// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-static-file-publisher/event"
	"github.com/aws/aws-sdk-go/aws/session"
	"sync"
)

var (
	lockS3ClientMockChecker sync.RWMutex
	lockS3ClientMockSession sync.RWMutex
)

// Ensure, that S3ClientMock does implement event.S3Client.
// If this is not the case, regenerate this file with moq.
var _ event.S3Client = &S3ClientMock{}

// S3ClientMock is a mock implementation of event.S3Client.
//
//     func TestSomethingThatUsesS3Client(t *testing.T) {
//
//         // make and configure a mocked event.S3Client
//         mockedS3Client := &S3ClientMock{
//             CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error {
// 	               panic("mock out the Checker method")
//             },
//             SessionFunc: func() *session.Session {
// 	               panic("mock out the Session method")
//             },
//         }
//
//         // use mockedS3Client in code that requires event.S3Client
//         // and then make assertions.
//
//     }
type S3ClientMock struct {
	// CheckerFunc mocks the Checker method.
	CheckerFunc func(ctx context.Context, state *healthcheck.CheckState) error

	// SessionFunc mocks the Session method.
	SessionFunc func() *session.Session

	// calls tracks calls to the methods.
	calls struct {
		// Checker holds details about calls to the Checker method.
		Checker []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// State is the state argument value.
			State *healthcheck.CheckState
		}
		// Session holds details about calls to the Session method.
		Session []struct {
		}
	}
}

// Checker calls CheckerFunc.
func (mock *S3ClientMock) Checker(ctx context.Context, state *healthcheck.CheckState) error {
	if mock.CheckerFunc == nil {
		panic("S3ClientMock.CheckerFunc: method is nil but S3Client.Checker was just called")
	}
	callInfo := struct {
		Ctx   context.Context
		State *healthcheck.CheckState
	}{
		Ctx:   ctx,
		State: state,
	}
	lockS3ClientMockChecker.Lock()
	mock.calls.Checker = append(mock.calls.Checker, callInfo)
	lockS3ClientMockChecker.Unlock()
	return mock.CheckerFunc(ctx, state)
}

// CheckerCalls gets all the calls that were made to Checker.
// Check the length with:
//     len(mockedS3Client.CheckerCalls())
func (mock *S3ClientMock) CheckerCalls() []struct {
	Ctx   context.Context
	State *healthcheck.CheckState
} {
	var calls []struct {
		Ctx   context.Context
		State *healthcheck.CheckState
	}
	lockS3ClientMockChecker.RLock()
	calls = mock.calls.Checker
	lockS3ClientMockChecker.RUnlock()
	return calls
}

// Session calls SessionFunc.
func (mock *S3ClientMock) Session() *session.Session {
	if mock.SessionFunc == nil {
		panic("S3ClientMock.SessionFunc: method is nil but S3Client.Session was just called")
	}
	callInfo := struct {
	}{}
	lockS3ClientMockSession.Lock()
	mock.calls.Session = append(mock.calls.Session, callInfo)
	lockS3ClientMockSession.Unlock()
	return mock.SessionFunc()
}

// SessionCalls gets all the calls that were made to Session.
// Check the length with:
//     len(mockedS3Client.SessionCalls())
func (mock *S3ClientMock) SessionCalls() []struct {
} {
	var calls []struct {
	}
	lockS3ClientMockSession.RLock()
	calls = mock.calls.Session
	lockS3ClientMockSession.RUnlock()
	return calls
}