package event

import (
	"context"

	dpkafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-static-file-publisher/schema"
	"github.com/ONSdigital/log.go/log"
)

//go:generate moq -out mock/handler.go -pkg mock . Handler

// Handler represents a handler for processing a single event.
type Handler interface {
	Handle(ctx context.Context, ImagePublished *ImagePublished) error
}

// Consume converts messages to event instances, and pass the event to the provided handler.
func Consume(ctx context.Context, consumerGroup dpkafka.IConsumerGroup, handler Handler, numWorkers int) {

	workerConsume := func(workerNum int) {
		for {
			select {
			case message, ok := <-consumerGroup.Channels().Upstream:
				logData := log.Data{"message_offset": message.Offset(), "workers": workerNum}
				if !ok {
					log.Event(ctx, "upstream channel closed - closing event consumer loop", log.INFO, logData)
					return
				}

				err := processMessage(ctx, message, handler)
				if err != nil {
					log.Event(ctx, "failed to process message", log.ERROR, log.Error(err), logData)
				}

				log.Event(ctx, "message committed", log.INFO, logData)
				message.Release()
				log.Event(ctx, "message released", log.INFO, logData)

			case <-consumerGroup.Channels().Closer:
				log.Event(ctx, "closing event consumer loop", log.INFO)
				return
			}
		}
	}

	// workers to consume messages in parallel
	for w := 1; w <= numWorkers; w++ {
		go workerConsume(w)
	}
}

// processMessage unmarshals the provided kafka message into an event and calls the handler.
// After the message is successfully handled, it is committed.
func processMessage(ctx context.Context, message dpkafka.Message, handler Handler) (err error) {
	defer message.Commit() //deferring the commit till the processing is done

	event, err := unmarshal(message)
	if err != nil {
		log.Event(ctx, "failed to unmarshal event", log.ERROR, log.Error(err))
		return err
	}

	log.Event(ctx, "event received", log.INFO, log.Data{"event": event})

	err = handler.Handle(ctx, event)
	if err != nil {
		log.Event(ctx, "failed to handle event", log.ERROR, log.Error(err))
		return err
	}

	return nil
}

// unmarshal converts a event instance to []byte.
func unmarshal(message dpkafka.Message) (*ImagePublished, error) {
	var event ImagePublished
	err := schema.ImagePublishedEvent.Unmarshal(message.GetData(), &event)
	return &event, err
}
