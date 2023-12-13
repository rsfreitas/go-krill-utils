package errors

import (
	"context"
	"encoding/json"

	"github.com/rsfreitas/go-pocket-utils/logger"
)

// ServiceError is a structure that holds internal error details to improve
// error log description for the end-user, and it implements the errorApi.Error
// interface.
type ServiceError struct {
	err        *Error
	attributes []logger.Attribute
	logger     func(ctx context.Context, msg string, attrs ...logger.Attribute)
}

type serviceErrorOptions struct {
	HideDetails bool
	Code        int32
	Kind        ErrorKind
	ServiceName string
	Message     string
	Destination string
	Logger      func(ctx context.Context, msg string, attrs ...logger.Attribute)
	Error       error
}

func newServiceError(options *serviceErrorOptions) *ServiceError {
	return &ServiceError{
		err: &Error{
			hideDetails:   options.HideDetails,
			Code:          options.Code,
			ServiceName:   options.ServiceName,
			Message:       options.Message,
			Destination:   options.Destination,
			Kind:          options.Kind,
			SublevelError: options.Error,
		},
		logger: options.Logger,
	}
}

func (s *ServiceError) WithCode(code int32) *ServiceError {
	s.err.Code = code
	return s
}

func (s *ServiceError) WithAttributes(attrs ...logger.Attribute) *ServiceError {
	s.attributes = attrs
	return s
}

func (s *ServiceError) Submit(ctx context.Context) error {
	// Display the error message onto the output
	if s.logger != nil {
		logFields := []logger.Attribute{withKind(s.err.Kind)}
		if s.err.SublevelError != nil {
			logFields = append(logFields, logger.Error(s.err.SublevelError))
		}

		s.logger(ctx, s.err.Message, append(logFields, s.attributes...)...)
	}

	// And give back the proper error for the API
	return s.err
}

// withKind wraps an ErrorKind into a structured log Attribute.
func withKind(kind ErrorKind) logger.Attribute {
	return logger.String("error.kind", string(kind))
}

// Error is the framework error type that a service handler should return to
// keep a standard error between services.
type Error struct {
	Code          int32     `json:"code"`
	ServiceName   string    `json:"service_name,omitempty"`
	Message       string    `json:"message,omitempty"`
	Destination   string    `json:"destination,omitempty"`
	Kind          ErrorKind `json:"kind"`
	SublevelError error     `json:"details,omitempty"`

	hideDetails bool
}

func (e *Error) Error() string {
	return e.String()
}

func (e *Error) String() string {
	out := Error{
		Code:        e.Code,
		Destination: e.Destination,
		Kind:        e.Kind,
		Message:     e.Message,
	}

	// The framework can be initialized disabling error message details at the
	// output to avoid showing internal information.
	if !e.hideDetails {
		out.SublevelError = e.SublevelError
		out.ServiceName = e.ServiceName
		out.Destination = e.Destination
	}

	b, _ := json.Marshal(out)
	return string(b)
}
