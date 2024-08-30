package common

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   net.Conn
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config, 
	}
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	c.conn = conn
	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	// There is an autoincremental msgID to identify every message sent
	// Messages if the message amount threshold has not been surpassed\
out: 
	for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {

		// Create the connection the server in every loop iteration. Send an
		c.createClientSocket()

		// TODO: Modify the send to avoid short-write
		fmt.Fprintf(
			c.conn,
			"[CLIENT %v] Message N°%v\n",
			c.config.ID,
			msgID,
		)
			select {
			case <-ctx.Done():
				// Closes the connection if the signal arrives before writing the message
				c.conn.Close()
				log.Infof("action: close_connection | result: success | client_id: %v", c.config.ID)
				break out

			default: 
				msg, err := bufio.NewReader(c.conn).ReadString('\n')
				c.conn.Close()
				if err != nil {
					log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
						c.config.ID,
						err,
					)
					break out
				}
	
				log.Infof("action: receive_message | result: success | client_id: %v | msg: %v",
					c.config.ID,
					msg,
				)
	
				// Wait a time between sending one message and the next one
				time.Sleep(c.config.LoopPeriod)
			}
		}
	
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}
