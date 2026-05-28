package response

import (
	"encoding/json"
	"net/http"
)

type Envelope struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

func Success(data any) Envelope {
	return Envelope{
		Code: 0,
		Msg:  "success",
		Data: data,
	}
}

func Error(code int, msg string) Envelope {
	return Envelope{
		Code: code,
		Msg:  msg,
	}
}

func WriteJSON(w http.ResponseWriter, status int, payload Envelope) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
