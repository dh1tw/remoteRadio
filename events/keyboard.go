package events

import (
	"bufio"
	"fmt"
	"os"

	"github.com/cskr/pubsub"
)

func CaptureKeyboard(evPS *pubsub.PubSub) {

	scanner := bufio.NewScanner(os.Stdin)

	for {
		if scanner.Scan() {
			switch scanner.Text() {
			default:
				fmt.Println("keyboard input:", scanner.Text())
			}
		}
	}
}
