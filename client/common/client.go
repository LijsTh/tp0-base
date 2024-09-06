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
		return err
	}
	c.conn = conn
	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop(ctx context.Context, wg *sync.WaitGroup, channel chan bool) {
	defer wg.Done()
	stopped := false
	// There is an autoincremental msgID to identify every message sent
	// Messages if the message amount threshold has not been surpassed\
	for msgID := 1; msgID <= c.config.LoopAmount && !stopped; msgID++ {

		// Create the connection the server in every loop iteration. Send an
		err := c.createClientSocket()
		if err != nil {
			log.Errorf("action: create_client_socket | result: fail | client_id: %v | error: %v",)
			return
		}

		// wait for signal for current iteration
		go wait_for_signal(ctx, &c.conn, channel, &stopped)

		// Sends a message to the server and waits for a response
		_, err1 := fmt.Fprintf(
			c.conn,
			"[CLIENT %v] Message N°%v\n",
			c.config.ID,
			msgID,
		)

		msg, err2 := bufio.NewReader(c.conn).ReadString('\n')

		if err1 != nil || err2 != nil {
			log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return 
		} else {
			log.Infof("action: receive_message | result: success | client_id: %v | msg: %v",
				c.config.ID,
				msg,
			)
		}
		c.conn.Close()
		time.Sleep(c.config.LoopPeriod)
		if !stopped {
			channel <- true
		}
	}
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}
