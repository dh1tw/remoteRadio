package utils

import "log"

func NewChanLogger(ch chan string, prefix string) *log.Logger {

	cl := chanLogger{}
	cl.ch = ch
	logger := log.New(&cl, prefix, log.Ltime)
	return logger
}

type chanLogger struct {
	ch chan string
}

func (c *chanLogger) Write(input []byte) (n int, err error) {
	c.ch <- string(input)

	return len(input), nil
}
