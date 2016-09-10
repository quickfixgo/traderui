package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"sync"
	"text/template"

	"github.com/gorilla/mux"
	"github.com/quickfixgo/quickfix"
	"github.com/quickfixgo/quickfix/enum"
	"github.com/quickfixgo/quickfix/field"
	"github.com/quickfixgo/traderui/internal"
	"github.com/shopspring/decimal"
)

type fixFactory interface {
	NewOrderSingle(ord internal.Order) (msg quickfix.Messagable, err error)
	OrderCancelRequest(ord internal.Order, clOrdID string) (msg quickfix.Messagable, err error)
}

type clOrdIDFactory interface {
	NextClOrdID() string
}

var clOrdIDLock sync.Mutex
var clOrdID = 0
var factory internal.BasicFIXFactory
var idFactory = new(internal.BasicClOrdIDFactory)

var app = newTradeClient(factory, idFactory)

func nextOrderID() int {
	clOrdIDLock.Lock()
	defer clOrdIDLock.Unlock()

	clOrdID++
	return clOrdID
}

type tradeClient struct {
	SessionIDs map[string]quickfix.SessionID
	Orders     map[string]*internal.Order
	OrderLock  sync.Mutex
	fixFactory
	clOrdIDFactory
}

func newTradeClient(factory fixFactory, idFactory clOrdIDFactory) *tradeClient {
	tc := &tradeClient{
		SessionIDs:     make(map[string]quickfix.SessionID),
		Orders:         make(map[string]*internal.Order),
		fixFactory:     factory,
		clOrdIDFactory: idFactory,
	}

	return tc
}

func (e *tradeClient) SessionsAsJSON() (s string, err error) {
	sessionIDs := make([]string, 0, len(e.SessionIDs))

	for s := range e.SessionIDs {
		sessionIDs = append(sessionIDs, s)
	}

	var b []byte
	b, err = json.Marshal(sessionIDs)
	s = string(b)
	return
}

func (e *tradeClient) OrdersAsJSON() (string, error) {
	e.OrderLock.Lock()
	defer e.OrderLock.Unlock()

	var orders = make([]*internal.Order, 0, len(e.Orders))
	for _, v := range e.Orders {
		orders = append(orders, v)
	}

	b, err := json.Marshal(orders)
	return string(b), err
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

	var leavesQty field.LeavesQtyField
	if err := msg.Body.Get(&leavesQty); err != nil {
		return err
	}

	order.Closed = cumQty.String()
	order.Open = leavesQty.String()
	order.AvgPx = avgPx.String()

	return nil
}

func traderView(w http.ResponseWriter, r *http.Request) {
	var templates = template.Must(template.New("traderui").ParseFiles("tmpl/index.html"))
	if err := templates.ExecuteTemplate(w, "index.html", app); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	app.OrderLock.Lock()
	defer app.OrderLock.Unlock()

	order, ok := app.Orders[id]
	if !ok {
		http.Error(w, "Unknown Order", http.StatusNotFound)
		return
	}

	outgoingJSON, err := json.Marshal(order)
	if err != nil {
		log.Printf("[ERROR] err = %+v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(outgoingJSON))
}

func deleteOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	app.OrderLock.Lock()
	defer app.OrderLock.Unlock()

	order, ok := app.Orders[id]
	if !ok {
		http.Error(w, "Unknown Order", http.StatusNotFound)
		return
	}

	clOrdID := app.NextClOrdID()
	app.Orders[clOrdID] = order

	msg, err := app.OrderCancelRequest(*order, clOrdID)
	if err != nil {
		log.Printf("[ERROR] err = %+v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = quickfix.SendToTarget(msg, order.SessionID)
	if err != nil {
		log.Printf("[ERROR] err = %+v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getOrders(w http.ResponseWriter, r *http.Request) {
	app.OrderLock.Lock()
	defer app.OrderLock.Unlock()

	var orders = make([]*internal.Order, 0, len(app.Orders))
	for _, v := range app.Orders {
		orders = append(orders, v)
	}

	outgoingJSON, err := json.Marshal(orders)
	if err != nil {
		log.Printf("[ERROR] err = %+v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(outgoingJSON))
}

func newOrder(w http.ResponseWriter, r *http.Request) {
	var order internal.Order
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&order)

	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	err = initOrder(&order)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	app.OrderLock.Lock()
	app.Orders[order.ClOrdID] = &order
	app.OrderLock.Unlock()

	msg, err := app.NewOrderSingle(order)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = quickfix.SendToTarget(msg, order.SessionID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func initOrder(order *internal.Order) error {
	if sessionID, ok := app.SessionIDs[order.Session]; ok {
		order.SessionID = sessionID
	} else {
		return fmt.Errorf("Invalid SessionID")
	}

	if qty, err := decimal.NewFromString(order.Quantity); err == nil {
		order.QuantityDecimal = qty
	} else {
		return fmt.Errorf("Invalid Qty")
	}

	switch order.OrdType {
	case enum.OrdType_LIMIT, enum.OrdType_STOP_LIMIT:
		if price, err := decimal.NewFromString(order.Price); err == nil {
			order.PriceDecimal = price
		} else {
			return fmt.Errorf("Invalid Price")
		}
	}

	switch order.OrdType {
	case enum.OrdType_STOP, enum.OrdType_STOP_LIMIT:
		if stopPrice, err := decimal.NewFromString(order.StopPrice); err == nil {
			order.StopPriceDecimal = stopPrice
		} else {
			return fmt.Errorf("Invalid StopPrice")
		}
	}

	order.ID = nextOrderID()
	order.ClOrdID = app.NextClOrdID()

	return nil
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

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/orders", newOrder).Methods("POST")
	router.HandleFunc("/orders", getOrders).Methods("GET")
	router.HandleFunc("/orders/{id:[0-9]+}", getOrder).Methods("GET")
	router.HandleFunc("/orders/{id:[0-9]+}", deleteOrder).Methods("DELETE")
	router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	router.HandleFunc("/", traderView)

	log.Fatal(http.ListenAndServe(":8080", router))
}
