package config

import (
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfig(t *testing.T) {
	Convey("Given an environment with no environment variables set", t, func() {
		os.Clearenv()
		cfg, err := Get()

		Convey("When the config values are retrieved", func() {

			Convey("Then there should be no error returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the values should be set to the expected defaults", func() {
				So(cfg.BindAddr, ShouldEqual, "localhost:24900")
				So(cfg.ServiceAuthToken, ShouldEqual, "4424A9F2-B903-40F4-85F1-240107D1AFAF")
				So(cfg.EncryptionDisabled, ShouldBeFalse)
				So(cfg.GracefulShutdownTimeout, ShouldEqual, 5*time.Second)
				So(cfg.HealthCheckInterval, ShouldEqual, 30*time.Second)
				So(cfg.HealthCheckCriticalTimeout, ShouldEqual, 90*time.Second)
				So(cfg.VaultToken, ShouldEqual, "")
				So(cfg.VaultAddress, ShouldEqual, "")
				So(cfg.VaultRetries, ShouldEqual, 3)
				So(cfg.VaultPath, ShouldEqual, "secret/shared/psk")
				So(cfg.ImageAPIURL, ShouldEqual, "http://localhost:24700")
				So(cfg.KafkaAddr, ShouldResemble, []string{"localhost:9092", "localhost:9093", "localhost:9094"})
				So(cfg.KafkaVersion, ShouldEqual, "1.0.2")
				So(cfg.KafkaConsumerWorkers, ShouldEqual, 1)
				So(cfg.StaticFilePublishedTopic, ShouldEqual, "static-file-published")
				So(cfg.ConsumerGroup, ShouldEqual, "dp-static-file-publisher")
				So(cfg.AwsRegion, ShouldEqual, "eu-west-1")
				So(cfg.PrivateBucketName, ShouldEqual, "csv-exported")
				So(cfg.PublicBucketName, ShouldEqual, "static-develop")
				So(cfg.PublicBucketURL, ShouldEqual, "https://static-develop.s3.eu-west-1.amazonaws.com")
			})

			Convey("Then a second call to config should return the same config", func() {
				newCfg, newErr := Get()
				So(newErr, ShouldBeNil)
				So(newCfg, ShouldResemble, cfg)
			})
		})
	})
}
