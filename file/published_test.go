package file_test

import (
	"context"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/dp-static-file-publisher/file"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

type MockMessage struct {
	Data []byte
}

func (m MockMessage) GetData() []byte {
	return m.Data
}

func (m MockMessage) Mark() {
}

func (m MockMessage) Commit() {

}

func (m MockMessage) Release() {
}

func (m MockMessage) CommitAndRelease() {
}

func (m MockMessage) Offset() int64 {
	return 1
}

func (m MockMessage) UpstreamDone() chan struct{} {
	return nil
}

func TestHandleFilePublishMessage(t *testing.T) {

	Convey("Given invalid message content", t, func() {
		dc := file.DecrypterCopier{}
		ctx := context.Background()
		msg := MockMessage{
			Data: []byte("Testing"),
		}

		Convey("When the message is handled", func() {
			err := dc.HandleFilePublishMessage(ctx, 1, msg)

			So(err, ShouldBeError)

			commiter, ok := err.(kafka.Commiter)

			So(ok, ShouldBeTrue)
			So(commiter.Commit(), ShouldBeFalse)
		})
	})

}
