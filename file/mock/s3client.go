// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"github.com/ONSdigital/dp-static-file-publisher/file"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
	"sync"
)

// Ensure, that S3ClientV2Mock does implement file.S3ClientV2.
// If this is not the case, regenerate this file with moq.
var _ file.S3ClientV2 = &S3ClientV2Mock{}

// S3ClientV2Mock is a mock implementation of file.S3ClientV2.
//
// 	func TestSomethingThatUsesS3ClientV2(t *testing.T) {
//
// 		// make and configure a mocked file.S3ClientV2
// 		mockedS3ClientV2 := &S3ClientV2Mock{
// 			FileExistsFunc: func(key string) (bool, error) {
// 				panic("mock out the FileExists method")
// 			},
// 			GetFunc: func(key string) (io.ReadCloser, *int64, error) {
// 				panic("mock out the Get method")
// 			},
// 			UploadFunc: func(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
// 				panic("mock out the Upload method")
// 			},
// 		}
//
// 		// use mockedS3ClientV2 in code that requires file.S3ClientV2
// 		// and then make assertions.
//
// 	}
type S3ClientV2Mock struct {
	// FileExistsFunc mocks the FileExists method.
	FileExistsFunc func(key string) (bool, error)

	// GetFunc mocks the Get method.
	GetFunc func(key string) (io.ReadCloser, *int64, error)

	// UploadFunc mocks the Upload method.
	UploadFunc func(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)

	// calls tracks calls to the methods.
	calls struct {
		// FileExists holds details about calls to the FileExists method.
		FileExists []struct {
			// Key is the key argument value.
			Key string
		}
		// Get holds details about calls to the Get method.
		Get []struct {
			// Key is the key argument value.
			Key string
		}
		// Upload holds details about calls to the Upload method.
		Upload []struct {
			// Input is the input argument value.
			Input *s3manager.UploadInput
			// Options is the options argument value.
			Options []func(*s3manager.Uploader)
		}
	}
	lockFileExists sync.RWMutex
	lockGet sync.RWMutex
	lockUpload     sync.RWMutex
}

// FileExists calls FileExistsFunc.
func (mock *S3ClientV2Mock) FileExists(key string) (bool, error) {
	if mock.FileExistsFunc == nil {
		panic("S3ClientV2Mock.FileExistsFunc: method is nil but S3ClientV2.FileExists was just called")
	}
	callInfo := struct {
		Key string
	}{
		Key: key,
	}
	mock.lockFileExists.Lock()
	mock.calls.FileExists = append(mock.calls.FileExists, callInfo)
	mock.lockFileExists.Unlock()
	return mock.FileExistsFunc(key)
}

// FileExistsCalls gets all the calls that were made to FileExists.
// Check the length with:
//     len(mockedS3ClientV2.FileExistsCalls())
func (mock *S3ClientV2Mock) FileExistsCalls() []struct {
	Key string
} {
	var calls []struct {
		Key string
	}
	mock.lockFileExists.RLock()
	calls = mock.calls.FileExists
	mock.lockFileExists.RUnlock()
	return calls
}

// Get calls GetFunc.
func (mock *S3ClientV2Mock) Get(key string) (io.ReadCloser, *int64, error) {
	if mock.GetFunc == nil {
		panic("S3ClientV2Mock.GetFunc: method is nil but S3ClientV2.Get was just called")
	}
	callInfo := struct {
		Key string
	}{
		Key: key,
	}
	mock.lockGet.Lock()
	mock.calls.Get = append(mock.calls.Get, callInfo)
	mock.lockGet.Unlock()
	return mock.GetFunc(key)
}

// GetCalls gets all the calls that were made to Get.
// Check the length with:
//     len(mockedS3ClientV2.GetCalls())
func (mock *S3ClientV2Mock) GetCalls() []struct {
	Key string
} {
	var calls []struct {
		Key string
	}
	mock.lockGet.RLock()
	calls = mock.calls.Get
	mock.lockGet.RUnlock()
	return calls
}

// Upload calls UploadFunc.
func (mock *S3ClientV2Mock) Upload(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
	if mock.UploadFunc == nil {
		panic("S3ClientV2Mock.UploadFunc: method is nil but S3ClientV2.Upload was just called")
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
//     len(mockedS3ClientV2.UploadCalls())
func (mock *S3ClientV2Mock) UploadCalls() []struct {
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
