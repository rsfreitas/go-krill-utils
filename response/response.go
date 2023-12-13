package response

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/valyala/fasthttp"
	"google.golang.org/grpc/status"
)

const (
	customHeaderPrefix = "handler-attribute-"
	customResponseCode = "handler-response-code"
)

var contentTypeHeader = http.CanonicalHeaderKey("Content-Type")

// Responser is a behavior that a struct may have to format its fields
// in case of an HTTP response.
type Responser interface {
	HttpResponse() interface{}
}

type Response struct {
	serviceName string
	contentType string
	ctx         context.Context
	errors      *ErrorOptions
}

type Options struct {
	ServiceName string
	Errors      *ErrorOptions
}

type ErrorOptions struct {
	ShowInternalMessages bool
	ShowSource           bool
}

// New creates a new response container for HTTP handlers return data using a
// specific standard.
func New(ctx context.Context, options *Options) *Response {
	var contentType string

	if fctx, ok := ctx.(*fasthttp.RequestCtx); ok {
		contentType = string(fctx.Request.Header.Peek(contentTypeHeader))
	}

	return &Response{
		serviceName: options.ServiceName,
		contentType: contentType,
		ctx:         ctx,
		errors:      options.Errors,
	}
}

func (r *Response) ForwardAuthenticationError(err error) {
	ferror, err := serviceErrorFromString(err.Error())
	if err != nil {
		r.forwardOutput(fasthttp.StatusInternalServerError,
			newResponseError(&responseErrorOptions{
				Message: internalServerErrorMsg,
				Details: err.Error(),
			}),
		)

		return
	}
	if ferror.IsKnownError() {
		r.forwardOutput(ferror.ResponseCode(), ferror.ToResponseError())
		return
	}
}

func (r *Response) ForwardError(err error) {
	ferror, err := serviceErrorFromString(err.Error())
	if err != nil {
		r.forwardOutput(fasthttp.StatusInternalServerError,
			newResponseError(&responseErrorOptions{
				Message: internalServerErrorMsg,
				Details: err.Error(),
			}),
		)

		return
	}
	if ferror.IsKnownError() {
		r.forwardOutput(ferror.ResponseCode(), ferror.ToResponseError())
		return
	}

	// A gRPC service can send "gRPC" errors in case of unexpected errors
	if sts, ok := status.FromError(err); ok {
		r.forwardOutput(fasthttp.StatusInternalServerError,
			newResponseError(&responseErrorOptions{
				Message: internalServerErrorMsg,
				Details: sts.Message(),
			}),
		)

		return
	}

	// In case some parsing failed.
	if res, ok := jsonError(err); ok {
		r.forwardOutput(fasthttp.StatusBadRequest, res)
		return
	}

	// Forward the original error if none of the above error checks were
	// successful.
	r.forwardOutput(fasthttp.StatusInternalServerError,
		newResponseError(&responseErrorOptions{
			Source:  r.serviceName,
			Message: internalServerErrorMsg,
			Details: err.Error(),
		}),
	)
}

func (r *Response) ForwardSuccess(data interface{}) {
	// Does the message have another format to send as response?
	if h, ok := data.(Responser); ok {
		data = h.HttpResponse()
	}

	r.forwardOutput(fasthttp.StatusOK, data)
}

func (r *Response) setCustomHeaders(ctx *fasthttp.RequestCtx) {
	// Set all handler's custom header values.
	ctx.VisitUserValues(func(key []byte, value interface{}) {
		if strings.HasPrefix(string(key), customHeaderPrefix) {
			ctx.Response.Header.Set(strings.TrimPrefix(string(key), customHeaderPrefix), value.(string))
		}
	})
}

func (r *Response) forwardOutput(statusCode int, data interface{}) {
	out, err := json.Marshal(data)
	if err != nil {
		r.ForwardError(err)
		return
	}

	if fctx, ok := r.ctx.(*fasthttp.RequestCtx); ok {
		r.setCustomHeaders(fctx)

		if v := fctx.UserValue(customResponseCode); v != nil {
			if c, ok := v.(int); ok {
				statusCode = c
			}
		}

		fctx.Response.SetStatusCode(statusCode)
		fctx.Response.Header.SetContentType(r.contentType)
		fctx.Response.SetBodyRaw(out)
		return
	}
}

func (r *Response) SetContentType(contentType string) {
	r.contentType = contentType
}
