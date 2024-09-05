package common

import (
	"context"
	"net"
	"os"
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
			// If the connection fails, the client is closed and exit 1 is returned
			c.conn.Close()
			os.Exit(1)
		}

		// Wait for a signal to stop the client
		go wait_for_signal(ctx, &c.conn, channel, &stopped)

		// Uses env to create a new bet
		bet := NewBetFromEnv()

		err = SendBet(c.conn, bet)	
		if err != nil {
			error_handler(err, "apuesta_enviada", &stopped)
			c.conn.Close()
			return
		}


		// Receives the response from the server
		answer, err := RecvAnswer(c.conn)
		if err != nil {
			error_handler(err, "apuesta_enviada", &stopped)
			c.conn.Close()
			return
		}

		if answer == SUCESS {
			log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v", bet.document, bet.number)		
		} else {
			log.Criticalf("action: apuesta_enviada | result: fail | dni: %v | numero: %v", bet.document, bet)
		}

		
		c.conn.Close()
		time.Sleep(c.config.LoopPeriod)
		if !stopped {
			// Makes the go routine stop waiting for signals 
			channel <- true
		}
	}
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}

func error_handler(err error, message string, stopped *bool) {
	if !(*stopped) {
		log.Criticalf(
			"action: %s | result: fail | error: %v",
			message,
			err,
		)
	}

}