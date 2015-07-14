package webserver

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/yangchenxing/cangshan/logging"
)

// A Request present a webserver request
type Request struct {
	*http.Request

	Attr         map[string]interface{}
	Param        map[string]interface{}
	response     http.ResponseWriter
	status       int
	content      bytes.Buffer
	contentType  string
	receiveTime  time.Time
	logFormatter *logging.Formatter
	done         bool
	stopped      bool
}

func newRequest(request *http.Request, response http.ResponseWriter, formatter *logging.Formatter) *Request {
	req := &Request{
		Request:      request,
		Attr:         make(map[string]interface{}),
		Param:        make(map[string]interface{}),
		response:     response,
		receiveTime:  time.Now(),
		logFormatter: formatter,
	}
	return req
}

// ResponseHeader returns response header map that will be sent
func (request *Request) ResponseHeader() http.Header {
	return request.response.Header()
}

// SetCookie add a Set-Cookie header to http response
func (request *Request) SetCookie(cookie *http.Cookie) {
	http.SetCookie(request.response, cookie)
}

// Write set or overwrite response status, content and content type that will be sent.
func (request *Request) Write(status int, content []byte, contentType string) error {
	request.status = status
	request.content.Reset()
	if content != nil {
		if _, err := request.content.Write(content); err != nil {
			return fmt.Errorf("Write buffer fail: %s", err.Error())
		}
	}
	if contentType != "" {
		request.ResponseHeader().Set("Content-Type", contentType)
	} else {
		request.ResponseHeader().Del("Content-Type")
	}
	return nil
}

// Stop the processing of the request and set the response status, content and content type.
// No handler will handle the request after stopped.
func (request *Request) WriteAndStop(status int, content []byte, contentType string) error {
	request.stopped = true
	return request.Write(status, content, contentType)
}

func (request *Request) Stop() {
	request.stopped = true
}

func (request *Request) buildResponse() error {
	if !request.done {
		request.response.WriteHeader(request.status)
		if _, err := request.response.Write(request.content.Bytes()); err != nil {
			return fmt.Errorf("Write response content fail: %s", err.Error())
		}
		request.logAccess()
		request.done = true
	}
	return nil
}

func (request *Request) logAccess() {
	var err error
	if request.Attr["request.clientip"], _, err = net.SplitHostPort(request.RemoteAddr); err != nil {
		request.Attr["request.clientip"] = request.RemoteAddr
	}
	if _, found := request.Attr["request.user"]; !found {
		request.Attr["request.user"] = "-"
	}
	if _, found := request.Attr["request.auth"]; !found {
		request.Attr["request.auth"] = "-"
	}
	request.Attr["request.time"] = request.receiveTime.Format("[02/Jan/2006:15:04:05 -0700]")
	request.Attr["request.method"] = request.Method
	request.Attr["request.url"] = request.URL.String()
	request.Attr["request.proto"] = request.Proto
	request.Attr["request.status"] = request.status
	request.Attr["request.bodylen"] = request.content.Len()
	request.Attr["request.referer"] = request.Referer()
	request.Attr["request.useragent"] = request.UserAgent()
	request.Attr["request.timecost"] = time.Now().Sub(request.receiveTime)
	logging.LogEx(2, "access", nil, request.Attr, "")
}

// Debug write debug log with web server specified log formatter
func (request *Request) Debug(format string, params ...interface{}) {
	logging.LogEx(2, "debug", request.logFormatter, request.Attr, format, params...)
}

// Info write info log with web server specified log formatter
func (request *Request) Info(format string, params ...interface{}) {
	logging.LogEx(2, "info", request.logFormatter, request.Attr, format, params...)
}

// Warn write warn log with web server specified log formatter
func (request *Request) Warn(format string, params ...interface{}) {
	logging.LogEx(2, "warn", request.logFormatter, request.Attr, format, params...)
}

// Error write error log with web server specified log formatter
func (request *Request) Error(format string, params ...interface{}) {
	logging.LogEx(2, "error", request.logFormatter, request.Attr, format, params...)
}

// Fatal write fatal log with web server specified log formatter
func (request *Request) Fatal(format string, params ...interface{}) {
	logging.LogEx(2, "fatal", request.logFormatter, request.Attr, format, params...)
}
