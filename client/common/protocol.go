package common

import (
	"os"
	"net"
	"encoding/binary"

)

const MAX_STR_SIZE = 255
const DOCUMENT_SIZE = 4
const NUMBER_SIZE = 2
const ANSWER_SIZE = 1

const SUCESS = 0
const FAIL = 1


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

func serializeUnknownString(message string, buf []byte) []byte{
	if len(message) > MAX_STR_SIZE {
		log.Criticalf( 
			"action: serialize_unknown_string | result: fail | error: string too long",
		)
		os.Exit(1)
	}
	buf = append(buf, byte(len(message)))
	buf = append(buf, []byte(message)...)
	return buf
}


func SendBet(conn net.Conn, bet *Bet) error {
	// agency
	msg := make([]byte, 0)
	msg = append(msg, bet.agency)

	// firstName
	msg = serializeUnknownString(bet.firstName, msg)

	// lastName
	msg = serializeUnknownString(bet.lastName, msg)

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

	err := send_all(conn, msg)
	if err != nil {return err} else {return nil}
}

func RecvAnswer(conn net.Conn) error {
	answer, err := RecvAll(conn, ANSWER_SIZE)
	if answer[0] != SUCESS {
		log.Error("action: receive_message | result: fail")
	} else {
		log.Info("action: receive_message | result: success") 
	}
	return err
}