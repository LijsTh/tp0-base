package common

import (
	"context"
	"net"
	"strconv"
	"sync"
	"time"
	"errors"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")
const FILEPATH = "/data/agency-"

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

// Open the file and create the reader while handling errors
func (c *Client) initialize_reader() (*BetReader) {
	file := FILEPATH + c.config.ID + ".csv"
	reader, err := NewBetReader(file, c.config.MaxBatch, c.config.ID)
	if err != nil {
		log.Criticalf("action: file_open | result: fail | error: %v", err)
		return nil
	}
	return reader
}

// StartClientLoop Send bets to the client until the reader is finished
// Then waits for the results from the server 
func (c *Client) StartClientLoop(ctx context.Context, wg *sync.WaitGroup, finished_iter chan bool) {
	stopped := false
	defer wg.Done()
	reader := c.initialize_reader()
	if reader == nil {return}
	defer reader.file.Close()
	for msgID := 1; !stopped; msgID++ {
		// Create the connection the server in every loop iteration. Send an
		err := c.createClientSocket()
		if err != nil {
			log.Infof("action: server_connect | result: fail | client_id: %v", c.config.ID)
			return
		}
		
		// Wait for a signal to stop the client
		go wait_for_signal(ctx, &c.conn, finished_iter, &stopped)

		// Send the bets to server until the reader is finished
		if !reader.finished {
			err := c.handleSendBatch(reader, &stopped)
			if err != nil {
				c.conn.Close()
				return
			}
		} else {
			c.awaitResults(&stopped)
			select {
				case <-ctx.Done():
					break
				// signal handler to stop waiting
				case finished_iter <- true:
					break
			}
			break
		}

		c.conn.Close()
		// signal handler to stop waiting
		if (!stopped) {finished_iter <- true}
	}

	c.conn.Close()
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}

// Send a message to the server to indicate that the client has finished
// The message send a 0 representing 0 bets and the agency. 
// Then waits for the server to send the results and sends a finish message
func (c *Client) awaitResults(stopped *bool) error {
	agency, _ := strconv.Atoi(c.config.ID)
	err := sendEndMessage(c.conn, agency)	
	if err != nil { return err}

	log.Info("action: awaiting results")
	// Wait for the server to send the results
	results , err := RecvResults(c.conn)
	if err != nil {
		error_handler(err, "consulta_ganadores", stopped)
		return err
	} else {
		log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %v", len(results))
	}

	err = sendFinish(c.conn)
	if err != nil { return err}
	return nil
}

// Send a batch of bets to the server
// First it read the bets from the file, then sends them to the server
// After it waits for the server to send an answer for the batch
func (c *Client) handleSendBatch(reader *BetReader, stopped *bool) error{
	bets, err := reader.ReadBets()
	if err != nil {
		error_handler(err, "read_bets", stopped)
		return err
	}

	err = SendBatch(c.conn, bets)
	if err != nil {
		error_handler(err, "send_batch", stopped)
		return err
	}

	answer, err := RecvAnswer(c.conn)
	if err != nil {
		error_handler(err, "recv_answer", stopped)
		return err
	}

	if answer == SUCESS {
		log.Infof("action: enviar_apuesta | result: success | client_id: %v | cantidad: %v", c.config.ID, len(bets))
	} else {
		log.Infof("action: enviar_apuesta | result: fail | client_id: %v | cantidad: %v", c.config.ID, len(bets))
	}

	return nil
}

 
func error_handler(err error, message string, stopped *bool) {
	if errors.Is(err ,net.ErrClosed) {return}
	if !(*stopped) {
		log.Criticalf(
			"action: %s | result: fail | error: %v",
			message,
			err,
		)
	}
} 


