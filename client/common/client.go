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
func (c *Client) StartClientLoop(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	// There is an autoincremental msgID to identify every message sent
	// Messages if the message amount threshold has not been surpassed\
loop: 
	for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {

		// Create the connection the server in every loop iteration. Send an
		err := c.createClientSocket()
		if err != nil {
			// If the connection fails, the client is closed and exit 1 is returned
			os.Exit(1)
		}
		// Uses env to create a new bet
		bet := NewBetFromEnv()

		err = SendBet(c.conn, bet)	
		if err != nil {
			log.Criticalf(
				"action: send_bet | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			c.conn.Close()
			return
		}
		log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v", bet.document, bet.number)

		// Receives the response from the server
		RecvAnswer(c.conn)
		c.conn.Close()

		// Checks if the context has been cancelled
		select {
		case <-ctx.Done():
			log.Infof("action: SIGTERM Received | result: success | client_id: %v", c.config.ID)
			break loop 
		default:
			c.conn.Close()
			time.Sleep(c.config.LoopPeriod)
		}
	}
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}
