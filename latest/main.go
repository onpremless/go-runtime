package doless

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Input interface{ any }
type Request[T Input] struct {
	Method  string
	Path    string
	Header  http.Header
	Payload T
}

type Error struct {
	Reason  string  `json:"reason"`
	Details *string `json:"details,omitempty"`
}

type LambdaF[T Input] func(ctx context.Context, req *Request[T]) (int, interface{})

func Handler[T Input](lambda LambdaF[T]) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			rec := recover()

			if rec == nil {
				return
			}

			ret := &Error{
				Reason: "Failed to handle request",
			}

			if err, isError := rec.(error); isError {
				details := err.Error()
				ret.Details = &details
			}

			if sErr, isString := rec.(string); isString {
				ret.Details = &sErr
			}

			resp, _ := json.Marshal(ret)

			w.WriteHeader(500)
			fmt.Fprint(w, string(resp))
		}()

		respError := func(err error) {
			w.WriteHeader(500)
			resp, _ := json.Marshal(&Error{
				Reason: err.Error(),
			})
			fmt.Fprint(w, string(resp))
		}

		defer req.Body.Close()

		var payload T
		switch interface{}(lambda).(type) {
		case LambdaF[io.Reader]:
			payload = req.Body.(T)
		case LambdaF[string]:
			data, err := io.ReadAll(req.Body)
			req.Body.Close()
			if err != nil {
				respError(err)
				return
			}
			payload = interface{}(string(data)).(T)
		default:
			data, err := io.ReadAll(req.Body)
			req.Body.Close()
			if err != nil {
				respError(err)
				return
			}

			if err := json.Unmarshal(data, &payload); err != nil {
				respError(err)
				return
			}
		}

		status, resp := lambda(req.Context(), &Request[T]{
			Method:  req.Method,
			Path:    req.URL.Path,
			Header:  req.Header.Clone(),
			Payload: payload,
		})

		w.WriteHeader(status)

		if resp == nil {
			return
		}

		if respT, ok := resp.(io.ReadCloser); ok {
			io.Copy(w, respT)
			respT.Close()
		} else if respT, ok := resp.(string); ok {
			fmt.Fprint(w, respT)
		} else {
			bytes, err := json.Marshal(resp)
			if err != nil {
				respError(err)
				return
			}

			fmt.Fprint(w, string(bytes))
		}
	}
}

func Lambda[T Input](lambda LambdaF[T]) {
	http.HandleFunc("/", Handler(lambda))
	http.HandleFunc("/health", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(200)
	})
	http.ListenAndServe(":3000", nil)
}
