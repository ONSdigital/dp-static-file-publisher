package event_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	kafka "github.com/ONSdigital/dp-kafka"
	"github.com/ONSdigital/dp-kafka/kafkatest"
	"github.com/ONSdigital/dp-static-file-publisher/event"
	"github.com/ONSdigital/dp-static-file-publisher/event/mock"
	"github.com/ONSdigital/dp-static-file-publisher/schema"
	. "github.com/smartystreets/goconvey/convey"
)

var testCtx = context.Background()

var errHandler = errors.New("Handler Error")

var testEvent = event.ImagePublished{
	SrcPath: "images/ID1/original/private.png",
	DstPath: "images/ID1/original/public.png",
}

// kafkaStubConsumer mock which exposes Channels function returning empty channels
// to be used on tests that are not supposed to receive any kafka message
var kafkaStubConsumer = &kafkatest.IConsumerGroupMock{
	ChannelsFunc: func() *kafka.ConsumerGroupChannels {
		return &kafka.ConsumerGroupChannels{}
	},
}

func TestConsume(t *testing.T) {

	Convey("Given an event consumer", t, func() {

		kafkaConsumerWg := &sync.WaitGroup{}
		cgChannels := &kafka.ConsumerGroupChannels{Upstream: make(chan kafka.Message, 2)}
		mockConsumer := &kafkatest.IConsumerGroupMock{
			ChannelsFunc: func() *kafka.ConsumerGroupChannels { return cgChannels },
			ReleaseFunc:  func() { kafkaConsumerWg.Done() },
		}

		handlerWg := &sync.WaitGroup{}
		mockEventHandler := &mock.HandlerMock{
			HandleFunc: func(ctx context.Context, event *event.ImagePublished) error {
				defer handlerWg.Done()
				return nil
			},
		}
		consumer := event.NewConsumer()

		Convey("And a kafka message with the valid schema being sent to the Upstream channel", func() {

			message := kafkatest.NewMessage(marshal(testEvent), 0)
			mockConsumer.Channels().Upstream <- message

			Convey("When consume message is called", func() {

				handlerWg.Add(1)
				kafkaConsumerWg.Add(1)
				consumer.Consume(testCtx, mockConsumer, mockEventHandler)
				handlerWg.Wait()

				Convey("An event is sent to the mockEventHandler ", func() {
					So(len(mockEventHandler.HandleCalls()), ShouldEqual, 1)
					So(*mockEventHandler.HandleCalls()[0].ImagePublished, ShouldResemble, testEvent)
				})

				Convey("The message is committed and the consumer is released", func() {
					kafkaConsumerWg.Wait()
					So(len(message.CommitCalls()), ShouldEqual, 1)
					So(len(mockConsumer.ReleaseCalls()), ShouldEqual, 1)
				})
			})
		})

		Convey("And two kafka messages, one with a valid schema and one with an invalid schema", func() {

			validMessage := kafkatest.NewMessage(marshal(testEvent), 1)
			invalidMessage := kafkatest.NewMessage([]byte("invalid schema"), 0)
			mockConsumer.Channels().Upstream <- invalidMessage
			mockConsumer.Channels().Upstream <- validMessage

			Convey("When consume messages is called", func() {

				handlerWg.Add(1)
				kafkaConsumerWg.Add(2)
				consumer.Consume(testCtx, mockConsumer, mockEventHandler)
				handlerWg.Wait()

				Convey("Only the valid event is sent to the mockEventHandler ", func() {
					So(len(mockEventHandler.HandleCalls()), ShouldEqual, 1)
					So(*mockEventHandler.HandleCalls()[0].ImagePublished, ShouldResemble, testEvent)
				})

				Convey("Only the valid message is committed, but the consumer is released for both messages", func() {
					kafkaConsumerWg.Wait()
					So(len(validMessage.CommitCalls()), ShouldEqual, 1)
					So(len(invalidMessage.CommitCalls()), ShouldEqual, 0)
					So(len(mockConsumer.ReleaseCalls()), ShouldEqual, 2)
				})
			})
		})

		Convey("With a failing handler and a kafka message with the valid schema being sent to the Upstream channel", func() {
			mockEventHandler.HandleFunc = func(ctx context.Context, event *event.ImagePublished) error {
				defer handlerWg.Done()
				return errHandler
			}
			message := kafkatest.NewMessage(marshal(testEvent), 0)
			mockConsumer.Channels().Upstream <- message

			Convey("When consume message is called", func() {

				handlerWg.Add(1)
				kafkaConsumerWg.Add(1)
				consumer.Consume(testCtx, mockConsumer, mockEventHandler)
				handlerWg.Wait()

				Convey("An event is sent to the mockEventHandler ", func() {
					So(len(mockEventHandler.HandleCalls()), ShouldEqual, 1)
					So(*mockEventHandler.HandleCalls()[0].ImagePublished, ShouldResemble, testEvent)
				})

				Convey("The message is committed and the consumer is released", func() {
					kafkaConsumerWg.Wait()
					So(len(message.CommitCalls()), ShouldEqual, 1)
					So(len(mockConsumer.ReleaseCalls()), ShouldEqual, 1)
					// TODO in this case, once we have an error reported, we should validate that the error is correctly reported.
				})
			})
		})
	})
}

func TestClose(t *testing.T) {

	Convey("Given a consumer", t, func() {
		consumer := event.NewConsumer()
		consumer.Consume(testCtx, kafkaStubConsumer, &mock.HandlerMock{})

		Convey("When close is called", func() {
			err := consumer.Close(context.Background())

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

// marshal helper method to marshal a event into a []byte
func marshal(event event.ImagePublished) []byte {
	bytes, err := schema.ImagePublishedEvent.Marshal(event)
	So(err, ShouldBeNil)
	return bytes
}
