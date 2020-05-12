package mif

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMif_Simple_Ok(t *testing.T) {
	// init
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	l := &testLogger{}
	m := New(WithLogger(l))

	handler := m.Simple(func(w http.ResponseWriter, r *http.Request) error {
		w.WriteHeader(http.StatusNotFound)
		_, err := w.Write([]byte("test simple"))
		return err
	})

	// test
	handler.ServeHTTP(w, r)

	// check
	equals(t, http.StatusNotFound, w.Code)
	equals(t, "test simple", w.Body.String())
	equals(t, "", l.msg)
}

func TestMif_Simple_Error(t *testing.T) {
	// init
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	l := &testLogger{}
	m := New(WithLogger(l))

	handler := m.Simple(func(w http.ResponseWriter, r *http.Request) error {
		return errors.New("test error")
	})

	// test
	handler.ServeHTTP(w, r)

	// check
	equals(t, http.StatusInternalServerError, w.Code)
	equals(t, "server error", w.Body.String())
	equals(t, "test error", l.msg)
}

func TestMif_Wrap(t *testing.T) {
	type args struct {
		data *RawData
		err  error
	}

	type want struct {
		header http.Header
		code   int
		body   []byte
		log    string
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "ok simple",
			args: args{
				data: &RawData{
					Raw: []byte("test raw"),
				},
				err: nil,
			},
			want: want{
				header: http.Header{},
				code:   http.StatusOK,
				body:   []byte("test raw"),
				log:    "",
			},
		},
		{
			name: "ok status BadRequest",
			args: args{
				data: &RawData{
					Raw:    []byte("test raw"),
					Status: http.StatusBadRequest,
				},
				err: nil,
			},
			want: want{
				header: http.Header{},
				code:   http.StatusBadRequest,
				body:   []byte("test raw"),
				log:    "",
			},
		},
		{
			name: "ok custom headers",
			args: args{
				data: &RawData{
					Header: http.Header{"Hello": {"World"}},
					Raw:    []byte("test raw"),
				},
				err: nil,
			},
			want: want{
				header: http.Header{"Hello": {"World"}},
				code:   http.StatusOK,
				body:   []byte("test raw"),
				log:    "",
			},
		},
		{
			name: "ok nil raw data",
			args: args{
				data: &RawData{
					Raw: nil,
				},
				err: nil,
			},
			want: want{
				header: http.Header{},
				code:   http.StatusOK,
				body:   nil,
				log:    "",
			},
		},
		{
			name: "error nil data",
			args: args{
				data: nil,
				err:  nil,
			},
			want: want{
				header: http.Header{},
				code:   http.StatusInternalServerError,
				body:   []byte("server error"),
				log:    "data is nil",
			},
		},
		{
			name: "error custom",
			args: args{
				data: nil,
				err:  errors.New("test error"),
			},
			want: want{
				header: http.Header{},
				code:   http.StatusInternalServerError,
				body:   []byte("server error"),
				log:    "test error",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// init
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)

			l := &testLogger{}
			m := New(WithLogger(l))

			handler := m.Wrap(func(r *http.Request) (*RawData, error) {
				return tt.args.data, tt.args.err
			})

			// test
			handler.ServeHTTP(w, r)

			// check
			equals(t, tt.want.header, w.Header())
			equals(t, tt.want.code, w.Code)
			equals(t, tt.want.body, w.Body.Bytes())
			equals(t, tt.want.log, l.msg)
		})
	}
}

func TestMif_WrapJSON(t *testing.T) {
	type jsonTest struct {
		Text string `json:"text"`
	}

	type args struct {
		data *JSONData
		err  error
	}

	type want struct {
		header http.Header
		code   int
		body   []byte
		log    string
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "ok simple",
			args: args{
				data: &JSONData{
					Data: jsonTest{Text: "test"},
				},
				err: nil,
			},
			want: want{
				header: http.Header{"Content-Type": []string{"application/json"}},
				code:   http.StatusOK,
				body:   []byte(`{"text":"test"}`),
				log:    "",
			},
		},
		{
			name: "ok status BadRequest",
			args: args{
				data: &JSONData{
					Status: http.StatusBadRequest,
					Data:   jsonTest{Text: "test"},
				},
				err: nil,
			},
			want: want{
				header: http.Header{"Content-Type": []string{"application/json"}},
				code:   http.StatusBadRequest,
				body:   []byte(`{"text":"test"}`),
				log:    "",
			},
		},
		{
			name: "ok custom headers",
			args: args{
				data: &JSONData{
					Header: http.Header{
						"Hello": {"World"},
					},
					Data: jsonTest{Text: "test"},
				},
				err: nil,
			},
			want: want{
				header: http.Header{
					"Hello":        {"World"},
					"Content-Type": []string{"application/json"},
				},
				code: http.StatusOK,
				body: []byte(`{"text":"test"}`),
				log:  "",
			},
		},
		{
			name: "ok nil json data",
			args: args{
				data: &JSONData{},
				err:  nil,
			},
			want: want{
				header: http.Header{"Content-Type": []string{"application/json"}},
				code:   http.StatusOK,
				body:   []byte("null"),
				log:    "",
			},
		},
		{
			name: "ok not json data string",
			args: args{
				data: &JSONData{
					NotJSON: true,
					Data:    "test non json",
				},
				err: nil,
			},
			want: want{
				header: http.Header{},
				code:   http.StatusOK,
				body:   []byte("test non json"),
				log:    "",
			},
		},
		{
			name: "ok not json data bytes",
			args: args{
				data: &JSONData{
					NotJSON: true,
					Data:    []byte("test non json"),
				},
				err: nil,
			},
			want: want{
				header: http.Header{},
				code:   http.StatusOK,
				body:   []byte("test non json"),
				log:    "",
			},
		},
		{
			name: "error not json data",
			args: args{
				data: &JSONData{
					NotJSON: true,
					Data:    1234,
				},
				err: nil,
			},
			want: want{
				header: http.Header{"Content-Type": []string{"application/json"}},
				code:   http.StatusInternalServerError,
				body:   []byte(`{"error_message":"server error"}`),
				log:    `bad "not json" data, expected "[]byte" or "string"`,
			},
		},
		{
			name: "error nil data",
			args: args{
				data: nil,
				err:  nil,
			},
			want: want{
				header: http.Header{"Content-Type": []string{"application/json"}},
				code:   http.StatusInternalServerError,
				body:   []byte(`{"error_message":"server error"}`),
				log:    "data is nil",
			},
		},
		{
			name: "error custom",
			args: args{
				data: nil,
				err:  errors.New("test error"),
			},
			want: want{
				header: http.Header{"Content-Type": []string{"application/json"}},
				code:   http.StatusInternalServerError,
				body:   []byte(`{"error_message":"server error"}`),
				log:    "test error",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// init
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)

			l := &testLogger{}
			m := New(WithLogger(l), SetDisablePanic())

			handler := m.WrapJSON(func(r *http.Request) (*JSONData, error) {
				return tt.args.data, tt.args.err
			})

			// test
			handler.ServeHTTP(w, r)

			// check
			equals(t, tt.want.header, w.Header())
			equals(t, tt.want.code, w.Code)
			equals(t, tt.want.body, w.Body.Bytes())
			equals(t, tt.want.log, l.msg)
		})
	}
}

func TestMif_WrapJSON_JSONConfig(t *testing.T) {
	type jsonTest struct {
		Text string `json:"text"`
	}

	// init
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	l := &testLogger{}
	m := New(WithLogger(l), WithJSONConfig(JSONConfig{Prefix: "<prefix>", Indent: "<indent>"}))

	handler := m.WrapJSON(func(r *http.Request) (*JSONData, error) {
		return &JSONData{
			Data: jsonTest{Text: "test"},
		}, nil
	})

	// test
	handler.ServeHTTP(w, r)

	// check
	equals(t, http.Header{"Content-Type": []string{"application/json"}}, w.Header())
	equals(t, http.StatusOK, w.Code)
	equals(t, []byte(`{
<prefix><indent>"text": "test"
<prefix>}`), w.Body.Bytes())
	equals(t, "", l.msg)
}
