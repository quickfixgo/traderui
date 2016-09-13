package secmaster

import "github.com/quickfixgo/quickfix"

//SecurityDefinitionRequest is the SecurityDefinitionRequest type
type SecurityDefinitionRequest struct {
	ID                  int                `json:"id"`
	SessionID           quickfix.SessionID `json:"-"`
	Session             string             `json:"session_id"`
	SecurityRequestType int                `json:"security_request_type"`
	Symbol              string             `json:"symbol"`
	SecurityType        string             `json:"security_type"`
}
