package common

import (
	"os"
	"bufio"
	"net"
	"encoding/binary"

)

const MAX_STR_SIZE = 255
const DOCUMENT_SIZE = 4
const NUMBER_SIZE = 2

func RecvAll(conn net.Conn, size int) []byte {
	reader := bufio.NewReader(conn)
	msg := make([]byte, size)
	read := int(0)
	for read < size {
		n, err := reader.Read(msg)
		if err != nil || n == 0 {
			log.Criticalf(
				"action: recv_all | result: fail | total read: %v | error: %v",
				read,
				err,
			)
			os.Exit(1)
		}
		read += n
	} 
	
	return msg

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
