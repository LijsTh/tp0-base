package common

import (
	"context"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")
const FILEPATH = "/data/agency-"
// const FILEPATH = "../.data/agency-"

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
	MaxBatch      int
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
	file := FILEPATH + c.config.ID + ".csv"
	// file := FILEPATH + "1" + ".csv"
	reader, err := NewBetReader(file, c.config.MaxBatch, c.config.ID)
	if err != nil {
		log.Criticalf("action: file_open | result: fail | error: %v", err)
		os.Exit(1)
	}
	defer reader.file.Close()
loop: 
	for msgID := 1; !reader.finished; msgID++ {
		// Create the connection the server in every loop iteration. Send an
		err := c.createClientSocket()
		if err != nil {
			// If the connection fails, the client is closed and exit 1 is returned
			c.conn.Close()
			os.Exit(1)
		}

		bets, err := reader.ReadBets()
		if err != nil {
			log.Criticalf("action: read_bets | result: fail | error: %v", err)
			c.conn.Close()
			return
		}


		err = SendBatch(c.conn, bets)
		if err != nil {
			log.Criticalf("action: send_batch | result: fail | error: %v", err)
			c.conn.Close()
			return
		}

		err = RecvAnswer(c.conn)
		if err != nil {
			log.Criticalf("action: recv_answer | result: fail | error: %v", err)
			c.conn.Close()
			return
		}

		// Checks if the context has been cancelled
		select {
		case <-ctx.Done():
			log.Infof("action: SIGTERM Received | result: success | client_id: %v", c.config.ID)
			break loop 
		default:
			time.Sleep(c.config.LoopPeriod)
			c.conn.Close()
		}
	}
	err = c.awaitResults()
	if err != nil {
		log.Criticalf("action: consulta_ganadores | result: fail | error: %v", err)
	}


	c.conn.Close()
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}

func (c *Client) awaitResults() error {
	err := c.createClientSocket()
	if err != nil {
		// If the connection fails, the client is closed and exit 1 is returned
		os.Exit(1)
	}
	agency, _ := strconv.Atoi(c.config.ID)
	err = sendEndMessage(c.conn, agency)	
	if err != nil { return err}

	log.Info("action: awaiting results")
	// Wait for the server to send the results
	results , err := RecvResults(c.conn)
	if err != nil {
		return err	
	} else {
		log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %v", len(results))
	}

	err = sendFinish(c.conn)
	if err != nil { return err}
	return nil
}
