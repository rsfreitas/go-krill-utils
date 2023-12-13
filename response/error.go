package response

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	invalidJsonBodyMsg       = "invalid json body"
	invalidValueFieldJsonMsg = "invalid value for json field"
	internalServerErrorMsg   = "internal service error"
)

type responseError struct {
	Code        int      `json:"code,omitempty"`
	Source      string   `json:"source,omitempty"`
	Message     string   `json:"message,omitempty"`
	Details     string   `json:"details,omitempty"`
	Destination string   `json:"destination,omitempty"`
	Fields      []*Field `json:"fields,omitempty"`
}

type Field struct {
	Field    string `json:"field,omitempty"`
	Message  string `json:"message,omitempty"`
	Location string `json:"location,omitempty"`
}

type responseErrorOptions struct {
	Code        int      `json:"code,omitempty"`
	Source      string   `json:"source,omitempty"`
	Message     string   `json:"message,omitempty"`
	Details     string   `json:"details,omitempty"`
	Destination string   `json:"destination,omitempty"`
	Fields      []*Field `json:"fields,omitempty"`
}

func newResponseError(options *responseErrorOptions) *responseError {
	return &responseError{
		Code:        options.Code,
		Source:      options.Source,
		Message:     options.Message,
		Details:     options.Details,
		Destination: options.Destination,
		Fields:      options.Fields,
	}
}

func jsonError(err error) (*responseError, bool) {
	switch t := err.(type) {
	case *json.SyntaxError:
		return &responseError{
			Message: invalidJsonBodyMsg,
		}, true

	case *json.UnmarshalTypeError:
		return &responseError{
			Message: invalidValueFieldJsonMsg,
			Details: fmt.Sprintf("unexpected value for field '%v'", t.Field),
		}, true
	}

	return nil, false
}

func newValidationErrorFields(err string) []*Field {
	var (
		res  = strings.Split(err, "\",\"")
		errs []*Field
	)

	for _, re := range res {
		re = strings.ReplaceAll(re, `"`, "")
		msg := strings.Split(strings.TrimSpace(re), ":")

		if len(msg) == 2 {
			field, location := getFieldNameAndLocation(msg[0])

			errs = append(errs, &Field{
				Field:    field,
				Message:  strings.TrimSuffix(strings.TrimSpace(msg[1]), "."),
				Location: location,
			})
		}

		if len(msg) > 2 {
			field, location := getFieldNameAndLocation(msg[0])
			errs = append(errs, &Field{
				Field:    field,
				Message:  strings.TrimSuffix(strings.TrimSpace(formatMessage(msg[0], err)), "."),
				Location: location,
			})
		}
	}

	return errs
}

func formatMessage(s string, v string) string {
	return strings.ReplaceAll(v, s+":", "")
}

func getFieldNameAndLocation(data string) (string, string) {
	res := strings.Split(strings.TrimSpace(data), "@")
	if len(res) == 1 {
		return data, "body"
	}

	return strings.TrimSpace(res[0]), strings.TrimSpace(res[1])
}
