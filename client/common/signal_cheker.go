package common

import (
	"net"
	"context"
)


func wait_for_signal(ctx context.Context, connection *net.Conn, finished_iter chan bool, stopped *bool) {
	select {
		case <-ctx.Done():
			log.Infof("action: SIGTERM | result: sucess")
			err := (*connection).Close()
			if err == nil {
				log.Infof("action: close_connection | result: success")
			}
			*stopped = true

		case <-finished_iter:
			return
	}
}	