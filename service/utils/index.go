package utils

import (
	"encoding/json"
	"net/http"

	"service/log"

	"github.com/samber/mo"
)

type WebRes[T any] struct {
	Payload *T     `json:"payload"`
	Error   string `json:"error"`
}

func WriteHeaders(header *http.Header, method string, raw bool) {
	header.Set("Access-Control-Allow-Origin", "*")
	header.Set("Access-Control-Allow-Methods", method)
	header.Set("Access-Control-Allow-Headers", "Content-Type")

	contentType := "application/json"
	if raw {
		contentType = "application/octet-stream"
	}

	header.Set("Content-Type", contentType)
}

func WriteWebRes[T any](w http.ResponseWriter, payload mo.Option[T], code int) mo.Result[T] {
	var out WebRes[T]

	if v, ok := payload.Get(); ok {
		out.Payload = &v
	} else {
		out.Error = "Failed to encode response"

		w.WriteHeader(code)
		if err := json.NewEncoder(w).Encode(out); err != nil {
			log.Error("Failed to encode response: %s", err.Error())
			http.Error(w, "Failed to encode response", code)

			return mo.Err[T](err)
		}

		return mo.Errf[T]("%s", out.Error)
	}

	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(out); err != nil {
		log.Error("Failed to encode response: %s", err.Error())
		http.Error(w, "Failed to encode response", code)

		return mo.Err[T](err)
	}

	log.Info("%v: %v", code, *out.Payload)
	return mo.Ok(*out.Payload)
}

func WriteWebErr(w http.ResponseWriter, message string, code int) mo.Result[bool] {
	w.Header().Set("Content-Type", "application/json")

	var out WebRes[bool]
	out.Payload = nil
	out.Error = message

	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(out); err != nil {
		log.Error("Failed to encode error response: %s", err.Error())
		http.Error(w, "Failed to encode error response", code)

		return mo.Err[bool](err)
	}

	log.Error("%v: %s", code, message)
	return mo.Ok(true)
}

func WriteWebErrMethod(w http.ResponseWriter) mo.Result[bool] {
	webRes := WriteWebErr(w, "Method not allowed", http.StatusMethodNotAllowed)
	if webRes.IsError() {
		return mo.Err[bool](webRes.Error())
	}
	return webRes
}
