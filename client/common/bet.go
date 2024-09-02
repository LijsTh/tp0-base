package common

import (
	"strconv"
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


