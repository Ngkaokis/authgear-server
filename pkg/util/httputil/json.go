package httputil

import (
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

const BodyMaxSize = 1024 * 1024 * 10

func IsJSONContentType(contentType string) bool {
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}
	if mediaType != "application/json" {
		return false
	}
	// No params is good
	if len(params) == 0 {
		return true
	}
	// Contains unknown params
	if len(params) > 1 {
		return false
	}
	// The sole param must be charset=utf-8
	charset := params["charset"]
	return strings.ToLower(charset) == "utf-8"
}

func ParseJSONBody(r *http.Request, w http.ResponseWriter, parse func(io.Reader, interface{}) error, payload interface{}) error {
	if !IsJSONContentType(r.Header.Get("Content-Type")) {
		return apierrors.NewBadRequest("request content type is invalid")
	}
	body := http.MaxBytesReader(w, r.Body, BodyMaxSize)
	defer body.Close()
	return parse(body, payload)
}

type BodyDefaulter interface {
	SetDefaults()
}

func BindJSONBody(r *http.Request, w http.ResponseWriter, v *validation.SchemaValidator, payload interface{}) error {
	const errorMessage = "invalid request body"
	return ParseJSONBody(r, w, func(reader io.Reader, value interface{}) error {
		err := v.ParseWithMessage(reader, errorMessage, value)
		if err != nil {
			if !apierrors.IsKind(err, apierrors.ValidationFailed) {
				return apierrors.NewBadRequest(errorMessage)
			}
			return err
		}

		if value, ok := value.(BodyDefaulter); ok {
			value.SetDefaults()
		}
		return validation.ValidateValueWithMessage(value, errorMessage)
	}, payload)
}

type JSONResponseWriterLogger struct{ *log.Logger }

func NewJSONResponseWriterLogger(lf *log.Factory) JSONResponseWriterLogger {
	return JSONResponseWriterLogger{lf.New("json-response-writer")}
}

type JSONResponseWriter struct {
	Logger JSONResponseWriterLogger
}

func (w *JSONResponseWriter) WriteResponse(rw http.ResponseWriter, resp *api.Response) {
	httpStatus := http.StatusOK
	encoder := json.NewEncoder(rw)
	err := apierrors.AsAPIError(resp.Error)

	if err != nil {
		httpStatus = err.Code
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(httpStatus)
	if err := encoder.Encode(resp); err != nil {
		panic(err)
	}

	if err != nil && err.Code >= 500 && err.Code < 600 {
		w.Logger.WithError(resp.Error).Error("unexpected error occurred")
	}
}
