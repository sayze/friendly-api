package http

import (
	"net/http"

	"github.com/go-chi/render"
	validator2 "github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

// ErrResponse renderer type for handling all sorts of errors.
type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

// SuccessResponse renderer type for handling successful responses.
type SuccessResponse struct {
	HTTPStatusCode int         `json:"-"`
	StatusText     string      `json:"status"`
	Data           interface{} `json:"data,omitempty"`
}

// Render renders an ErrorResponse struct.
func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

// Render renders an SuccessResponse struct.
func (s *SuccessResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, s.HTTPStatusCode)
	return nil
}

// SuccessDataRequest returns a data response struct.
func SuccessDataRequest(data interface{}) render.Renderer {
	return &SuccessResponse{
		HTTPStatusCode: 200,
		StatusText:     "ok",
		Data:           data,
	}
}

// SuccessNoContentRequest returns a no content response struct.
func SuccessNoContentRequest(message string) render.Renderer {
	return &SuccessResponse{
		HTTPStatusCode: 404,
		StatusText:     "resource not found",
		Data:           message,
	}
}

// ErrInvalidRequest returns an invalid request error.
func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

// ErrFatalRequest returns an internal error response.
func ErrFatalRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 500,
		StatusText:     "Internal error.",
		ErrorText:      err.Error(),
	}
}

// ErrValidationFailed returns a response containing all the failed validation rules.
func ErrValidationFailed(err error) render.Renderer {
	for _, err := range err.(validator2.ValidationErrors) {

		logrus.Println(err.Namespace())
		logrus.Println(err.Field())
		logrus.Println(err.StructNamespace()) // can differ when a custom TagNameFunc is registered or
		logrus.Println(err.StructField())     // by passing alt name to ReportError like below
		logrus.Println(err.Tag())
		logrus.Println(err.ActualTag())
		logrus.Println(err.Kind())
		logrus.Println(err.Type())
		logrus.Println(err.Value())
		logrus.Println(err.Param())
		logrus.Println()
	}

	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}
