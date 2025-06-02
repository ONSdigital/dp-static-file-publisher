package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"

	kafka "github.com/ONSdigital/dp-kafka/v4"
	"github.com/ONSdigital/dp-static-file-publisher/config"
	"github.com/ONSdigital/dp-static-file-publisher/event"
	"github.com/ONSdigital/dp-static-file-publisher/schema"
	"github.com/ONSdigital/log.go/v2/log"
)

const serviceName = "dp-static-file-publisher"

func main() {
	log.Namespace = serviceName
	ctx := context.Background()

	// Get Config
	cfg, err := config.Get()
	if err != nil {
		log.Error(ctx, "error getting config", err)
		os.Exit(1)
	}

	// Create Kafka Producer
	pConfig := &kafka.ProducerConfig{
		BrokerAddrs:     cfg.KafkaConfig.Addr,
		Topic:           cfg.KafkaConfig.ImageFilePublishedTopic,
		KafkaVersion:    &cfg.KafkaConfig.Version,
		MaxMessageBytes: &cfg.KafkaConfig.MaxBytes,
	}
	if cfg.KafkaConfig.SecProtocol == "TLS" {
		pConfig.SecurityConfig = kafka.GetSecurityConfig(
			cfg.KafkaConfig.SecCACerts,
			cfg.KafkaConfig.SecClientCert,
			cfg.KafkaConfig.SecClientKey,
			cfg.KafkaConfig.SecSkipVerify,
		)
	}

	kafkaProducer, err := kafka.NewProducer(ctx, pConfig)
	if err != nil {
		log.Error(ctx, "fatal error trying to create kafka producer", err, log.Data{"topic": cfg.ImageFilePublishedTopic})
		os.Exit(1)
	}

	// kafka error logging go-routines
	kafkaProducer.LogErrors(ctx)

	time.Sleep(500 * time.Millisecond)
	scanner := bufio.NewScanner(os.Stdin)
	for {
		e := scanEvent(scanner)
		log.Info(ctx, "sending image published event", log.Data{"imagePublishedEvent": e})

		bytes, err := schema.ImagePublishedEvent.Marshal(e)
		if err != nil {
			log.Error(ctx, "image published event error", err)
			os.Exit(1)
		}

		// Send bytes to Output channel, after calling Initialise just in case it is not initialised.
		kafkaProducer.Initialise(ctx) //nolint
		kafkaProducer.Channels().Output <- kafka.BytesMessage{Value: bytes, Context: ctx}
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
