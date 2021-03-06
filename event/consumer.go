package event

import (
	"context"
	"errors"

	kafka "github.com/ONSdigital/dp-kafka"
	"github.com/ONSdigital/dp-static-file-publisher/schema"
	"github.com/ONSdigital/log.go/log"
)

//go:generate moq -out mock/handler.go -pkg mock . Handler

// MessageConsumer provides a generic interface for consuming []byte messages
type MessageConsumer interface {
	Channels() *kafka.ConsumerGroupChannels
	Release()
}

// Handler represents a handler for processing a single event.
type Handler interface {
	Handle(ctx context.Context, ImagePublished *ImagePublished) error
}

// Consumer consumes event messages.
type Consumer struct {
	closing chan bool
	closed  chan bool
}

// NewConsumer returns a new consumer instance.
func NewConsumer() *Consumer {
	return &Consumer{
		closing: make(chan bool),
		closed:  make(chan bool),
	}
}

// Consume converts messages to event instances, and pass the event to the provided handler.
func (consumer *Consumer) Consume(ctx context.Context, messageConsumer MessageConsumer, handler Handler) {

	go func() {
		defer close(consumer.closed)

		for {
			select {
			case message := <-messageConsumer.Channels().Upstream:
				messageCtx := context.Background()
				processMessage(messageCtx, message, handler)
				messageConsumer.Release()
			case <-consumer.closing:
				log.Event(ctx, "closing event consumer loop", log.INFO)
				return
			}
		}
	}()

}

// Close safely closes the consumer and releases all resources
func (consumer *Consumer) Close(ctx context.Context) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	close(consumer.closing)

	select {
	case <-consumer.closed:
		log.Event(ctx, "successfully closed event consumer", log.INFO)
		return nil
	case <-ctx.Done():
		log.Event(ctx, "shutdown context time exceeded, skipping graceful shutdown of event consumer", log.INFO)
		return errors.New("shutdown context timed out")
	}
}

// processMessage unmarshals the provided kafka message into an event and calls the handler.
// After the message is successfully handled, it is committed.
func processMessage(ctx context.Context, message kafka.Message, handler Handler) {

	event, err := unmarshal(message)
	if err != nil {
		log.Event(ctx, "failed to unmarshal event", log.ERROR, log.Error(err))
		return
	}

	log.Event(ctx, "event received", log.INFO, log.Data{"event": event})

	err = handler.Handle(ctx, event)
	if err != nil {
		log.Event(ctx, "failed to handle event", log.ERROR, log.Error(err))
	}

	log.Event(ctx, "event processed - committing message", log.INFO, log.Data{"event": event})
	message.Commit()
	log.Event(ctx, "message committed", log.INFO, log.Data{"event": event})
}

// unmarshal converts a event instance to []byte.
func unmarshal(message kafka.Message) (*ImagePublished, error) {
	var event ImagePublished
	err := schema.ImagePublishedEvent.Unmarshal(message.GetData(), &event)
	return &event, err
}
