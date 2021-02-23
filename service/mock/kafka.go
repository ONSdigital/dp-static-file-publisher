// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-kafka"
	"github.com/ONSdigital/dp-static-file-publisher/service"
	"sync"
)

// Ensure, that KafkaConsumerMock does implement service.KafkaConsumer.
// If this is not the case, regenerate this file with moq.
var _ service.KafkaConsumer = &KafkaConsumerMock{}

// KafkaConsumerMock is a mock implementation of service.KafkaConsumer.
//
//     func TestSomethingThatUsesKafkaConsumer(t *testing.T) {
//
//         // make and configure a mocked service.KafkaConsumer
//         mockedKafkaConsumer := &KafkaConsumerMock{
//             ChannelsFunc: func() *kafka.ConsumerGroupChannels {
// 	               panic("mock out the Channels method")
//             },
//             CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error {
// 	               panic("mock out the Checker method")
//             },
//             CloseFunc: func(ctx context.Context) error {
// 	               panic("mock out the Close method")
//             },
//             ReleaseFunc: func()  {
// 	               panic("mock out the Release method")
//             },
//             StopListeningToConsumerFunc: func(ctx context.Context) error {
// 	               panic("mock out the StopListeningToConsumer method")
//             },
//         }
//
//         // use mockedKafkaConsumer in code that requires service.KafkaConsumer
//         // and then make assertions.
//
//     }
type KafkaConsumerMock struct {
	// ChannelsFunc mocks the Channels method.
	ChannelsFunc func() *kafka.ConsumerGroupChannels

	// CheckerFunc mocks the Checker method.
	CheckerFunc func(ctx context.Context, state *healthcheck.CheckState) error

	// CloseFunc mocks the Close method.
	CloseFunc func(ctx context.Context) error

	// ReleaseFunc mocks the Release method.
	ReleaseFunc func()

	// StopListeningToConsumerFunc mocks the StopListeningToConsumer method.
	StopListeningToConsumerFunc func(ctx context.Context) error

	// calls tracks calls to the methods.
	calls struct {
		// Channels holds details about calls to the Channels method.
		Channels []struct {
		}
		// Checker holds details about calls to the Checker method.
		Checker []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// State is the state argument value.
			State *healthcheck.CheckState
		}
		// Close holds details about calls to the Close method.
		Close []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
		// Release holds details about calls to the Release method.
		Release []struct {
		}
		// StopListeningToConsumer holds details about calls to the StopListeningToConsumer method.
		StopListeningToConsumer []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
	}
	lockChannels                sync.RWMutex
	lockChecker                 sync.RWMutex
	lockClose                   sync.RWMutex
	lockRelease                 sync.RWMutex
	lockStopListeningToConsumer sync.RWMutex
}

// Channels calls ChannelsFunc.
func (mock *KafkaConsumerMock) Channels() *kafka.ConsumerGroupChannels {
	if mock.ChannelsFunc == nil {
		panic("KafkaConsumerMock.ChannelsFunc: method is nil but KafkaConsumer.Channels was just called")
	}
	callInfo := struct {
	}{}
	mock.lockChannels.Lock()
	mock.calls.Channels = append(mock.calls.Channels, callInfo)
	mock.lockChannels.Unlock()
	return mock.ChannelsFunc()
}

// ChannelsCalls gets all the calls that were made to Channels.
// Check the length with:
//     len(mockedKafkaConsumer.ChannelsCalls())
func (mock *KafkaConsumerMock) ChannelsCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockChannels.RLock()
	calls = mock.calls.Channels
	mock.lockChannels.RUnlock()
	return calls
}

// Checker calls CheckerFunc.
func (mock *KafkaConsumerMock) Checker(ctx context.Context, state *healthcheck.CheckState) error {
	if mock.CheckerFunc == nil {
		panic("KafkaConsumerMock.CheckerFunc: method is nil but KafkaConsumer.Checker was just called")
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
//     len(mockedKafkaConsumer.CheckerCalls())
func (mock *KafkaConsumerMock) CheckerCalls() []struct {
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

// Close calls CloseFunc.
func (mock *KafkaConsumerMock) Close(ctx context.Context) error {
	if mock.CloseFunc == nil {
		panic("KafkaConsumerMock.CloseFunc: method is nil but KafkaConsumer.Close was just called")
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.lockClose.Lock()
	mock.calls.Close = append(mock.calls.Close, callInfo)
	mock.lockClose.Unlock()
	return mock.CloseFunc(ctx)
}

// CloseCalls gets all the calls that were made to Close.
// Check the length with:
//     len(mockedKafkaConsumer.CloseCalls())
func (mock *KafkaConsumerMock) CloseCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	mock.lockClose.RLock()
	calls = mock.calls.Close
	mock.lockClose.RUnlock()
	return calls
}

// Release calls ReleaseFunc.
func (mock *KafkaConsumerMock) Release() {
	if mock.ReleaseFunc == nil {
		panic("KafkaConsumerMock.ReleaseFunc: method is nil but KafkaConsumer.Release was just called")
	}
	callInfo := struct {
	}{}
	mock.lockRelease.Lock()
	mock.calls.Release = append(mock.calls.Release, callInfo)
	mock.lockRelease.Unlock()
	mock.ReleaseFunc()
}

// ReleaseCalls gets all the calls that were made to Release.
// Check the length with:
//     len(mockedKafkaConsumer.ReleaseCalls())
func (mock *KafkaConsumerMock) ReleaseCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockRelease.RLock()
	calls = mock.calls.Release
	mock.lockRelease.RUnlock()
	return calls
}

// StopListeningToConsumer calls StopListeningToConsumerFunc.
func (mock *KafkaConsumerMock) StopListeningToConsumer(ctx context.Context) error {
	if mock.StopListeningToConsumerFunc == nil {
		panic("KafkaConsumerMock.StopListeningToConsumerFunc: method is nil but KafkaConsumer.StopListeningToConsumer was just called")
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.lockStopListeningToConsumer.Lock()
	mock.calls.StopListeningToConsumer = append(mock.calls.StopListeningToConsumer, callInfo)
	mock.lockStopListeningToConsumer.Unlock()
	return mock.StopListeningToConsumerFunc(ctx)
}

// StopListeningToConsumerCalls gets all the calls that were made to StopListeningToConsumer.
// Check the length with:
//     len(mockedKafkaConsumer.StopListeningToConsumerCalls())
func (mock *KafkaConsumerMock) StopListeningToConsumerCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	mock.lockStopListeningToConsumer.RLock()
	calls = mock.calls.StopListeningToConsumer
	mock.lockStopListeningToConsumer.RUnlock()
	return calls
}