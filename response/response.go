package response

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc/status"
)

const (
	customHeaderPrefix = "handler-attribute-"
	customResponseCode = "handler-response-code"
)

var contentTypeHeader = http.CanonicalHeaderKey("Content-Type")

// ResponserFasthttp is a behavior that a struct may have to format its fields
// in case of an HTTP response.
type ResponserFasthttp interface {
	HttpResponse() interface{}
}

type ResponserEcho interface {
	HttpResponseBytes() ([]byte, error)
}

type Response struct {
	customCode  int
	serviceName string
	contentType string
	ctx         interface{}
}

type Options struct {
	ServiceName string
}

// NewFromFasthttp creates a new response container for HTTP handlers return data using a
// specific standard.
func NewFromFasthttp(ctx *fasthttp.RequestCtx, options *Options) *Response {
	return &Response{
		serviceName: options.ServiceName,
		contentType: string(ctx.Request.Header.Peek(contentTypeHeader)),
		ctx:         ctx,
	}
}

func NewFromEcho(ctx echo.Context, options *Options) *Response {
	return &Response{
		serviceName: options.ServiceName,
		contentType: "application/json",
		ctx:         ctx,
	}
}

func (r *Response) ForwardAuthenticationError(err error) error {
	ferror, err := serviceErrorFromString(err.Error())
	if err != nil {
		return r.forwardOutput(fasthttp.StatusInternalServerError,
			newResponseError(&responseErrorOptions{
				Message: internalServerErrorMsg,
				Details: err.Error(),
			}),
		)
	}
	if ferror.IsKnownError() {
		return r.forwardOutput(ferror.ResponseCode(), ferror.ToResponseError())
	}

	return nil
}

func (r *Response) ForwardError(err error) error {
	ferror, err := serviceErrorFromString(err.Error())
	if err != nil {
		return r.forwardOutput(fasthttp.StatusInternalServerError,
			newResponseError(&responseErrorOptions{
				Message: internalServerErrorMsg,
				Details: err.Error(),
			}),
		)
	}
	if ferror.IsKnownError() {
		return r.forwardOutput(ferror.ResponseCode(), ferror.ToResponseError())
	}

	// A gRPC service can send "gRPC" errors in case of unexpected errors
	if sts, ok := status.FromError(err); ok {
		return r.forwardOutput(fasthttp.StatusInternalServerError,
			newResponseError(&responseErrorOptions{
				Message: internalServerErrorMsg,
				Details: sts.Message(),
			}),
		)
	}

	// In case some parsing failed.
	if res, ok := jsonError(err); ok {
		return r.forwardOutput(fasthttp.StatusBadRequest, res)
	}

	// Forward the original error if none of the above error checks were
	// successful.
	return r.forwardOutput(fasthttp.StatusInternalServerError,
		newResponseError(&responseErrorOptions{
			Source:  r.serviceName,
			Message: internalServerErrorMsg,
			Details: err.Error(),
		}),
	)
}

func (r *Response) ForwardSuccess(data interface{}) error {
	if _, ok := r.ctx.(*fasthttp.RequestCtx); ok {
		// Does the message have another format to send as response?
		if h, ok := data.(ResponserFasthttp); ok {
			data = h.HttpResponse()
		}

		return r.forwardOutput(fasthttp.StatusOK, data)
	}

	if _, ok := r.ctx.(echo.Context); ok {
		if h, ok := data.(ResponserEcho); ok {
			b, err := h.HttpResponseBytes()
			if err != nil {
				return err
			}

			return r.forwardOutput(fasthttp.StatusOK, string(b))
		}
	}

	return nil
}

func (r *Response) forwardOutput(statusCode int, data interface{}) error {
	if fctx, ok := r.ctx.(*fasthttp.RequestCtx); ok {
		out, err := json.Marshal(data)
		if err != nil {
			return r.ForwardError(err)
		}

		r.setFasthttpCustomHeaders(fctx)

		if v := fctx.UserValue(customResponseCode); v != nil {
			if c, ok := v.(int); ok {
				statusCode = c
			}
		}

		fctx.Response.SetStatusCode(statusCode)
		fctx.Response.Header.SetContentType(r.contentType)
		fctx.Response.SetBodyRaw(out)

		return nil
	}

	if ectx, ok := r.ctx.(echo.Context); ok {
		if r.customCode != 0 {
			statusCode = r.customCode
		}

		ectx.Response().Header().Set("Content-Type", r.contentType)
		out, ok := data.(string)
		if !ok {
			b, err := json.Marshal(data)
			if err != nil {
				return r.ForwardError(err)
			}
			out = string(b)
		}

		if err := ectx.String(statusCode, out); err != nil {
			return err
		}
	}

	return nil
}

func (r *Response) setFasthttpCustomHeaders(ctx *fasthttp.RequestCtx) {
	// Set all handler's custom header values.
	ctx.VisitUserValues(func(key []byte, value interface{}) {
		if strings.HasPrefix(string(key), customHeaderPrefix) {
			ctx.Response.Header.Set(strings.TrimPrefix(string(key), customHeaderPrefix), value.(string))
		}
	})
}

func (r *Response) SetContentType(contentType string) {
	r.contentType = contentType
}

func SetResponseCode(ctx context.Context, code int) {
	if c, ok := ctx.(*fasthttp.RequestCtx); ok {
		c.SetUserValue(customResponseCode, code)
		return
	}

	r := RetrieveFromContext(ctx)
	r.customCode = code
}

func AppendResponseToContext(ctx context.Context, r *Response) context.Context {
	return context.WithValue(ctx, "response", r)
}

func RetrieveFromContext(ctx context.Context) *Response {
	r := ctx.Value("response")
	return r.(*Response)
}
