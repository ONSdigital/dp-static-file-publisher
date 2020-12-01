package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"

	kafka "github.com/ONSdigital/dp-kafka"
	"github.com/ONSdigital/dp-static-file-publisher/config"
	"github.com/ONSdigital/dp-static-file-publisher/event"
	"github.com/ONSdigital/dp-static-file-publisher/schema"
	"github.com/ONSdigital/log.go/log"
)

const serviceName = "dp-static-file-publisher"

func main() {
	log.Namespace = serviceName
	ctx := context.Background()

	// Get Config
	config, err := config.Get()
	if err != nil {
		log.Event(ctx, "error getting config", log.FATAL, log.Error(err))
		os.Exit(1)
	}

	// Create Kafka Producer
	pChannels := kafka.CreateProducerChannels()
	kafkaProducer, err := kafka.NewProducer(ctx, config.KafkaAddr, config.StaticFilePublishedTopic, 0, pChannels)
	if err != nil {
		log.Event(ctx, "fatal error trying to create kafka producer", log.FATAL, log.Error(err), log.Data{"topic": config.StaticFilePublishedTopic})
		os.Exit(1)
	}

	// kafka error logging go-routines
	kafkaProducer.Channels().LogErrors(ctx, "kafka producer")

	time.Sleep(500 * time.Millisecond)
	scanner := bufio.NewScanner(os.Stdin)
	for {
		e := scanEvent(scanner)
		log.Event(ctx, "sending image published event", log.INFO, log.Data{"imagePublishedEvent": e})

		bytes, err := schema.ImagePublishedEvent.Marshal(e)
		if err != nil {
			log.Event(ctx, "image published event error", log.FATAL, log.Error(err))
			os.Exit(1)
		}

		// Send bytes to Output channel, after calling Initialise just in case it is not initialised.
		kafkaProducer.Initialise(ctx)
		kafkaProducer.Channels().Output <- bytes
	}
}

// scanEvent creates an ImagePublished event according to the user input
func scanEvent(scanner *bufio.Scanner) *event.ImagePublished {
	fmt.Println("--- [Send Kafka ImagePublished] ---")

	fmt.Println("1 - Please type the source path")
	fmt.Printf("$ ")
	scanner.Scan()
	srcPath := scanner.Text()

	fmt.Println("2 - Please type the destination path")
	fmt.Printf("$ ")
	scanner.Scan()
	dstPath := scanner.Text()

	fmt.Println("3 - Please type the image ID")
	fmt.Printf("$ ")
	scanner.Scan()
	imageID := scanner.Text()

	// NB right now, only 'original' variant is supported so don't bother prompting for one
	variant := "original"

	return &event.ImagePublished{
		SrcPath:      srcPath,
		DstPath:      dstPath,
		ImageID:      imageID,
		ImageVariant: variant,
	}
}
