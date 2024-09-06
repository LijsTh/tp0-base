package common

import (
	"encoding/binary"
	"errors"
	"net"
)

const MAX_STR_SIZE = 255
const DOCUMENT_SIZE = 4
const NUMBER_SIZE = 2
const ANSWER_SIZE = 1
const BATCH_SIZE = 2
const MAX_BATCH_BYTES = 8000 // 8KB

const SUCESS = 0
const FAIL = 1
/// Read all the bytes from the connection to avoid partial reads
func RecvAll(conn net.Conn, size int) ([]byte, error) {
    buf := make([]byte, size)
    total := 0

    for total < size {
        n, err := conn.Read(buf[total:])
        if err != nil || n == 0 {
            return nil, err
        }
        total += n
    }
    return buf, nil
}

// Send all the bytes from the message to avoid partial writes
func send_all(conn net.Conn, message []byte) error{
	written := 0 
	for written < len(message) {
		n, err := conn.Write(message)
				if err != nil || n == 0 {
			log.Criticalf(
				"action: send_all | result: fail| total read: %v |  error: %v",
				err,
			)
			return err
		}
		written += n
	}
	return nil
}
// Serialize a string with unknown size, the first byte is the size of the string
func serializeUnknownString(message string, buf []byte) []byte{
	if len(message) > MAX_STR_SIZE {
		log.Criticalf( 
			"action: serialize_unknown_string | result: fail | error: string too long",
		)
		return nil
	}
	buf = append(buf, byte(len(message)))
	buf = append(buf, []byte(message)...)
	return buf
}


func SendBet(conn net.Conn, bet *Bet) error {
	msg, err := encodeBet(bet)
	if err != nil {return err}
	err = send_all(conn, msg)
	if err != nil {return err} else {return nil}
}

func encodeBet (bet *Bet) ([]byte, error) {
	// agency
	msg := make([]byte, 0)
	msg = append(msg, bet.agency)

	// firstName
	msg = serializeUnknownString(bet.firstName, msg)
	if msg == nil {return nil, errors.New("error serializing firstName")}

	// lastName
	msg = serializeUnknownString(bet.lastName, msg)
	if msg == nil {return nil, errors.New("error serializing lastName")}

	// document
	docBytes := make([]byte, DOCUMENT_SIZE)
	binary.BigEndian.PutUint32(docBytes, bet.document)
	msg = append(msg, docBytes...)

	// birthDate
	msg = append(msg, []byte(bet.birthDate)...) // SIZE 10

	// number
	numBytes := make([]byte, NUMBER_SIZE)
	binary.BigEndian.PutUint16(numBytes, bet.number)
	msg = append(msg, numBytes...)

	return msg, nil 
}
// Send a batch of bets to the server
// The first two bytes are the number of bets
// Then it sends the bets
func SendBatch(conn net.Conn, bets []*Bet) error {
	var batches [][]byte
	var currentBatch []byte

	// Add total number of bets at the beginning of the first batch
	currentBatch = make([]byte, BATCH_SIZE) // Initialize with 2 bytes
	binary.BigEndian.PutUint16(currentBatch, uint16(len(bets)))

	for _, bet := range bets {
		betMsg, err := encodeBet(bet)
		if err != nil {
			return err
		}

		// If adding betMsg exceeds the max batch size, flush the current batch
		if len(currentBatch)+len(betMsg) > MAX_BATCH_BYTES {
			batches = append(batches, currentBatch) // Save current batch
			currentBatch = make([]byte, 0)          // Start a new empty batch (no bet count in subsequent batches)
		}

		// Append the bet message to the current batch
		currentBatch = append(currentBatch, betMsg...)
	}

	// Add the last batch if there's any remaining data
	if len(currentBatch) > 0 {
		batches = append(batches, currentBatch)
	}

	// Send all batches
	for _, batch := range batches {
		if err := send_all(conn, batch); err != nil {
			return err
		}
	}

	return nil
}

func RecvAnswer(conn net.Conn) (int,error) {
	answer, err := RecvAll(conn, ANSWER_SIZE)
	if err != nil {return -1, err}
	answer_v := int(answer[0])			
	return answer_v, nil 
}