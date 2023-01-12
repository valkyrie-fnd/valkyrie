package ops

import (
	"bytes"
	"context"

	"github.com/valyala/fasthttp"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

type userContextProvider interface {
	UserContext() context.Context
}

// AddLoggingContext adds the key and value to all logging messages in this request context.
func AddLoggingContext(c userContextProvider, key, value string) {
	logger := zerolog.Ctx(c.UserContext())
	logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Str(key, value)
	})
}

// LogHTTPRequest return a function that adds request data to the zerolog Event
func LogHTTPRequest(req *fasthttp.Request) func(event *zerolog.Event) {
	return func(event *zerolog.Event) {
		event.Bytes("requestUrl", req.URI().FullURI()).
			Bytes("requestMethod", req.Header.Method()).
			Bytes("protocol", req.Header.Protocol()).
			Bytes("userAgent", req.Header.UserAgent())
	}
}

// LogHTTPResponse return a function that adds request and response data to the zerolog Event
func LogHTTPResponse(req *fasthttp.Request, resp *fasthttp.Response, err error) func(event *zerolog.Event) {
	return func(event *zerolog.Event) {
		event.Err(err)

		// Google cloud structured logging recognize special payload fields below, which
		// allows for automatic highlighting in UI.
		// https://cloud.google.com/logging/docs/structured-logging#special-payload-fields
		requestDict := zerolog.Dict()

		LogHTTPRequest(req)(requestDict)

		logRequestHeaders(req, requestDict)
		logRequestBody(req, requestDict)

		logResponseHeaders(resp, requestDict)
		logResponseBody(resp, requestDict)

		event.Dict("httpRequest", requestDict)
	}
}

func logRequestHeaders(req *fasthttp.Request, requestDict *zerolog.Event) {
	// Request headers
	requestHeaders := zerolog.Dict()
	if contentType := req.Header.ContentType(); len(contentType) > 0 {
		requestHeaders.Bytes("Content-Type", contentType)
	}
	if contentEncoding := req.Header.ContentEncoding(); len(contentEncoding) > 0 {
		requestHeaders.Bytes("Content-Encoding", contentEncoding)
	}
	requestDict.Dict("requestHeaders", requestHeaders)
}

func logRequestBody(req *fasthttp.Request, requestDict *zerolog.Event) {
	if body := req.Body(); body != nil {
		// content is encoded, don't bother decompressing
		if encoding := req.Header.ContentEncoding(); len(encoding) > 0 {
			requestDict.Bytes("request", encoding)
		} else if contentType := req.Header.ContentType(); isContentTypeLogged(contentType) {
			if isContentTypeJSON(contentType) {
				requestDict.RawJSON("request", body)
			} else {
				requestDict.Bytes("request", body)
			}
		}
		requestDict.Int("requestSize", req.Header.ContentLength())
	}
}

func logResponseBody(resp *fasthttp.Response, requestDict *zerolog.Event) {
	if body := resp.Body(); body != nil {
		// content is encoded, don't bother decompressing
		if encoding := resp.Header.ContentEncoding(); len(encoding) > 0 {
			requestDict.Bytes("response", encoding)
		} else if contentType := resp.Header.ContentType(); isContentTypeLogged(contentType) {
			if isContentTypeJSON(contentType) {
				requestDict.RawJSON("response", body)
			} else {
				requestDict.Bytes("response", body)
			}
		}
		requestDict.Int("responseSize", resp.Header.ContentLength())
	}
}

func logResponseHeaders(resp *fasthttp.Response, requestDict *zerolog.Event) {
	responseHeaders := zerolog.Dict()
	if contentType := resp.Header.ContentType(); len(contentType) > 0 {
		responseHeaders.Bytes("Content-Type", contentType)
	}
	if contentEncoding := resp.Header.ContentEncoding(); len(contentEncoding) > 0 {
		responseHeaders.Bytes("Content-Encoding", contentEncoding)
	}
	requestDict.Dict("responseHeaders", responseHeaders)
}

var loggedContentTypes = map[string]struct{}{
	fiber.MIMEApplicationJSON:            {},
	fiber.MIMETextPlain:                  {},
	fiber.MIMEApplicationXML:             {},
	fiber.MIMEApplicationForm:            {},
	fiber.MIMEMultipartForm:              {},
	"application/vnd.kafka.json.v2+json": {},
	"application/vnd.kafka.v2+json":      {},
}

// isContentTypeLogged returns true for the content types that should be logged.
// This function skips verbose content such as html, images, octet streams etc.
func isContentTypeLogged(contentType []byte) bool {
	_, found := loggedContentTypes[string(contentType)]
	return found
}

// isContentTypeJSON returns true for content types that contains json data, otherwise false.
func isContentTypeJSON(contentType []byte) bool {
	return bytes.Contains(contentType, []byte("/json")) || bytes.Contains(contentType, []byte("+json"))
}
