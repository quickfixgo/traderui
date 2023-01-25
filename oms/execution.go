package oms

import "github.com/quickfixgo/enum"

// Execution is the execution type
type Execution struct {
	ID       int       `json:"id"`
	Symbol   string    `json:"symbol"`
	Quantity string    `json:"quantity"`
	Side     enum.Side `json:"side"`
	Price    string    `json:"price"`
	Session  string    `json:"session_id"`
}
