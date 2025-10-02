package httpapi

import (
    "encoding/json"
    "net/http"
)

type errorBody struct {
    Error struct {
        Code    string      `json:"code"`
        Message string      `json:"message"`
        Details interface{} `json:"details,omitempty"`
    } `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(status)
    enc := json.NewEncoder(w)
    _ = enc.Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, msg string, details any) {
    var eb errorBody
    eb.Error.Code = code
    eb.Error.Message = msg
    eb.Error.Details = details
    writeJSON(w, status, eb)
}

