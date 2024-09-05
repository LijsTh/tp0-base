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
const MAX_BATCH_BYTES = 8000 // 8KB
const BATCH_SIZE = 2 
const AGENCY_SIZE = 1
const WINNERS_SIZE = 1

const SUCESS = 0
const FAIL = 1
const END_BATCH = 0
const FINISH = 3


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

	return msg, nil 
}

func SendBatch(conn net.Conn, bets []*Bet) error {
	msg := make([]byte, BATCH_SIZE) // 2 bytes
	// sends the number of bets
	binary.BigEndian.PutUint16(msg, uint16(len(bets)))
	for _, bet := range bets {
		if len(msg) > MAX_BATCH_BYTES {
			log.Critical("action: send_batch | result: fail | error: batch too big")
			os.Exit(1)
		}
		betMsg, err := encodeBet(bet)
		if err != nil {return err}
		msg = append(msg, betMsg...)
	}
	err := send_all(conn, msg)
	if err != nil {return err} else {return nil}
}

func RecvAnswer(conn net.Conn) (int,error) {
	answer, err := RecvAll(conn, ANSWER_SIZE)
	if err != nil {return -1, err}
	answer_v := int(answer[0])			
	return answer_v, nil 
}

func RecvResults(conn net.Conn) ([]uint32, error) {
	winners_bytes, err := RecvAll(conn, WINNERS_SIZE)
	if err != nil {return nil, err}
	winners_n := int(winners_bytes[0])
	winners := make([]uint32, winners_n)
	for i := 0; i < winners_n; i++ {
		winner, err := RecvAll(conn, DOCUMENT_SIZE)
		if err != nil {panic(err)}
		winners[i] = binary.BigEndian.Uint32(winner)
	}
	return winners, nil

}

func sendEndMessage(conn net.Conn, agency int) error {
	// The end message needs to have the same lenght as BATCH size as the server will read it as a batch with 0 bets
	msg := make([]byte, BATCH_SIZE + AGENCY_SIZE)
	binary.BigEndian.PutUint16(msg, END_BATCH)
	msg[2] = byte(agency)
	err := send_all(conn, msg) 
	if err != nil {return err} else {return nil}

}


func sendFinish(conn net.Conn) error {
	msg := make([]byte, ANSWER_SIZE) 
	msg[0] = FINISH 
	err := send_all(conn, msg)
	if err != nil {return err} else {return nil}
}