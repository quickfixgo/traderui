package main

import (
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/quickfixgo/quickfix"
	"github.com/quickfixgo/quickfix/enum"
	"github.com/quickfixgo/quickfix/field"

	fix40nos "github.com/quickfixgo/quickfix/fix40/newordersingle"
	fix41nos "github.com/quickfixgo/quickfix/fix41/newordersingle"
	fix42nos "github.com/quickfixgo/quickfix/fix42/newordersingle"
	fix43nos "github.com/quickfixgo/quickfix/fix43/newordersingle"
	fix44nos "github.com/quickfixgo/quickfix/fix44/newordersingle"
	fix50nos "github.com/quickfixgo/quickfix/fix50/newordersingle"

	"github.com/shopspring/decimal"
)

var clOrdIDLock sync.Mutex
var clOrdID = 0

var app = newTradeClient()

func nextClOrdID() string {
	clOrdIDLock.Lock()
	defer clOrdIDLock.Unlock()

	clOrdID++
	return strconv.Itoa(clOrdID)
}

type order struct {
	SessionID quickfix.SessionID
	ClOrdID   string
	Symbol    string
	Quantity  decimal.Decimal
	Account   string
	Session   string
	Side      string
	OrdType   string
	Price     decimal.Decimal
	StopPrice decimal.Decimal
	Closed    string
	Open      decimal.Decimal
	AvgPx     string
}

type tradeClient struct {
	SessionIDs map[string]quickfix.SessionID
	Orders     map[string]*order
	OrderLock  sync.Mutex
}

func newTradeClient() *tradeClient {
	tc := &tradeClient{
		SessionIDs: make(map[string]quickfix.SessionID),
		Orders:     make(map[string]*order),
	}

	return tc
}

func (e *tradeClient) OnLogon(sessionID quickfix.SessionID)                       {}
func (e *tradeClient) OnLogout(sessionID quickfix.SessionID)                      {}
func (e *tradeClient) ToAdmin(msg quickfix.Message, sessionID quickfix.SessionID) {}
func (e *tradeClient) OnCreate(sessionID quickfix.SessionID) {
	if e.SessionIDs == nil {
		e.SessionIDs = make(map[string]quickfix.SessionID)
	}
	e.SessionIDs[sessionID.String()] = sessionID
}

func (e *tradeClient) FromAdmin(msg quickfix.Message, sessionID quickfix.SessionID) (reject quickfix.MessageRejectError) {
	return
}

func (e *tradeClient) ToApp(msg quickfix.Message, sessionID quickfix.SessionID) (err error) {
	return
}

func (e *tradeClient) FromApp(msg quickfix.Message, sessionID quickfix.SessionID) quickfix.MessageRejectError {
	var msgType field.MsgTypeField
	if err := msg.Header.Get(&msgType); err != nil {
		return err
	}

	switch msgType.String() {
	case enum.MsgType_EXECUTION_REPORT:
		return e.OnExecutionReport(msg, sessionID)
	}

	return quickfix.UnsupportedMessageType()
}

func (e *tradeClient) OnExecutionReport(msg quickfix.Message, sessionID quickfix.SessionID) quickfix.MessageRejectError {
	e.OrderLock.Lock()
	defer e.OrderLock.Unlock()

	var clOrdID field.ClOrdIDField
	if err := msg.Body.Get(&clOrdID); err != nil {
		return err
	}

	order, ok := e.Orders[clOrdID.String()]
	if !ok {
		log.Printf("[ERROR] could not find order with clordid %v", clOrdID.String())
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

	order.Closed = cumQty.String()
	order.Open = order.Quantity.Sub(cumQty.Decimal)
	order.AvgPx = avgPx.String()

	return nil
}

var templates = template.Must(template.New("traderui").Funcs(tmplFuncs).ParseFiles("tmpl/index.html", "tmpl/orders.html"))

func traderView(w http.ResponseWriter, r *http.Request) {
	if err := templates.ExecuteTemplate(w, "index.html", app); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func orders(w http.ResponseWriter, r *http.Request) {
	if err := templates.ExecuteTemplate(w, "orders.html", app); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func newOrder(w http.ResponseWriter, r *http.Request) {
	symbol := r.FormValue("symbol")
	account := r.FormValue("account")
	ordType := r.FormValue("ordType")
	side := r.FormValue("side")

	sessionID, ok := app.SessionIDs[r.FormValue("session")]

	if !ok {
		http.Error(w, "Invalid SessionID", http.StatusBadRequest)
		return
	}

	qty, err := decimal.NewFromString(r.FormValue("quantity"))
	if err != nil {
		http.Error(w, "Invalid Qty", http.StatusBadRequest)
		return
	}

	ord := order{
		SessionID: sessionID,
		ClOrdID:   nextClOrdID(),
		Symbol:    symbol,
		Quantity:  qty,
		Open:      qty,
		Closed:    "0",
		Account:   account,
		Session:   sessionID.String(),
		Side:      side,
		OrdType:   ordType,
	}

	switch ordType {
	case enum.OrdType_LIMIT, enum.OrdType_STOP_LIMIT:
		price, err := decimal.NewFromString(r.FormValue("price"))

		if err != nil {
			http.Error(w, "Invalid Price", http.StatusBadRequest)
			return
		}

		ord.Price = price
	}

	switch ordType {
	case enum.OrdType_STOP, enum.OrdType_STOP_LIMIT:
		stopPrice, err := decimal.NewFromString(r.FormValue("stopPrice"))

		if err != nil {
			http.Error(w, "Invalid StopPrice", http.StatusBadRequest)
			return
		}

		ord.StopPrice = stopPrice
	}

	app.OrderLock.Lock()
	app.Orders[ord.ClOrdID] = &ord
	app.OrderLock.Unlock()

	switch sessionID.BeginString {
	case enum.BeginStringFIX40:
		err = send40NOS(ord)
	case enum.BeginStringFIX41:
		err = send41NOS(ord)
	case enum.BeginStringFIX42:
		err = send42NOS(ord)
	case enum.BeginStringFIX43:
		err = send43NOS(ord)
	case enum.BeginStringFIX44:
		err = send44NOS(ord)
	case enum.BeginStringFIXT11:
		err = send50NOS(ord)
	default:
		err = errors.New("Unhandled BeginString")
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func populateOrder(genMessage quickfix.Messagable, ord order) (msg quickfix.Message) {
	msg = genMessage.ToMessage()

	switch ord.OrdType {
	case enum.OrdType_LIMIT, enum.OrdType_STOP_LIMIT:
		msg.Body.Set(field.NewPrice(ord.Price, 2))
	}

	switch ord.OrdType {
	case enum.OrdType_STOP, enum.OrdType_STOP_LIMIT:
		msg.Body.Set(field.NewStopPx(ord.StopPrice, 2))
	}

	return
}

func send40NOS(ord order) error {
	nos := fix40nos.New(
		field.NewClOrdID(ord.ClOrdID),
		field.NewHandlInst("1"),
		field.NewSymbol(ord.Symbol),
		field.NewSide(ord.Side),
		field.NewOrderQty(ord.Quantity, 0),
		field.NewOrdType(ord.OrdType),
	)

	return quickfix.SendToTarget(populateOrder(nos, ord), ord.SessionID)
}

func send41NOS(ord order) error {
	nos := fix41nos.New(
		field.NewClOrdID(ord.ClOrdID),
		field.NewHandlInst("1"),
		field.NewSymbol(ord.Symbol),
		field.NewSide(ord.Side),
		field.NewOrdType(ord.OrdType),
	)
	nos.Set(field.NewOrderQty(ord.Quantity, 0))

	return quickfix.SendToTarget(populateOrder(nos, ord), ord.SessionID)
}

func send42NOS(ord order) error {
	nos := fix42nos.New(
		field.NewClOrdID(ord.ClOrdID),
		field.NewHandlInst("1"),
		field.NewSymbol(ord.Symbol),
		field.NewSide(ord.Side),
		field.NewTransactTime(time.Now()),
		field.NewOrdType(ord.OrdType),
	)
	nos.Set(field.NewOrderQty(ord.Quantity, 0))

	return quickfix.SendToTarget(populateOrder(nos, ord), ord.SessionID)
}

func send43NOS(ord order) error {
	nos := fix43nos.New(
		field.NewClOrdID(ord.ClOrdID),
		field.NewHandlInst("1"),
		field.NewSide(ord.Side),
		field.NewTransactTime(time.Now()),
		field.NewOrdType(ord.OrdType),
	)
	nos.Set(field.NewSymbol(ord.Symbol))
	nos.Set(field.NewOrderQty(ord.Quantity, 0))

	return quickfix.SendToTarget(populateOrder(nos, ord), ord.SessionID)
}

func send44NOS(ord order) error {
	nos := fix44nos.New(
		field.NewClOrdID(ord.ClOrdID),
		field.NewSide(ord.Side),
		field.NewTransactTime(time.Now()),
		field.NewOrdType(ord.OrdType),
	)
	nos.Set(field.NewSymbol(ord.Symbol))
	nos.Set(field.NewHandlInst("1"))
	nos.Set(field.NewOrderQty(ord.Quantity, 0))

	return quickfix.SendToTarget(populateOrder(nos, ord), ord.SessionID)
}

func send50NOS(ord order) error {
	nos := fix50nos.New(
		field.NewClOrdID(ord.ClOrdID),
		field.NewSide(ord.Side),
		field.NewTransactTime(time.Now()),
		field.NewOrdType(ord.OrdType),
	)
	nos.Set(field.NewHandlInst("1"))
	nos.Set(field.NewSymbol(ord.Symbol))
	nos.Set(field.NewOrderQty(ord.Quantity, 0))

	return quickfix.SendToTarget(populateOrder(nos, ord), ord.SessionID)
}

func prettyOrdType(e string) string {
	switch e {
	case enum.OrdType_LIMIT:
		return "Limit"
	case enum.OrdType_MARKET:
		return "Market"
	case enum.OrdType_STOP:
		return "Stop"
	case enum.OrdType_STOP_LIMIT:
		return "Stop Limit"
	}

	return e
}

func prettySide(e string) string {
	switch e {
	case enum.Side_BUY:
		return "Buy"
	case enum.Side_SELL:
		return "Sell"
	case enum.Side_SELL_SHORT:
		return "Sell Short"
	case enum.Side_SELL_SHORT_EXEMPT:
		return "Sell Short Exempt"
	case enum.Side_CROSS:
		return "Cross"
	case enum.Side_CROSS_SHORT:
		return "Cross Short"
	case enum.Side_CROSS_SHORT_EXEMPT:
		return "Cross Short Exempt"
	}
	return e
}

var tmplFuncs = template.FuncMap{
	"prettySide":    prettySide,
	"prettyOrdType": prettyOrdType,
}

func main() {
	cfgFileName := path.Join("config", "tradeclient.cfg")
	if flag.NArg() > 0 {
		cfgFileName = flag.Arg(0)
	}

	cfg, err := os.Open(cfgFileName)
	if err != nil {
		fmt.Printf("Error opening %v, %v\n", cfgFileName, err)
		return
	}

	appSettings, err := quickfix.ParseSettings(cfg)
	if err != nil {
		fmt.Println("Error reading cfg,", err)
		return
	}

	logFactory := quickfix.NewScreenLogFactory()

	initiator, err := quickfix.NewInitiator(app, quickfix.NewMemoryStoreFactory(), appSettings, logFactory)
	if err != nil {
		log.Fatalf("Unable to create Initiator: %s\n", err)
	}

	if err = initiator.Start(); err != nil {
		log.Fatal(err)
	}
	defer initiator.Stop()

	http.HandleFunc("/", traderView)
	http.HandleFunc("/order", newOrder)
	http.HandleFunc("/orders", orders)
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
