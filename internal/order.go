package internal

import (
	"github.com/quickfixgo/quickfix"
	"github.com/shopspring/decimal"
)

type Order struct {
	ID               int                `json:"id"`
	SessionID        quickfix.SessionID `json:"-"`
	ClOrdID          string             `json:"clord_id"`
	Symbol           string             `json:"symbol"`
	QuantityDecimal  decimal.Decimal    `json:"-"`
	Quantity         string             `json:"quantity"`
	Account          string             `json:"account"`
	Session          string             `json:"session_id"`
	Side             string             `json:"side"`
	OrdType          string             `json:"ord_type"`
	PriceDecimal     decimal.Decimal    `json:"-"`
	Price            string             `json:"price"`
	StopPriceDecimal decimal.Decimal    `json:"-"`
	StopPrice        string             `json:"stop_price"`
	Closed           string             `json:"closed"`
	Open             string             `json:"open"`
	AvgPx            string             `json:"avg_px"`
}
