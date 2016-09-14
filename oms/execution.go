package oms

//Execution is the execution type
type Execution struct {
	ID       int    `json:"id"`
	Symbol   string `json:"symbol"`
	Quantity string `json:"quantity"`
	Side     string `json:"side"`
	Price    string `json:"price"`
	Session  string `json:"session_id"`
}
