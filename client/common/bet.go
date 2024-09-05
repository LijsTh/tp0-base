package common

import (
	"strconv"
	"encoding/csv"
	"os"
)

type Bet struct {
	agency uint8 
	firstName string
	lastName string
	document uint32
	birthDate string
	number uint16
}


// NewBet Creates a new bet
func NewBet(agency string, firstName string, lastName string, document uint32, birthDate string, number uint16) *Bet {
	agency_n, _ := strconv.Atoi(agency)
	bet := new(Bet)
	bet.agency = uint8(agency_n)
	bet.firstName = firstName
	bet.lastName = lastName
	bet.document = document
	bet.birthDate = birthDate
	bet.number = number
	return bet
}


type BetReader struct {
	agency string
	finished bool
	reader *csv.Reader
	file *os.File
	batch_size int
}

func NewBetReader(csv_path string, batch_size int, agency string) (*BetReader, error) {
	file, err := os.Open(csv_path)
	if err != nil {
		return nil, err
	}
	reader := csv.NewReader(file)
	bet_reader := new(BetReader)
	bet_reader.reader = reader
	bet_reader.finished = false
	bet_reader.batch_size = batch_size
	bet_reader.file = file
	bet_reader.agency = agency
	
	return bet_reader, nil
}


func (br *BetReader) ReadBet() (*Bet, error) {
	if br.finished {
		return nil, nil
	}
	record, err := br.reader.Read()
	if err != nil {
		if err.Error() == "EOF" {
			br.finished = true
			return nil, nil
		} else {return nil, err}
	}

	firstName := record[0]
	lastName := record[1]
	document, err := strconv.Atoi(record[2])
	if err != nil { return nil, err}
	birthDate := record[3]
	number,err := strconv.Atoi(record[4])
	if err != nil { return nil, err}

	return NewBet(br.agency, firstName, lastName, uint32(document), birthDate, uint16(number)), nil
}

func (br *BetReader) Finished() bool {
	return br.finished
}

func (br *BetReader) ReadBets() ([]*Bet, error) {
	var bets []*Bet
	for i:= 0; i < br.batch_size; i++ {
		bet, err := br.ReadBet()
		if err != nil {
			return nil, err
		}
		if bet == nil {
			break
		}
		bets = append(bets, bet)
	}
	return bets, nil
}

func (br *BetReader) Close() {
	br.file.Close()
}
