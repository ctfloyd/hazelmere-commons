package hz_handler

import (
	"fmt"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_api"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_service_error"
	jsoniter "github.com/json-iterator/go"
	"io"
	"net/http"
	"time"
)

const RegexUuid string = `[0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{12}`

type ErrorResponse interface {
	GetStatus() int
}

type ConvertFn func(serviceError hz_service_error.ServiceError, message string) ErrorResponse

func ErrorWithConvertFn(w http.ResponseWriter, serviceError hz_service_error.ServiceError, message string, convertFn ConvertFn) {
	response := convertFn(serviceError, message)
	b, err := jsoniter.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(response.GetStatus())
	_, _ = w.Write(b)
}

func Error(w http.ResponseWriter, serviceError hz_service_error.ServiceError, message string) {
	ErrorWithConvertFn(w, serviceError, message, convertServiceErrorToResponse)
}

func ErrorArgs(w http.ResponseWriter, serviceError hz_service_error.ServiceError, message string, args ...any) {
	ErrorArgsWithConvertFn(w, serviceError, convertServiceErrorToResponse, message, args)
}

func ErrorArgsWithConvertFn(w http.ResponseWriter, serviceError hz_service_error.ServiceError, convertFn ConvertFn, message string, args ...any) {
	ErrorWithConvertFn(w, serviceError, fmt.Sprintf(message, args...), convertFn)
}

func Json(w http.ResponseWriter, status int, response any) {
	if response != nil {
		b, err := jsoniter.Marshal(response)
		if err != nil {
			Error(w, hz_service_error.Internal, "Could not marshal response.")
			return
		}
		w.Header().Add("Content-Type", "application/json")
		_, err = w.Write(b)
		if err != nil {
			Error(w, hz_service_error.Internal, "Could not write all bytes in the response.")
		}
	}
	w.WriteHeader(status)
}

func Ok(w http.ResponseWriter, response any) {
	Json(w, http.StatusOK, response)
}

func ReadBody(w http.ResponseWriter, r *http.Request, body any) bool {
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		Error(w, hz_service_error.Internal, "An unexpected error occurred while reading request body.")
		return false
	}

	err = jsoniter.Unmarshal(bytes, body)
	if err != nil {
		Error(w, hz_service_error.BadRequest, fmt.Sprintf("The request body could not be parsed. %v", err))
		return false
	}

	return true
}

func convertServiceErrorToResponse(serviceError hz_service_error.ServiceError, message string) ErrorResponse {
	return hz_api.ErrorResponse{
		Code:      serviceError.Code,
		Status:    serviceError.Status,
		Message:   message,
		Timestamp: time.Now(),
	}
}
