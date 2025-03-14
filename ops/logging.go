package ops

import (
	"bytes"
	"context"
	"regexp"
	"strings"

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

// logHTTPRequest return a function that adds request data to the zerolog Event
func logHTTPRequest(req *fasthttp.Request) func(event *zerolog.Event) {
	return func(event *zerolog.Event) {
		event.Bytes("requestUrl", req.URI().FullURI()).
			Bytes("requestMethod", req.Header.Method()).
			Bytes("protocol", req.Header.Protocol()).
			Bytes("userAgent", req.Header.UserAgent())

		logRequestHeaders(req, event)
		logRequestBody(req, event)
	}
}

// logHTTPResponse return a function that adds request and response data to the zerolog Event
func logHTTPResponse(req *fasthttp.Request, resp *fasthttp.Response, err error) func(event *zerolog.Event) {
	return func(event *zerolog.Event) {
		event.Err(err)

		// Google cloud structured logging recognize special payload fields below, which
		// allows for automatic highlighting in UI.
		// https://cloud.google.com/logging/docs/structured-logging#special-payload-fields
		requestDict := zerolog.Dict()

		logHTTPRequest(req)(requestDict)

		requestDict.Int("status", resp.StatusCode())

		logResponseHeaders(resp, requestDict)
		logResponseBody(resp, requestDict)

		event.Dict("httpRequest", requestDict)
	}
}

func logRequestHeaders(req *fasthttp.Request, requestDict *zerolog.Event) {
	// Request headers
	requestHeaders := zerolog.Dict()
	req.Header.VisitAll(func(key, value []byte) {
		if isHeaderLogged(key) {
			requestHeaders.Bytes(string(key), value)
		}
	})
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
	resp.Header.VisitAll(func(key, value []byte) {
		if isHeaderLogged(key) {
			responseHeaders.Bytes(string(key), value)
		}
	})
	requestDict.Dict("responseHeaders", responseHeaders)
}

func isHeaderLogged(header []byte) bool {
	header = bytes.ToLower(header)
	for _, re := range loggedHeaders {
		if re.Match(header) {
			return true
		}
	}

	return false
}

// isContentTypeLogged returns true for the content types that should be logged.
// This function skips verbose content such as html, images, octet streams etc.
func isContentTypeLogged(contentType []byte) bool {
	contentType = bytes.ToLower(contentType)
	// Handle headers like "application/json; charset=utf-8"
	prefix, _, _ := bytes.Cut(contentType, []byte(";"))

	for _, re := range loggedContentTypes {
		if re.Match(prefix) {
			return true
		}
	}

	return false
}

// isContentTypeJSON returns true for content types that contains json data, otherwise false.
func isContentTypeJSON(contentType []byte) bool {
	contentType = bytes.ToLower(contentType)
	return bytes.Contains(contentType, []byte("/json")) || bytes.Contains(contentType, []byte("+json"))
}

// Set default values for Header and Content-Type logging filters.
func init() {
	SetHeaderWhitelist([]string{
		`Content-Type`,
		`Content-Encoding`,
		`X-Forwarded-For`,
		`X-Correlation-ID`,
		`traceparent`,
		`X-Msg-Timestamp`,
		`X-Request-Id`,
	})

	SetContentTypeWhitelist([]string{
		fiber.MIMEApplicationJSON,
		fiber.MIMETextPlain,
		fiber.MIMEApplicationXML,
		fiber.MIMEApplicationForm,
		fiber.MIMEMultipartForm,
		`application/vnd.kafka.json.v2+json`,
		`application/vnd.kafka.v2+json`,
	})
}

var loggedHeaders = []*regexp.Regexp{}
var loggedContentTypes = []*regexp.Regexp{}

func SetHeaderWhitelist(headers []string) {
	loggedHeaders = make([]*regexp.Regexp, 0, len(headers))
	for _, header := range headers {
		loggedHeaders = append(loggedHeaders, wildcardToRegexp(header))
	}
}

func SetContentTypeWhitelist(contentTypes []string) {
	loggedContentTypes = make([]*regexp.Regexp, 0, len(contentTypes))
	for _, contentType := range contentTypes {
		loggedContentTypes = append(loggedContentTypes, wildcardToRegexp(contentType))
	}
}

func wildcardToRegexp(wildcard string) *regexp.Regexp {
	wildcard = strings.ToLower(wildcard)
	pattern := strings.ReplaceAll(`^`+regexp.QuoteMeta(wildcard)+`$`, `\*`, `.*`)
	return regexp.MustCompile(pattern)
}
