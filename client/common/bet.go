package common

import (
	"strconv"
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


func NewBetFromEnv() *Bet {
	agency := os.Getenv("AGENCY")
	firstName := os.Getenv("FIRSTNAME")
	lastName := os.Getenv("LASTNAME")
	document, err := strconv.Atoi(os.Getenv("DOCUMENT"))
	if err != nil {
		panic(err)
	}
	birthDate := os.Getenv("BIRTHDATE")
	number, err:= strconv.Atoi(os.Getenv("NUMBER"))
	if err != nil {
		panic(err)
	}
	return NewBet(agency, firstName, lastName, uint32(document), birthDate, uint16(number))
}