package events

import (
	"os"
	"os/signal"

	"github.com/cskr/pubsub"
)

// Event channel names used for event Pubsub

// internal
const (
	MqttConnStatus    = "mqttConnStatus"    // int
	ForwardCat        = "forwardAudio"      //bool
	Shutdown          = "shutdown"          // bool
	OsExit            = "osExit"            // bool
)

// for message handling
const (
	ServerOnline         = "serverOnline" //bool
	Ping                 = "ping"         // int64
)

func WatchSystemEvents(evPS *pubsub.PubSub) {

	// Channel to handle OS signals
	osSignals := make(chan os.Signal, 1)

	//subscribe to os.Interrupt (CTRL-C signal)
	signal.Notify(osSignals, os.Interrupt)

	select {
	case osSignal := <-osSignals:
		if osSignal == os.Interrupt {
			evPS.Pub(true, OsExit)
		}
	}
}
