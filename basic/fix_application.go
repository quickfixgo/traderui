package basic

import (
	"log"

	"github.com/quickfixgo/quickfix"
	"github.com/quickfixgo/quickfix/enum"
	"github.com/quickfixgo/quickfix/field"
	"github.com/quickfixgo/quickfix/tag"
	"github.com/quickfixgo/traderui/oms"
)

//FIXApplication implements a basic quickfix.Application
type FIXApplication struct {
	SessionIDs map[string]quickfix.SessionID
	*oms.OrderManager
}

//OnLogon is ignored
func (a *FIXApplication) OnLogon(sessionID quickfix.SessionID) {}

//OnLogout is ignored
func (a *FIXApplication) OnLogout(sessionID quickfix.SessionID) {}

//ToAdmin is ignored
func (a *FIXApplication) ToAdmin(msg quickfix.Message, sessionID quickfix.SessionID) {}

//OnCreate initialized SessionIDs
func (a *FIXApplication) OnCreate(sessionID quickfix.SessionID) {
	a.SessionIDs[sessionID.String()] = sessionID
}

//FromAdmin is ignored
func (a *FIXApplication) FromAdmin(msg quickfix.Message, sessionID quickfix.SessionID) (reject quickfix.MessageRejectError) {
	return
}

//ToApp is ignored
func (a *FIXApplication) ToApp(msg quickfix.Message, sessionID quickfix.SessionID) (err error) {
	return
}

//FromApp listens for just execution reports
func (a *FIXApplication) FromApp(msg quickfix.Message, sessionID quickfix.SessionID) quickfix.MessageRejectError {
	msgType, err := msg.MsgType()
	if err != nil {
		return err
	}

	switch msgType {
	case enum.MsgType_EXECUTION_REPORT:
		return a.onExecutionReport(msg, sessionID)
	}

	return quickfix.UnsupportedMessageType()
}

func (a *FIXApplication) onExecutionReport(msg quickfix.Message, sessionID quickfix.SessionID) quickfix.MessageRejectError {
	a.Lock()
	defer a.Unlock()

	var clOrdID field.ClOrdIDField
	if err := msg.Body.Get(&clOrdID); err != nil {
		return err
	}

	order, err := a.GetByClOrdID(clOrdID.String())
	if err != nil {
		log.Printf("[ERROR] err= %v", err)
		return nil
	}

	var cumQty field.CumQtyField
	if err := msg.Body.Get(&cumQty); err != nil {
		return err
	}

	var avgPx field.AvgPxField
	if err := msg.Body.Get(&avgPx); err != nil {
		return err
	}

	var leavesQty field.LeavesQtyField
	if err := msg.Body.Get(&leavesQty); err != nil {
		return err
	}

	order.Closed = cumQty.String()
	order.Open = leavesQty.String()
	order.AvgPx = avgPx.String()

	if msg.Body.Has(tag.LastShares) {
		var lastShares field.LastSharesField
		if err := msg.Body.Get(&lastShares); err != nil {
			return err
		}

		var price field.LastPxField
		if err := msg.Body.Get(&price); err != nil {
			return err
		}

		exec := new(oms.Execution)
		exec.Symbol = order.Symbol
		exec.Side = order.Side
		exec.Session = order.Session

		exec.Quantity = lastShares.String()
		exec.Price = price.String()
		a.SaveExecution(exec)
	}

	return nil
}
