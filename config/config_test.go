package config

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfig(t *testing.T) {
	Convey("Given an environment with no environment variables set", t, func() {
		cfg, err := Get()

		Convey("When the config values are retrieved", func() {

			Convey("Then there should be no error returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the values should be set to the expected defaults", func() {
				So(cfg.BindAddr, ShouldEqual, ":24900")
				So(cfg.GracefulShutdownTimeout, ShouldEqual, 5*time.Second)
				So(cfg.HealthCheckInterval, ShouldEqual, 30*time.Second)
				So(cfg.HealthCheckCriticalTimeout, ShouldEqual, 90*time.Second)
				So(cfg.VaultToken, ShouldEqual, "")
				So(cfg.VaultAddress, ShouldEqual, "")
				So(cfg.VaultRetries, ShouldEqual, 3)
				So(cfg.ImageAPIURL, ShouldEqual, "http://localhost:24700")
				So(cfg.KafkaAddr, ShouldResemble, []string{"localhost:9092"})
				So(cfg.StaticFilePublishedTopic, ShouldEqual, "static-file-published")
				So(cfg.ConsumerGroup, ShouldEqual, "dp-static-file-publisher")
				So(cfg.AwsRegion, ShouldEqual, "eu-west-1")
				So(cfg.PrivateBucketName, ShouldEqual, "csv-exported")
				So(cfg.PublicBucketName, ShouldEqual, "static-develop")
			})

			Convey("Then a second call to config should return the same config", func() {
				newCfg, newErr := Get()
				So(newErr, ShouldBeNil)
				So(newCfg, ShouldResemble, cfg)
			})
		})
	})
}
