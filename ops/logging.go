package ops

import (
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

// LogHTTPEvent returned function adds request and response data to the zerolog Event
func LogHTTPEvent(req *fasthttp.Request, resp *fasthttp.Response, err error) func(event *zerolog.Event) {
	return func(event *zerolog.Event) {
		event.Err(err)

		// Google cloud structured logging recognize the following fields:
		// https://cloud.google.com/logging/docs/structured-logging#special-payload-fields
		requestEvent := zerolog.Dict()

		// Standard http request fields
		requestEvent.Bytes("requestUrl", req.URI().FullURI()).
			Bytes("requestMethod", req.Header.Method()).
			Int("status", resp.StatusCode()).
			Bytes("protocol", req.Header.Protocol())

		// Request headers
		requestHeaders := zerolog.Dict()
		if contentType := req.Header.ContentType(); contentType != nil {
			requestHeaders.Bytes("Content-Type", contentType)
		}
		if contentEncoding := req.Header.ContentEncoding(); contentEncoding != nil {
			requestHeaders.Bytes("Content-Encoding", contentEncoding)
		}
		requestEvent.Dict("requestHeaders", requestHeaders)

		// Request body
		if body := req.Body(); body != nil {
			// content is encoded, don't bother decompressing
			if encoding := req.Header.ContentEncoding(); encoding != nil {
				requestEvent.Bytes("request", encoding)
			} else if contentType := req.Header.ContentType(); isContentTypeLogged(contentType) {
				requestEvent.RawJSON("request", body)
			}
			requestEvent.Int("requestSize", req.Header.ContentLength())
		}

		// Response headers
		responseHeaders := zerolog.Dict()
		if contentType := resp.Header.ContentType(); contentType != nil {
			responseHeaders.Bytes("Content-Type", contentType)
		}
		if contentEncoding := resp.Header.ContentEncoding(); contentEncoding != nil {
			responseHeaders.Bytes("Content-Encoding", contentEncoding)
		}
		requestEvent.Dict("responseHeaders", responseHeaders)

		// Response body
		if body := resp.Body(); body != nil {
			// content is encoded, don't bother decompressing
			if encoding := resp.Header.ContentEncoding(); encoding != nil {
				requestEvent.Bytes("response", encoding)
			} else if contentType := resp.Header.ContentType(); isContentTypeLogged(contentType) {
				requestEvent.RawJSON("response", body)
			}
			requestEvent.Int("responseSize", resp.Header.ContentLength())
		}
		event.Dict("httpRequest", requestEvent)
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
