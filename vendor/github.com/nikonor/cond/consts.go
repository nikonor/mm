package cond

import (
	"errors"
	"regexp"
)

var (
	ErrParseCond      = errors.New("condition parsing error")
	ErrWrongCommand   = errors.New("wrong command")
	ErrIncompleteData = errors.New("incomplete data")
	rSpaces           *regexp.Regexp
)

func init() {
	rSpaces = regexp.MustCompile(`\s+`)
}

const (
	CmdEQ = iota + 1
	CmdNE
	CmdGT
	CmdLT
	CmdGTE
	CmdLTE
	CmdAND
	CmdOR
	CmdNot
	CmdEQI
	CmdCONTAIN
	CmdICONTAIN

	Open  = '('
	Close = ')'
	Quote = '"'
	T     = "TRUE"
	F     = "FALSE"

	EQ       = "eq"
	NE       = "ne"
	GT       = "gt"
	LT       = "lt"
	GTE      = "gte"
	LTE      = "lte"
	AND      = "and"
	OR       = "or"
	NOT      = "not"
	EQI      = "eqi"
	CONTAIN  = "contain"
	ICONTAIN = "icontain"
)
