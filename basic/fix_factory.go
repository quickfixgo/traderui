package basic

import (
	"errors"
	"time"

	"github.com/quickfixgo/quickfix"
	"github.com/quickfixgo/quickfix/enum"
	"github.com/quickfixgo/quickfix/field"
	"github.com/quickfixgo/traderui/oms"
	"github.com/quickfixgo/traderui/secmaster"

	fix40nos "github.com/quickfixgo/quickfix/fix40/newordersingle"
	fix41nos "github.com/quickfixgo/quickfix/fix41/newordersingle"
	fix42nos "github.com/quickfixgo/quickfix/fix42/newordersingle"
	fix43nos "github.com/quickfixgo/quickfix/fix43/newordersingle"
	fix44nos "github.com/quickfixgo/quickfix/fix44/newordersingle"
	fix50nos "github.com/quickfixgo/quickfix/fix50/newordersingle"

	fix42cxl "github.com/quickfixgo/quickfix/fix42/ordercancelrequest"
)

//FIXFactory builds vanilla fix messages, implements traderui.fixFactory
type FIXFactory struct{}

func (FIXFactory) NewOrderSingle(order oms.Order) (msg quickfix.Messagable, err error) {
	switch order.SessionID.BeginString {
	case enum.BeginStringFIX40:
		msg, err = nos40(order)
	case enum.BeginStringFIX41:
		msg, err = nos41(order)
	case enum.BeginStringFIX42:
		msg, err = nos42(order)
	case enum.BeginStringFIX43:
		msg, err = nos43(order)
	case enum.BeginStringFIX44:
		msg, err = nos44(order)
	case enum.BeginStringFIXT11:
		msg, err = nos50(order)
	default:
		err = errors.New("Unhandled BeginString")
	}

	return
}

func (FIXFactory) OrderCancelRequest(order oms.Order, clOrdID string) (msg quickfix.Messagable, err error) {
	switch order.SessionID.BeginString {
	case enum.BeginStringFIX42:
		msg, err = cxl42(order, clOrdID)
	default:
		err = errors.New("Unhandled BeginString")
	}

	return
}

func (FIXFactory) SecurityDefinitionRequest(req secmaster.SecurityDefinitionRequest) (msg quickfix.Messagable, err error) {
	err = errors.New("Not Implemented")
	return
}

func populateOrder(genMessage quickfix.Messagable, ord oms.Order) (quickfix.Messagable, error) {
	msg := genMessage.ToMessage()

	switch ord.OrdType {
	case enum.OrdType_LIMIT, enum.OrdType_STOP_LIMIT:
		msg.Body.Set(field.NewPrice(ord.PriceDecimal, 2))
	}

	switch ord.OrdType {
	case enum.OrdType_STOP, enum.OrdType_STOP_LIMIT:
		msg.Body.Set(field.NewStopPx(ord.StopPriceDecimal, 2))
	}

	return msg, nil
}

func nos40(ord oms.Order) (quickfix.Messagable, error) {
	nos := fix40nos.New(
		field.NewClOrdID(ord.ClOrdID),
		field.NewHandlInst("1"),
		field.NewSymbol(ord.Symbol),
		field.NewSide(ord.Side),
		field.NewOrderQty(ord.QuantityDecimal, 0),
		field.NewOrdType(ord.OrdType),
	)

	return populateOrder(nos, ord)
}

func nos41(ord oms.Order) (quickfix.Messagable, error) {
	nos := fix41nos.New(
		field.NewClOrdID(ord.ClOrdID),
		field.NewHandlInst("1"),
		field.NewSymbol(ord.Symbol),
		field.NewSide(ord.Side),
		field.NewOrdType(ord.OrdType),
	)
	nos.Set(field.NewOrderQty(ord.QuantityDecimal, 0))

	return populateOrder(nos, ord)
}

func nos42(ord oms.Order) (quickfix.Messagable, error) {
	nos := fix42nos.New(
		field.NewClOrdID(ord.ClOrdID),
		field.NewHandlInst("1"),
		field.NewSymbol(ord.Symbol),
		field.NewSide(ord.Side),
		field.NewTransactTime(time.Now()),
		field.NewOrdType(ord.OrdType),
	)
	nos.Set(field.NewOrderQty(ord.QuantityDecimal, 0))

	return populateOrder(nos, ord)
}

func cxl42(ord oms.Order, clOrdID string) (quickfix.Messagable, error) {
	cxl := fix42cxl.New(
		field.NewOrigClOrdID(ord.ClOrdID),
		field.NewClOrdID(clOrdID),
		field.NewSymbol(ord.Symbol),
		field.NewSide(ord.Side),
		field.NewTransactTime(time.Now()),
	)

	return cxl, nil
}

func nos43(ord oms.Order) (quickfix.Messagable, error) {
	nos := fix43nos.New(
		field.NewClOrdID(ord.ClOrdID),
		field.NewHandlInst("1"),
		field.NewSide(ord.Side),
		field.NewTransactTime(time.Now()),
		field.NewOrdType(ord.OrdType),
	)
	nos.Set(field.NewSymbol(ord.Symbol))
	nos.Set(field.NewOrderQty(ord.QuantityDecimal, 0))

	return populateOrder(nos, ord)
}

func nos44(ord oms.Order) (quickfix.Messagable, error) {
	nos := fix44nos.New(
		field.NewClOrdID(ord.ClOrdID),
		field.NewSide(ord.Side),
		field.NewTransactTime(time.Now()),
		field.NewOrdType(ord.OrdType),
	)
	nos.Set(field.NewSymbol(ord.Symbol))
	nos.Set(field.NewHandlInst("1"))
	nos.Set(field.NewOrderQty(ord.QuantityDecimal, 0))

	return populateOrder(nos, ord)
}

func nos50(ord oms.Order) (quickfix.Messagable, error) {
	nos := fix50nos.New(
		field.NewClOrdID(ord.ClOrdID),
		field.NewSide(ord.Side),
		field.NewTransactTime(time.Now()),
		field.NewOrdType(ord.OrdType),
	)
	nos.Set(field.NewHandlInst("1"))
	nos.Set(field.NewSymbol(ord.Symbol))
	nos.Set(field.NewOrderQty(ord.QuantityDecimal, 0))

	return populateOrder(nos, ord)
}
