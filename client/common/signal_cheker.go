package common

import (
	"net"
	"context"
)


func wait_for_signal(ctx context.Context, connection *net.Conn, finished_iter chan bool, stopped *bool) {
	// Wait for a signal to stop the client
	// If the signal is received, close the connection and set the stopped flag to true
	// If the connection finished before the signal received by the channel, return
	select {
		case <-ctx.Done():
			log.Infof("action: SIGTERM | result: success")
			err := (*connection).Close()
			if err == nil {
				log.Infof("action: close_connection | result: success")
			}
			*stopped = true

		case <-finished_iter:
			return
	}
}	