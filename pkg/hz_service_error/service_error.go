package hz_service_error

import (
	"net/http"
)

type ServiceError struct {
	Code   string
	Status int
}

var Internal = ServiceError{Code: "INTERNAL_SERVICE_ERROR", Status: http.StatusInternalServerError}
var BadRequest = ServiceError{Code: "BAD_REQUEST", Status: http.StatusBadRequest}
