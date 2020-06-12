package service_test

import (
	"testing"

	"github.com/ONSdigital/dp-static-file-publisher/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRun(t *testing.T) {

	Convey("TestNothing", t, func() {
		_, err := config.Get()
		So(err, ShouldBeNil)
	})

}
