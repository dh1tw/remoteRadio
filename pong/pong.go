package pong

import (
	"sync"

	"github.com/cskr/pubsub"
	"github.com/dh1tw/remoteRadio/comms"
	"github.com/dh1tw/remoteRadio/events"
)

type Settings struct {
	PingRxCh  chan []byte
	ToWireCh  chan comms.IOMsg
	PongTopic string
	WaitGroup *sync.WaitGroup
	Events    *pubsub.PubSub
}

//HandlePong echos a ping request
func HandlePong(ps Settings) {

	defer ps.WaitGroup.Done()

	shutdownCh := ps.Events.Sub(events.Shutdown)

	for {
		select {
		case <-shutdownCh:
			return

		case msg := <-ps.PingRxCh:
			pongReply := comms.IOMsg{}
			pongReply.Data = msg
			pongReply.Topic = ps.PongTopic
			ps.ToWireCh <- pongReply
		}
	}
}
