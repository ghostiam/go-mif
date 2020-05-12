package mif

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type JSONData struct {
	NotJSON bool
	Header  http.Header
	Status  int
	Data    interface{}
}

type RawData struct {
	Header http.Header
	Status int
	Raw    []byte
}

type SimpleFunc func(w http.ResponseWriter, r *http.Request) error
type WrapFunc func(r *http.Request) (*RawData, error)
type WrapJSONFunc func(r *http.Request) (*JSONData, error)

type MIF interface {
	Simple(next SimpleFunc) http.HandlerFunc
	Wrap(next WrapFunc) http.HandlerFunc
	WrapJSON(next WrapJSONFunc) http.HandlerFunc
}

type mif struct {
	logger       Logger
	json         JSONConfig
	disablePanic bool
}

func New(opts ...Option) MIF {
	m := &mif{
		logger: noopLogger{},
	}

	// apply the list of options to Server
	for _, opt := range opts {
		opt.apply(m)
	}

	return m
}

func (m *mif) Simple(next SimpleFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		err := next(w, r)
		if err != nil {
			m.logger.Error(err)

			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("server error"))
		}
	}

	return fn
}

func (m *mif) Wrap(next WrapFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		data, err := next(r)
		if err != nil {
			m.logger.Error(err)

			data = &RawData{
				Status: http.StatusInternalServerError,
				Raw:    []byte("server error"),
			}
		}

		if data == nil {
			m.logger.Error("data is nil")

			data = &RawData{
				Status: http.StatusInternalServerError,
				Raw:    []byte("server error"),
			}
		}

		for key, val := range data.Header {
			w.Header()[key] = val
		}

		if data.Status == 0 {
			data.Status = http.StatusOK
		}

		w.WriteHeader(data.Status)
		_, err = w.Write(data.Raw)
		if err != nil {
			m.logger.Error(fmt.Errorf("write response: %w", err))
			return
		}
	}

	return fn
}

type UserError struct {
	Message string `json:"error_message"`
}

func (m *mif) WrapJSON(next WrapJSONFunc) http.HandlerFunc {
	fn := func(r *http.Request) (*RawData, error) {
		data, err := next(r)
		if err != nil {
			m.logger.Error(err)

			data = &JSONData{
				Status: http.StatusInternalServerError,
				Data:   UserError{"server error"},
			}
		}

		if data == nil {
			m.logger.Error("data is nil")

			data = &JSONData{
				Status: http.StatusInternalServerError,
				Data:   UserError{"server error"},
			}
		}

		var resp []byte
		if data.NotJSON {
			switch v := data.Data.(type) {
			case []byte:
				resp = v
			case string:
				resp = []byte(v)
			default:
				msg := `bad "not json" data, expected "[]byte" or "string"`
				if !m.disablePanic {
					panic(msg)
				}

				m.logger.Error(msg)

				data = &JSONData{
					Status: http.StatusInternalServerError,
					Data:   UserError{"server error"},
				}
			}
		}

		if !data.NotJSON {
			if data.Header == nil {
				data.Header = make(http.Header)
			}

			if data.Header.Get("Content-Type") == "" {
				data.Header.Set("Content-Type", "application/json")
			}

			if m.json.Prefix != "" || m.json.Indent != "" {
				resp, err = json.MarshalIndent(data.Data, m.json.Prefix, m.json.Indent)
			} else {
				resp, err = json.Marshal(data.Data)
			}

			if err != nil {
				return nil, fmt.Errorf("json encode: %w", err)
			}
		}

		return &RawData{
			Header: data.Header,
			Status: data.Status,
			Raw:    resp,
		}, nil
	}

	return m.Wrap(fn)
}
