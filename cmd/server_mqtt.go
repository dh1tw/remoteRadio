// Copyright © 2017 Tobias Wellnitz, DH1TW <Tobias.Wellnitz@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/cskr/pubsub"
	"github.com/dh1tw/remoteRadio/comms"
	"github.com/dh1tw/remoteRadio/events"
	"github.com/dh1tw/remoteRadio/ping"
	"github.com/dh1tw/remoteRadio/radio"
	"github.com/dh1tw/remoteRadio/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	hl "github.com/dh1tw/goHamlib"
	sbRadio "github.com/dh1tw/remoteRadio/sb_radio"
	sbStatus "github.com/dh1tw/remoteRadio/sb_status"
)

// serverMqttCmd represents the mqtt command
var serverMqttCmd = &cobra.Command{
	Use:   "mqtt",
	Short: "MQTT Server for a remote Radio",
	Long:  `MQTT Server for a remote Radio`,
	Run:   mqttRadioServer,
}

func init() {
	serverCmd.AddCommand(serverMqttCmd)
	serverMqttCmd.Flags().StringP("broker-url", "u", "localhost", "Broker URL")
	serverMqttCmd.Flags().IntP("broker-port", "p", 1883, "Broker Port")
	serverMqttCmd.Flags().StringP("station", "X", "mystation", "Your station callsign")
	serverMqttCmd.Flags().StringP("radio", "Y", "myradio", "Radio ID")
	serverMqttCmd.Flags().DurationP("polling_interval", "t", time.Duration(time.Millisecond*100), "Timer for polling the rig")
}

func mqttRadioServer(cmd *cobra.Command, args []string) {

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	// bind the pflags to viper settings
	viper.BindPFlag("mqtt.broker_url", cmd.Flags().Lookup("broker-url"))
	viper.BindPFlag("mqtt.broker_port", cmd.Flags().Lookup("broker-port"))
	viper.BindPFlag("mqtt.station", cmd.Flags().Lookup("station"))
	viper.BindPFlag("mqtt.radio", cmd.Flags().Lookup("radio"))
	viper.BindPFlag("radio.polling_interval", cmd.Flags().Lookup("polling_interval"))

	if viper.IsSet("general.user_id") {
		viper.Set("general.user_id", utils.RandStringRunes(5))
	} else {
		viper.Set("general.user_id", "unknown_"+utils.RandStringRunes(5))
	}

	// profiling server can be enabled through a hidden pflag
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// viper settings need to be copied in local variables
	// since viper lookups allocate of each lookup a copy
	// and are quite inperformant

	mqttBrokerURL := viper.GetString("mqtt.broker_url")
	mqttBrokerPort := viper.GetInt("mqtt.broker_port")
	mqttClientID := viper.GetString("general.user_id")

	hlDebugLevel := viper.GetInt("radio.hl-debug-level")

	baseTopic := viper.GetString("mqtt.station") +
		"/radios/" + viper.GetString("mqtt.radio") +
		"/cat"

	serverCatRequestTopic := baseTopic + "/setstate"
	serverStatusTopic := baseTopic + "/status"
	serverPingTopic := baseTopic + "/ping"
	// errorTopic := baseTopic + "/error"

	// tx topics
	serverCatResponseTopic := baseTopic + "/state"
	serverCapsTopic := baseTopic + "/caps"
	serverPongTopic := baseTopic + "/pong"

	mqttRxTopics := []string{serverCatRequestTopic, serverPingTopic}

	toWireCh := make(chan comms.IOMsg, 20)
	// toSerializeCatDataCh := make(chan comms.IOMsg, 20)
	toDeserializeCatRequestCh := make(chan []byte, 10)
	toDeserializePingRequestCh := make(chan []byte, 10)

	// Event PubSub
	evPS := pubsub.New(10)

	// WaitGroup to coordinate a graceful shutdown
	var wg sync.WaitGroup

	// mqtt Last Will Message
	binaryWillMsg, err := createLastWillMsg()
	if err != nil {
		fmt.Println(err)
	}

	lastWill := comms.LastWill{
		Topic:  serverStatusTopic,
		Data:   binaryWillMsg,
		Qos:    0,
		Retain: true,
	}

	appLogger := utils.NewStdLogger("")

	mqttSettings := comms.MqttSettings{
		WaitGroup:  &wg,
		Transport:  "tcp",
		BrokerURL:  mqttBrokerURL,
		BrokerPort: mqttBrokerPort,
		ClientID:   mqttClientID,
		Topics:     mqttRxTopics,
		ToDeserializeCatRequestCh:  toDeserializeCatRequestCh,
		ToDeserializePingRequestCh: toDeserializePingRequestCh,
		ToWire:   toWireCh,
		Events:   evPS,
		LastWill: &lastWill,
		Logger:   appLogger,
	}

	pongSettings := ping.Settings{
		PongCh:    toDeserializePingRequestCh,
		ToWireCh:  toWireCh,
		PongTopic: serverPongTopic,
		WaitGroup: &wg,
		Events:    evPS,
	}

	rigModel := viper.GetInt("radio.rig-model")

	port := hl.Port{}
	port.Baudrate = viper.GetInt("radio.baudrate")
	port.Databits = viper.GetInt("radio.databits")
	port.Stopbits = viper.GetInt("radio.stopbits")
	port.Portname = viper.GetString("radio.portname")
	port.RigPortType = hl.RIG_PORT_SERIAL
	switch viper.GetString("radio.parity") {
	case "none":
		port.Parity = hl.N
	case "even":
		port.Parity = hl.E
	case "odd":
		port.Parity = hl.O
	default:
		port.Parity = hl.N
	}

	switch viper.GetString("radio.handshake") {
	case "none":
		port.Handshake = hl.NO_HANDSHAKE
	case "RTSCTS":
		port.Handshake = hl.RTSCTS_HANDSHAKE
	default:
		port.Handshake = hl.NO_HANDSHAKE
	}

	pollingInterval := viper.GetDuration("radio.polling_interval")

	radioSettings := radio.RadioSettings{
		RigModel:         rigModel,
		Port:             port,
		HlDebugLevel:     hlDebugLevel,
		CatRequestCh:     toDeserializeCatRequestCh,
		ToWireCh:         toWireCh,
		CatResponseTopic: serverCatResponseTopic,
		CapsTopic:        serverCapsTopic,
		WaitGroup:        &wg,
		Events:           evPS,
		PollingInterval:  pollingInterval,
	}

	wg.Add(4) //MQTT + Ping + Radio + Events

	connectionStatusCh := evPS.Sub(events.MqttConnStatus)
	shutdownCh := evPS.Sub(events.Shutdown)
	prepareShutdownCh := evPS.Sub(events.PrepareShutdown)

	go events.WatchSystemEvents(evPS, &wg)
	go comms.MqttClient(mqttSettings)
	go ping.EchoPing(pongSettings)

	time.Sleep(time.Millisecond * 1300)
	go radio.HandleRadio(radioSettings)

	status := serverStatus{}
	status.topic = serverStatusTopic
	status.toWireCh = toWireCh

	for {
		select {
		case <-prepareShutdownCh:

			// publish that the server is going offline
			status.online = false
			if err := status.sendUpdate(); err != nil {
				fmt.Println(err)
			}
			time.Sleep(time.Millisecond * 500)
			// inform the other goroutines to shut down
			evPS.Pub(true, events.Shutdown)

		// shutdown the application gracefully
		case <-shutdownCh:
			//force exit after 1 sec
			exitTimeout := time.NewTimer(time.Second)
			go func() {
				<-exitTimeout.C
				fmt.Println("quitting forcefully")
				os.Exit(0)
			}()

			wg.Wait()
			os.Exit(0)

		case ev := <-connectionStatusCh:
			connStatus := ev.(int)
			fmt.Println("connstatus:", connStatus)
			if connStatus == comms.CONNECTED {
				status.online = true
				if err := status.sendUpdate(); err != nil {
					fmt.Println(err)
				}
			} else {
				status.online = false
			}
		}
	}
}

type serverStatus struct {
	online   bool
	topic    string
	toWireCh chan comms.IOMsg
}

func (s *serverStatus) sendUpdate() error {

	msg := sbStatus.Status{}
	msg.Online = s.online
	data, err := msg.Marshal()
	if err != nil {
		return err
	}

	m := comms.IOMsg{}
	m.Data = data
	m.Topic = s.topic
	m.Retain = true

	s.toWireCh <- m

	return nil
}

func createLastWillMsg() ([]byte, error) {

	willMsg := sbRadio.Status{}
	willMsg.Online = false
	data, err := willMsg.Marshal()

	return data, err
}
