package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"text/template"

	"github.com/gorilla/mux"
	"github.com/quickfixgo/quickfix"
	"github.com/quickfixgo/quickfix/enum"
	"github.com/quickfixgo/traderui/basic"
	"github.com/quickfixgo/traderui/oms"
	"github.com/shopspring/decimal"
)

type fixFactory interface {
	NewOrderSingle(ord oms.Order) (msg quickfix.Messagable, err error)
	OrderCancelRequest(ord oms.Order, clOrdID string) (msg quickfix.Messagable, err error)
}

type tradeClient struct {
	SessionIDs map[string]quickfix.SessionID
	fixFactory
	*oms.OrderManager
}

func newTradeClient(factory fixFactory, idGen oms.ClOrdIDGenerator) *tradeClient {
	tc := &tradeClient{
		SessionIDs:   make(map[string]quickfix.SessionID),
		fixFactory:   factory,
		OrderManager: oms.NewOrderManager(idGen),
	}

	return tc
}

func (c tradeClient) SessionsAsJSON() (string, error) {
	sessionIDs := make([]string, 0, len(c.SessionIDs))

	for s := range c.SessionIDs {
		sessionIDs = append(sessionIDs, s)
	}

	b, err := json.Marshal(sessionIDs)
	return string(b), err
}

func (c tradeClient) OrdersAsJSON() (string, error) {
	c.RLock()
	defer c.RUnlock()

	b, err := json.Marshal(c.GetAll())
	return string(b), err
}

func (c tradeClient) traderView(w http.ResponseWriter, r *http.Request) {
	var templates = template.Must(template.New("traderui").ParseFiles("tmpl/index.html"))
	if err := templates.ExecuteTemplate(w, "index.html", c); err != nil {
		log.Printf("[ERROR] err = %+v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c tradeClient) fetchRequestedOrder(r *http.Request) (*oms.Order, error) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		panic(err)
	}

	return c.Get(id)
}

func (c tradeClient) getOrder(w http.ResponseWriter, r *http.Request) {
	c.RLock()
	defer c.RUnlock()

	order, err := c.fetchRequestedOrder(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
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

func (c tradeClient) deleteOrder(w http.ResponseWriter, r *http.Request) {
	c.Lock()
	defer c.Unlock()

	order, err := c.fetchRequestedOrder(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	clOrdID := c.AssignNextClOrdID(order)
	msg, err := c.OrderCancelRequest(*order, clOrdID)
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

func (c tradeClient) getOrders(w http.ResponseWriter, r *http.Request) {
	outgoingJSON, err := c.OrdersAsJSON()
	if err != nil {
		log.Printf("[ERROR] err = %+v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, outgoingJSON)
}

func (c tradeClient) newOrder(w http.ResponseWriter, r *http.Request) {
	var order oms.Order
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&order)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	err = c.initOrder(&order)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c.Lock()
	c.OrderManager.Save(&order)
	c.Unlock()

	msg, err := c.NewOrderSingle(order)
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

func (c tradeClient) initOrder(order *oms.Order) error {
	if sessionID, ok := c.SessionIDs[order.Session]; ok {
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

	var app = newTradeClient(basic.FIXFactory{}, new(basic.ClOrdIDGenerator))

	fixApp := &basic.FIXApplication{
		SessionIDs:   app.SessionIDs,
		OrderManager: app.OrderManager,
	}

	initiator, err := quickfix.NewInitiator(fixApp, quickfix.NewMemoryStoreFactory(), appSettings, logFactory)
	if err != nil {
		log.Fatalf("Unable to create Initiator: %s\n", err)
	}

	if err = initiator.Start(); err != nil {
		log.Fatal(err)
	}
	defer initiator.Stop()

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/orders", app.newOrder).Methods("POST")
	router.HandleFunc("/orders", app.getOrders).Methods("GET")
	router.HandleFunc("/orders/{id:[0-9]+}", app.getOrder).Methods("GET")
	router.HandleFunc("/orders/{id:[0-9]+}", app.deleteOrder).Methods("DELETE")
	router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	router.HandleFunc("/", app.traderView)

	log.Fatal(http.ListenAndServe(":8080", router))
}
