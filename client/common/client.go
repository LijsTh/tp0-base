package common

import (
	"os"
	"context"
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
	MaxBatch      int
}

const FILEPATH = "/data/agency-"

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

// Open the file and create the reader while handling errors
func (c *Client) initialize_reader() (*BetReader) {
	file := FILEPATH + c.config.ID + ".csv"
	reader, err := NewBetReader(file, c.config.MaxBatch, c.config.ID)
	if err != nil {
		log.Criticalf("action: file_open | result: fail | error: %v", err)
		os.Exit(1)
	}
	return reader
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop(ctx context.Context, wg *sync.WaitGroup, finished_iter chan bool) {
	stopped := false
	defer wg.Done()
	reader := c.initialize_reader()
	defer reader.file.Close()

	for msgID := 1; !reader.finished && !stopped ; msgID++ {
		// Create the connection the server in every loop iteration. Send an
		err := c.createClientSocket()
		if err != nil {
			// If the connection fails, the client is closed and exit 1 is returned
			c.conn.Close()
			reader.file.Close()
			os.Exit(1)
		}

		// Wait for a signal to stop the client
		go wait_for_signal(ctx, &c.conn, finished_iter, &stopped)

		bets, err := reader.ReadBets()
		if err != nil {
			error_handler(err, "read_bets", &stopped)
			c.conn.Close()
			return
		}


		err = SendBatch(c.conn, bets)
		if err != nil {
			error_handler(err, "send_batch", &stopped)
			c.conn.Close()
			return
		}

		answer, err := RecvAnswer(c.conn)
		if err != nil {
			error_handler(err, "recv_answer", &stopped)
			c.conn.Close()
			return
		}

		if answer == SUCESS {
			log.Infof("action: enviar_apuesta | result: success | client_id: %v | cantidad: %v", c.config.ID, len(bets))
		} else {
			log.Infof("action: enviar_apuesta | result: fail | client_id: %v | cantidad: %v", c.config.ID, len(bets))
		}

		c.conn.Close()
		// Stop the go routine for the next connection
		if !stopped {finished_iter <- true}
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