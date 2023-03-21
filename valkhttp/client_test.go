package valkhttp

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

type testStruct struct {
	Some       string  `json:"some" xml:"some"`
	Count      float64 `json:"cnt" xml:"cnt"`
	Found      bool    `json:"found" xml:"found"`
	Validation *string `json:"validation" xml:"validation" validate:"required"`
}

func init() {
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 100
}

// A stub of fasthttp in order to run the `Do()` offline
type StubFasthttp struct {
	resp *fasthttp.Response
	err  error
}

// Fake Do-method that ignores input and returns canned response
func (s *StubFasthttp) DoTimeout(_ *fasthttp.Request, resp *fasthttp.Response, _ time.Duration) error {
	s.resp.CopyTo(resp)
	return s.err
}

func createStub(body []byte, statusCode int, err error) *Client {
	response := fasthttp.Response{}
	response.SetBodyRaw(body)
	response.SetStatusCode(statusCode)
	return &Client{
		fastClient: &StubFasthttp{
			err:  err,
			resp: &response,
		},
	}
}

func Test_get_httpCodes(t *testing.T) {
	testCases := []struct {
		parser       Parser
		desc         string
		responseBody string
		statusCode   int
		err          error
		wantErr      error
	}{
		{
			parser:       &JSONParser,
			desc:         "Http 401 to error",
			responseBody: "pelle",
			statusCode:   401,
			wantErr:      NewHTTPError(401, "pelle"),
		},
		{
			parser:       &JSONParser,
			desc:         "Http 200 no error",
			responseBody: "{}",
			statusCode:   200,
			wantErr:      nil,
		},
		{
			parser:       &JSONParser,
			desc:         "Http timeout",
			responseBody: "{}",
			statusCode:   0,
			err:          fasthttp.ErrTimeout,
			wantErr:      TimeoutError,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			c := createStub([]byte(tC.responseBody), tC.statusCode, tC.err)
			var resp struct{}
			req := &HTTPRequest{
				URL: "what/ever",
			}
			err := c.Get(context.Background(), tC.parser, req, &resp)
			assert.Equal(t, tC.wantErr, err)
		})
	}
}

func Test_post(t *testing.T) {
	testCases := []struct {
		parser       Parser
		desc         string
		responseBody string
		statusCode   int
		err          error
		wantErr      error
	}{
		{
			parser:       &JSONParser,
			desc:         "Post 200 with body",
			responseBody: "{}",
			statusCode:   200,
			wantErr:      nil,
		},
		{
			parser:       &JSONParser,
			desc:         "Post 201 no content",
			responseBody: "",
			statusCode:   201,
			wantErr:      nil,
		},
		{
			parser:       &JSONParser,
			desc:         "Post getting 500",
			responseBody: "total chaos",
			statusCode:   500,
			wantErr:      NewHTTPError(500, "total chaos"),
		},
		{
			parser:       &JSONParser,
			desc:         "Http timeout",
			responseBody: "{}",
			statusCode:   0,
			err:          fasthttp.ErrTimeout,
			wantErr:      TimeoutError,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			c := createStub([]byte(tC.responseBody), tC.statusCode, tC.err)
			var resp struct{}
			req := &HTTPRequest{
				URL: "dont/care",
			}
			err := c.Post(context.Background(), tC.parser, req, &resp)
			assert.Equal(t, tC.wantErr, err)
		})
	}
}

func Test_put(t *testing.T) {
	testCases := []struct {
		parser       Parser
		desc         string
		responseBody string
		statusCode   int
		err          error
		wantErr      error
	}{
		{
			parser:       &JSONParser,
			desc:         "Put 200 with body",
			responseBody: "{}",
			statusCode:   200,
			wantErr:      nil,
		},
		{
			parser:       &JSONParser,
			desc:         "Put 201 no content",
			responseBody: "",
			statusCode:   201,
			wantErr:      nil,
		},
		{
			parser:       &JSONParser,
			desc:         "Put getting 500",
			responseBody: "total chaos",
			statusCode:   500,
			wantErr:      NewHTTPError(500, "total chaos"),
		},
		{
			parser:       &JSONParser,
			desc:         "Http timeout",
			responseBody: "{}",
			statusCode:   0,
			err:          fasthttp.ErrTimeout,
			wantErr:      TimeoutError,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			c := createStub([]byte(tC.responseBody), tC.statusCode, tC.err)
			var resp struct{}
			req := &HTTPRequest{
				URL: "dont/care",
			}
			err := c.Put(context.Background(), tC.parser, req, &resp)
			assert.Equal(t, tC.wantErr, err)
		})
	}
}

func Test_read_json_validation(t *testing.T) {
	testCases := []struct {
		name string
		want error
		data []byte
	}{
		{
			name: "No error when required field exist",
			want: nil,
			data: []byte(`{
						  "some":"thing",
						  "cnt":1000.32,
						  "found":false,
							"validation": "hello"
						}`),
		},
		{
			name: "Error when required field is missing",
			want: errors.New("Validation error: Key: 'testStruct.Validation' Error:Field validation for 'Validation' failed on the 'required' tag"),
			data: []byte(`{
						  "some":"thing",
						  "cnt":1000.32,
						  "found":false
						}`),
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(tt *testing.T) {
			var res testStruct
			parse := JSONParser.Read(&res)
			resp := fasthttp.Response{}
			resp.Header.SetContentLength(len(test.data))
			resp.SetBodyRaw(test.data)
			err := parse(&resp)
			if test.want == nil {
				assert.Nil(tt, err)
			} else {
				assert.EqualError(t, err, test.want.Error())
			}
		})
	}
}

func Test_Plain_parser_read(t *testing.T) {
	data := []byte("returned data")
	var res []byte
	parse := PlainParser.Read(&res)
	resp := fasthttp.Response{}
	resp.Header.SetContentLength(len(data))
	resp.SetBodyRaw(data)
	_ = parse(&resp)
	assert.Equal(t, "returned data", string(res))
}

func Test_Plain_parser_read_error(t *testing.T) {
	data := []byte("returned data")
	var res struct{}
	parse := PlainParser.Read(&res)
	resp := fasthttp.Response{}
	resp.Header.SetContentLength(len(data))
	resp.SetBodyRaw(data)
	err := parse(&resp)
	assert.EqualError(t, err, "Invalid type of target, should be *[]byte")
}

func Test_Plain_parser_write(t *testing.T) {
	parse := PlainParser.Write([]byte("sending data"))
	req := fasthttp.Request{}
	_ = parse(&req)
	assert.Equal(t, "sending data", string(req.Body()))
}

func Test_Plain_parser_write_error(t *testing.T) {
	parse := PlainParser.Write("sending data")
	req := fasthttp.Request{}
	err := parse(&req)
	assert.EqualError(t, err, "Invalid type of content, should be []byte")
}

func Test_retry(t *testing.T) {
	tests := []struct {
		name          string
		maxRetries    int
		retriedErrors []error
		err           error
		called        int
	}{
		{
			name:          "run once",
			maxRetries:    0,
			retriedErrors: nil,
			err:           nil,
			called:        1,
		},
		{
			name:          "run once with no retried errors",
			maxRetries:    1,
			retriedErrors: nil,
			err:           assert.AnError,
			called:        1,
		},
		{
			name:          "run once with error not matching retried errors",
			maxRetries:    1,
			retriedErrors: []error{assert.AnError, errors.New("some error")},
			err:           errors.New("other error"),
			called:        1,
		},
		{
			name:          "retry with error matching retried error",
			maxRetries:    1,
			retriedErrors: []error{assert.AnError},
			err:           assert.AnError,
			called:        2,
		},
		{
			name:          "retry twice with error matching retried error",
			maxRetries:    2,
			retriedErrors: []error{assert.AnError},
			err:           assert.AnError,
			called:        3,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			called := 0

			err := retry(func() error {
				called++
				return test.err
			}, test.maxRetries, test.retriedErrors)

			assert.Equal(t, test.called, called)
			assert.Equal(t, err, test.err)
		})
	}
}

func Benchmark_readJson_parse(b *testing.B) {
	rawJSON := []byte(`{
						  "some":"thing",
						  "cnt":1000.32,
						  "found":false,
							"validation": "hello"
						}`)
	var res testStruct
	parse := JSONParser.Read(&res)
	resp := fasthttp.Response{}
	for i := 0; i < b.N; i++ {
		resp.SetBodyRaw(rawJSON)
		if err := parse(&resp); err != nil {
			assert.FailNow(b, err.Error())
		}
	}
}

func Benchmark_readXml_parse(b *testing.B) {
	rawXML := []byte(`<x>
						<some>thing</some>
						<cnt>123.123</cnt>
						<found>true</found>
						<validation>hello</validation>
					  </x>`)
	var res testStruct
	parse := XMLParser.Read(&res)
	resp := fasthttp.Response{}
	for i := 0; i < b.N; i++ {
		resp.SetBodyRaw(rawXML)
		if err := parse(&resp); err != nil {
			assert.FailNow(b, err.Error())
		}
	}
}
