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
		// Google cloud structured logging recognize the following fields:
		// https://cloud.google.com/logging/docs/structured-logging#special-payload-fields
		requestDict := zerolog.Dict()

		// Standard http request fields
		requestDict.Bytes("requestUrl", req.URI().FullURI()).
			Bytes("requestMethod", req.Header.Method()).
			Bytes("protocol", req.Header.Protocol())

		// Request headers
		requestHeaders := zerolog.Dict()
		if contentType := req.Header.ContentType(); contentType != nil {
			requestHeaders.Bytes("Content-Type", contentType)
		}
		if contentEncoding := req.Header.ContentEncoding(); contentEncoding != nil {
			requestHeaders.Bytes("Content-Encoding", contentEncoding)
		}
		requestDict.Dict("requestHeaders", requestHeaders)

		// Request body
		if body := req.Body(); body != nil {
			// content is encoded, don't bother decompressing
			if encoding := req.Header.ContentEncoding(); encoding != nil {
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

		event.Dict("httpRequest", requestDict)
	}
}

// LogHTTPResponse return a function that adds response data to the zerolog Event
func LogHTTPResponse(resp *fasthttp.Response, err error) func(event *zerolog.Event) {
	return func(event *zerolog.Event) {
		event.Err(err)

		// Google cloud structured logging recognize the following fields:
		// https://cloud.google.com/logging/docs/structured-logging#special-payload-fields
		responseDict := zerolog.Dict()

		// Standard http response fields
		responseDict.Int("status", resp.StatusCode())

		// Response headers
		responseHeaders := zerolog.Dict()
		if contentType := resp.Header.ContentType(); contentType != nil {
			responseHeaders.Bytes("Content-Type", contentType)
		}
		if contentEncoding := resp.Header.ContentEncoding(); contentEncoding != nil {
			responseHeaders.Bytes("Content-Encoding", contentEncoding)
		}
		responseDict.Dict("responseHeaders", responseHeaders)

		// Response body
		if body := resp.Body(); body != nil {
			// content is encoded, don't bother decompressing
			if encoding := resp.Header.ContentEncoding(); encoding != nil {
				responseDict.Bytes("response", encoding)
			} else if contentType := resp.Header.ContentType(); isContentTypeLogged(contentType) {
				if isContentTypeJSON(contentType) {
					responseDict.RawJSON("response", body)
				} else {
					responseDict.Bytes("response", body)
				}
			}
			responseDict.Int("responseSize", resp.Header.ContentLength())
		}
		event.Dict("httpResponse", responseDict)
	}
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
