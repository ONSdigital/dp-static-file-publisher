// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-static-file-publisher/event"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"sync"
)

// Ensure, that S3WriterMock does implement event.S3Writer.
// If this is not the case, regenerate this file with moq.
var _ event.S3Writer = &S3WriterMock{}

// S3WriterMock is a mock implementation of event.S3Writer.
//
//	func TestSomethingThatUsesS3Writer(t *testing.T) {
//
//		// make and configure a mocked event.S3Writer
//		mockedS3Writer := &S3WriterMock{
//			BucketNameFunc: func() string {
//				panic("mock out the BucketName method")
//			},
//			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error {
//				panic("mock out the Checker method")
//			},
//			SessionFunc: func() *session.Session {
//				panic("mock out the Session method")
//			},
//			UploadFunc: func(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
//				panic("mock out the Upload method")
//			},
//		}
//
//		// use mockedS3Writer in code that requires event.S3Writer
//		// and then make assertions.
//
//	}
type S3WriterMock struct {
	// BucketNameFunc mocks the BucketName method.
	BucketNameFunc func() string

	// CheckerFunc mocks the Checker method.
	CheckerFunc func(ctx context.Context, state *healthcheck.CheckState) error

	// SessionFunc mocks the Session method.
	SessionFunc func() *session.Session

	// UploadFunc mocks the Upload method.
	UploadFunc func(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)

	// calls tracks calls to the methods.
	calls struct {
		// BucketName holds details about calls to the BucketName method.
		BucketName []struct {
		}
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
		// Upload holds details about calls to the Upload method.
		Upload []struct {
			// Input is the input argument value.
			Input *s3manager.UploadInput
			// Options is the options argument value.
			Options []func(*s3manager.Uploader)
		}
	}
	lockBucketName sync.RWMutex
	lockChecker    sync.RWMutex
	lockSession    sync.RWMutex
	lockUpload     sync.RWMutex
}

// BucketName calls BucketNameFunc.
func (mock *S3WriterMock) BucketName() string {
	if mock.BucketNameFunc == nil {
		panic("S3WriterMock.BucketNameFunc: method is nil but S3Writer.BucketName was just called")
	}
	callInfo := struct {
	}{}
	mock.lockBucketName.Lock()
	mock.calls.BucketName = append(mock.calls.BucketName, callInfo)
	mock.lockBucketName.Unlock()
	return mock.BucketNameFunc()
}

// BucketNameCalls gets all the calls that were made to BucketName.
// Check the length with:
//
//	len(mockedS3Writer.BucketNameCalls())
func (mock *S3WriterMock) BucketNameCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockBucketName.RLock()
	calls = mock.calls.BucketName
	mock.lockBucketName.RUnlock()
	return calls
}

// Checker calls CheckerFunc.
func (mock *S3WriterMock) Checker(ctx context.Context, state *healthcheck.CheckState) error {
	if mock.CheckerFunc == nil {
		panic("S3WriterMock.CheckerFunc: method is nil but S3Writer.Checker was just called")
	}
	callInfo := struct {
		Ctx   context.Context
		State *healthcheck.CheckState
	}{
		Ctx:   ctx,
		State: state,
	}
	mock.lockChecker.Lock()
	mock.calls.Checker = append(mock.calls.Checker, callInfo)
	mock.lockChecker.Unlock()
	return mock.CheckerFunc(ctx, state)
}

// CheckerCalls gets all the calls that were made to Checker.
// Check the length with:
//
//	len(mockedS3Writer.CheckerCalls())
func (mock *S3WriterMock) CheckerCalls() []struct {
	Ctx   context.Context
	State *healthcheck.CheckState
} {
	var calls []struct {
		Ctx   context.Context
		State *healthcheck.CheckState
	}
	mock.lockChecker.RLock()
	calls = mock.calls.Checker
	mock.lockChecker.RUnlock()
	return calls
}

// Session calls SessionFunc.
func (mock *S3WriterMock) Session() *session.Session {
	if mock.SessionFunc == nil {
		panic("S3WriterMock.SessionFunc: method is nil but S3Writer.Session was just called")
	}
	callInfo := struct {
	}{}
	mock.lockSession.Lock()
	mock.calls.Session = append(mock.calls.Session, callInfo)
	mock.lockSession.Unlock()
	return mock.SessionFunc()
}

// SessionCalls gets all the calls that were made to Session.
// Check the length with:
//
//	len(mockedS3Writer.SessionCalls())
func (mock *S3WriterMock) SessionCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockSession.RLock()
	calls = mock.calls.Session
	mock.lockSession.RUnlock()
	return calls
}

// Upload calls UploadFunc.
func (mock *S3WriterMock) Upload(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
	if mock.UploadFunc == nil {
		panic("S3WriterMock.UploadFunc: method is nil but S3Writer.Upload was just called")
	}
	callInfo := struct {
		Input   *s3manager.UploadInput
		Options []func(*s3manager.Uploader)
	}{
		Input:   input,
		Options: options,
	}
	mock.lockUpload.Lock()
	mock.calls.Upload = append(mock.calls.Upload, callInfo)
	mock.lockUpload.Unlock()
	return mock.UploadFunc(input, options...)
}

// UploadCalls gets all the calls that were made to Upload.
// Check the length with:
//
//	len(mockedS3Writer.UploadCalls())
func (mock *S3WriterMock) UploadCalls() []struct {
	Input   *s3manager.UploadInput
	Options []func(*s3manager.Uploader)
} {
	var calls []struct {
		Input   *s3manager.UploadInput
		Options []func(*s3manager.Uploader)
	}
	mock.lockUpload.RLock()
	calls = mock.calls.Upload
	mock.lockUpload.RUnlock()
	return calls
}
