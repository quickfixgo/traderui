package secmaster

import (
	"github.com/quickfixgo/enum"
	"github.com/quickfixgo/quickfix"
)

//SecurityDefinitionRequest is the SecurityDefinitionRequest type
type SecurityDefinitionRequest struct {
	ID                  int                      `json:"id"`
	SessionID           quickfix.SessionID       `json:"-"`
	Session             string                   `json:"session_id"`
	SecurityRequestType enum.SecurityRequestType `json:"security_request_type"`
	Symbol              string                   `json:"symbol"`
	SecurityType        enum.SecurityType        `json:"security_type"`
}
