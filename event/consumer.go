package event

import (
	"context"

	dpkafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-static-file-publisher/schema"
	"github.com/ONSdigital/log.go/v2/log"
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
				logData := log.Data{"message_offset": message.Offset(), "worker_num": workerNum}
				if !ok {
					log.Info(ctx, "upstream channel closed - closing event consumer loop", logData)
					return
				}

				err := processMessage(ctx, message, handler)
				if err != nil {
					log.Error(ctx, "failed to process message", err, logData)
				}

				log.Info(ctx, "message committed", logData)
				message.Release()
				log.Info(ctx, "message released", logData)

			case <-consumerGroup.Channels().Closer:
				log.Info(ctx, "closing event consumer loop because closer channel is closed", log.Data{"worker_num": workerNum})
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
		log.Error(ctx, "failed to unmarshal event", err)
		return err
	}

	log.Info(ctx, "event received", log.Data{"event": event})

	err = handler.Handle(ctx, event)
	if err != nil {
		log.Error(ctx, "failed to handle event", err)
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
