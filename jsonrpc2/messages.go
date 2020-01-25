package jsonrpc2

import "encoding/json"

type (
	request struct {
		ProtocolVersion string      `json:"jsonrpc"`
		ID              int         `json:"id"`
		Method          string      `json:"method"`
		Params          interface{} `json:"params,omitempty"`
	}
	response struct {
		ProtocolVersion string          `json:"jsonrpc"`
		ID              *int             `json:"id,omitempty"`
		Result          json.RawMessage `json:"result,omitempty"`
		Error           *responseError  `json:"error,omitempty"`
	}
	responseError struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	notification struct {
		ProtocolVersion string      `json:"jsonrpc"`
		Method          string      `json:"method"`
		Params          interface{} `json:"params,omitempty"`
	}
)
