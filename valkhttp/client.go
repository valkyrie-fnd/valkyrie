// Package valkhttp provides convenience and sensible defaults for some low level http client
package valkhttp

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/internal/pipeline"

	"github.com/go-playground/validator/v10"

	"github.com/goccy/go-json"
	"github.com/valyala/fasthttp"
)

var (
	headerContentTypeJSON = []byte("application/json")
	headerContentTypeXML  = []byte("application/xml")
	headerContentTypeText = []byte("text/plain")
	validate              = validator.New()

	// Pipeline is used to allow for custom Handler functions (such as access logging or tracing)
	// to be registered and run before actual HTTP calls.
	Pipeline = pipeline.NewPipeline[PipelinePayload]()
)

type PipelinePayload struct {
	request  *fasthttp.Request
	response *fasthttp.Response
}

func (p PipelinePayload) Request() *fasthttp.Request {
	return p.request
}

func (p PipelinePayload) Response() *fasthttp.Response {
	return p.response
}

// fastHTTPClient interface of used methods of fasthttp.Client
type fastHTTPClient interface {
	DoTimeout(req *fasthttp.Request, resp *fasthttp.Response, timeout time.Duration) error
}
type Client struct {
	fastClient fastHTTPClient
	config     configs.HTTPClientConfig
}

func Create(config configs.HTTPClientConfig) *Client {
	return &Client{
		config: config,
		fastClient: &fasthttp.Client{
			ReadTimeout:                   config.ReadTimeout,
			WriteTimeout:                  config.WriteTimeout,
			MaxIdleConnDuration:           config.IdleTimeout,
			NoDefaultUserAgentHeader:      true, // Don't send: User-Agent: fasthttp
			DisableHeaderNamesNormalizing: true, // If you set the case on your headers correctly you can enable this
			DisablePathNormalizing:        true,
			RetryIfErr: func(_ *fasthttp.Request, _ int, _ error) (resetTimeout bool, retry bool) {
				return false, false // Disable automatic retries for GET/PATCH/PUT
			},
			MaxIdemponentCallAttempts: 1,
			// increase DNS cache time to an hour instead of default minute
			Dial: (&fasthttp.TCPDialer{
				Concurrency:      4096,
				DNSCacheDuration: time.Hour,
			}).Dial,
		},
	}
}

// Parser used to parse response and write to request body
type Parser interface {
	Read(target any) responseParseFn
	Write(content any) requestContentFn
}

type responseParseFn func(*fasthttp.Response) error
type requestContentFn func(*fasthttp.Request) error

type jsonParser struct{}
type xmlParser struct{}
type plainParser struct{}

// PlainParser just read and writes []byte
var PlainParser = plainParser{}

// JSONParser JSON marshal request and unmarshal response
var JSONParser = jsonParser{}

// XMLParser XML marshal request and unmarshal response
var XMLParser = xmlParser{}

// Convenience method for reading response JSON
func (p *jsonParser) Read(target interface{}) responseParseFn {
	return func(r *fasthttp.Response) error {
		if r.Header.ContentLength() <= 0 {
			return nil
		}
		err := json.Unmarshal(r.Body(), &target)
		if err != nil {
			return fmt.Errorf("json parsing error: %w", err)
		}
		err = validate.Struct(target)
		if err != nil {
			return fmt.Errorf("validation error: %w", err)
		}
		return nil
	}
}

// Write json marshaling the content to the request body
func (p *jsonParser) Write(content interface{}) requestContentFn {
	return func(r *fasthttp.Request) error {
		if len(r.Header.ContentType()) == 0 {
			r.Header.SetContentTypeBytes(headerContentTypeJSON)
		}
		bs, err := json.Marshal(content)
		if err != nil {
			return err
		}

		r.SetBodyRaw(bs)

		return nil
	}
}

// Convenience method for reading response XML
func (p *xmlParser) Read(target interface{}) responseParseFn {
	return func(r *fasthttp.Response) error {
		if r.Header.ContentLength() <= 0 {
			return nil
		}
		err := xml.Unmarshal(r.Body(), &target)
		if err != nil {
			return fmt.Errorf("xml parsing error: %w", err)
		}
		err = validate.Struct(target)
		if err != nil {
			return fmt.Errorf("validation error: %w", err)
		}
		return nil
	}
}

// Write xml marshaling the content to the request body
func (p *xmlParser) Write(content interface{}) requestContentFn {
	return func(r *fasthttp.Request) error {
		if len(r.Header.ContentType()) == 0 {
			r.Header.SetContentTypeBytes(headerContentTypeXML)
		}
		bs, err := xml.Marshal(content)
		if err != nil {
			return err
		}

		r.SetBodyRaw(bs)

		return nil
	}
}

// Read will read response body target *[]byte.
func (p *plainParser) Read(target any) responseParseFn {
	return func(r *fasthttp.Response) error {
		t, ok := target.(*[]byte)
		if !ok {
			return fmt.Errorf("invalid type of target, should be *[]byte")
		}
		*t = r.Body()
		return nil
	}
}

// Write will write content []byte into request body
func (p *plainParser) Write(content any) requestContentFn {
	return func(r *fasthttp.Request) error {
		c, ok := content.([]byte)
		if !ok {
			return fmt.Errorf("invalid type of content, should be []byte")
		}
		if len(r.Header.ContentType()) == 0 {
			r.Header.SetContentTypeBytes(headerContentTypeText)
		}
		r.SetBodyRaw(c)
		return nil
	}
}

// HTTPRequest represents a http request
type HTTPRequest struct {
	Body    any
	Headers map[string]string
	Query   map[string]string
	URL     string
}

// HTTPClient interface for client where user can provide Parser for request and response
type HTTPClient interface {
	Get(ctx context.Context, p Parser, req *HTTPRequest, resp any) error
	Post(ctx context.Context, p Parser, req *HTTPRequest, resp any) error
	Put(ctx context.Context, p Parser, req *HTTPRequest, resp any) error
}

// Get Issue Get request  with expected response body set to resp
func (c *Client) Get(ctx context.Context, p Parser, req *HTTPRequest, resp any) error {
	return c.get(ctx, req.URL, p.Read(resp), req.Headers, req.Query)
}

// Post issue Post request with expected response body set to resp
func (c *Client) Post(ctx context.Context, p Parser, req *HTTPRequest, resp any) error {
	return c.post(ctx, req.URL, p.Write(req.Body), p.Read(resp), req.Headers, req.Query)
}

// Put issue Put request  with expected response body set to resp
func (c *Client) Put(ctx context.Context, p Parser, req *HTTPRequest, resp any) error {
	return c.put(ctx, req.URL, p.Write(req.Body), p.Read(resp), req.Headers, req.Query)
}

func (c *Client) post(
	ctx context.Context,
	uri string,
	bodyFn requestContentFn,
	parseFn responseParseFn,
	headers map[string]string,
	query map[string]string) error {
	return c.handle(ctx, uri, bodyFn, parseFn, fasthttp.MethodPost, headers, query)
}

func (c *Client) put(
	ctx context.Context,
	uri string,
	bodyFn requestContentFn,
	parseFn responseParseFn,
	headers map[string]string,
	query map[string]string) error {
	return c.handle(ctx, uri, bodyFn, parseFn, fasthttp.MethodPut, headers, query)
}

func (c *Client) get(
	ctx context.Context,
	uri string,
	parseFn responseParseFn,
	headers map[string]string,
	query map[string]string) error {
	return c.handle(ctx, uri, nil, parseFn, fasthttp.MethodGet, headers, query)
}

const maxRetries = 1

var retriedErrors = []error{
	// Retry ErrConnectionClosed, caused by server closing keepalive connection
	// before notifying client
	fasthttp.ErrConnectionClosed,
}

func (c *Client) handle(
	ctx context.Context,
	uri string,
	bodyFn requestContentFn,
	parseFn responseParseFn,
	method string,
	headers map[string]string,
	query map[string]string) error {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.SetRequestURI(uri)

	req.Header.SetMethod(method)
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	for k, v := range query {
		req.URI().QueryArgs().Add(k, v)
	}

	if bodyFn != nil {
		if err := bodyFn(req); err != nil {
			return err
		}
	}

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err := Pipeline.Execute(ctx,
		PipelinePayload{req, resp},
		func(pc pipeline.PipelineContext[PipelinePayload]) error {
			return retry(func() error {
				return c.fastClient.DoTimeout(pc.Payload().Request(), pc.Payload().Response(), c.config.RequestTimeout)
			}, maxRetries, retriedErrors)
		})

	statusCode := resp.StatusCode()
	if err == nil {
		return handleResponse(statusCode, resp, parseFn)
	}

	return handleError(err)
}

func handleResponse(statusCode int, resp *fasthttp.Response, parseFn responseParseFn) error {
	switch statusCode {
	case http.StatusOK:
		return parseFn(resp)
	case http.StatusCreated:
		if resp.Header.ContentLength() > 0 {
			return parseFn(resp)
		}
	case http.StatusAccepted:
		return nil
	default:
		// if possible, still populate using response body if there is one
		_ = parseFn(resp)
		return NewHTTPError(statusCode, string(resp.Body()))
	}
	return nil
}

// retry will run call() and check its returned error. Errors matching any of retriedErrors
// are retried up to maxRetries amount of times.
func retry(call func() error, maxRetries int, retriedErrors []error) (err error) {
	for r := 0; r <= maxRetries; r++ {
		err = call()

		switch {
		case err == nil:
			// never retry successful calls
			return err
		case !containsError(err, retriedErrors):
			// don't retry errors not part of retriedErrors
			return err
		}
	}

	return err
}

// containsError returns true if checkedErrors contains err, otherwise false
func containsError(err error, checkedErrors []error) bool {
	for _, e := range checkedErrors {
		if errors.Is(err, e) {
			return true
		}
	}
	return false
}

func handleError(err error) error {
	if errors.Is(err, fasthttp.ErrTimeout) {
		return TimeoutError // don't leak fasthttp timeout error
	}

	return err
}
