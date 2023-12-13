package response

import (
	"encoding/json"
	"net/http"
)

var knownServiceErrors = map[string]bool{
	"ValidationError": true,
	"InternalError":   true,
	"NotFoundError":   true,
	"ConditionError":  true,
	"PermissionError": true,
}

type serviceError struct {
	Code          int32       `json:"code"`
	ServiceName   string      `json:"service_name"`
	Message       string      `json:"message"`
	Destination   string      `json:"destination"`
	Kind          string      `json:"kind"`
	SublevelError interface{} `json:"details"`
}

func serviceErrorFromString(s string) (*serviceError, error) {
	var e serviceError
	if err := json.Unmarshal([]byte(s), &e); err != nil {
		return nil, err
	}

	return &e, nil
}

func (s *serviceError) IsKnownError() bool {
	_, ok := knownServiceErrors[s.Kind]
	return ok
}

func (s *serviceError) ResponseCode() int {
	switch s.Kind {
	case "ValidationError":
		return http.StatusBadRequest
	case "NotFoundError":
		return http.StatusNotFound
	case "ConditionError":
		return http.StatusPreconditionFailed
	case "PermissionError":
		return http.StatusUnauthorized
	}

	return http.StatusInternalServerError
}

func (s *serviceError) ToResponseError() *responseError {
	opt := &responseErrorOptions{
		Code:        int(s.Code),
		Source:      s.ServiceName,
		Message:     s.Message,
		Destination: s.Destination,
	}

	if s.SublevelError != nil {
		if s.Kind == "ValidationError" {
			// Encode the error details into a json string
			b, _ := json.Marshal(s.SublevelError)
			err := string(b)[1 : len(b)-1]
			opt.Fields = newValidationErrorFields(err)
		}
	}

	return newResponseError(opt)
}
