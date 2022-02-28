package steps

import (
	"context"
	"net/http"
	"time"

	"github.com/ONSdigital/dp-static-file-publisher/config"

	kafka "github.com/ONSdigital/dp-kafka/v3"

	componenttest "github.com/ONSdigital/dp-component-test"
	dphttp "github.com/ONSdigital/dp-net/v2/http"
	"github.com/ONSdigital/dp-static-file-publisher/service"
	"github.com/ONSdigital/log.go/v2/log"
)

type FilePublisherComponent struct {
	DpHttpServer *dphttp.Server
	svc          *service.Service
	svcList      service.Initialiser
	ApiFeature   *componenttest.APIFeature
	errChan      chan error
	cg           *kafka.ConsumerGroup
}

const (
	localStackHost = "http://localstack:4566"
)

func NewFilePublisherComponent() *FilePublisherComponent {
	s := dphttp.NewServer("", http.NewServeMux())
	s.HandleOSSignals = false

	d := &FilePublisherComponent{
		DpHttpServer: s,
		errChan:      make(chan error),
	}

	log.Namespace = "dp-static-file-publisher"

	d.svcList = &fakeServiceContainer{s}

	return d
}

func (d *FilePublisherComponent) Initialiser() (http.Handler, error) {
	cfg, _ := config.Get()
	d.svc, _ = service.Run(context.Background(), cfg, service.NewServiceList(d.svcList), "0", "0", "1.0.0", d.errChan)
	time.Sleep(500 * time.Millisecond)

	return d.DpHttpServer.Handler, nil
}

func (d *FilePublisherComponent) Reset() {
}

func (d *FilePublisherComponent) Close() error {
	//if d.cg != nil {
	//	d.cg.Stop()
	//}
	//
	//cfg, _ := config.Get()
	//
	//if d.svc != nil {
	//	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	//	return d.svc.Close(ctx, cfg.GracefulShutdownTimeout)
	//}
	return nil
}
