package apierr

import (
	"net/http"
	"strconv"
)

type HttpStatus int

func (s HttpStatus) Value() int {
	return int(s)
}
func (s HttpStatus) String() string {
	return strconv.Itoa(s.Value())
}

const (
	StatusCreated             HttpStatus = http.StatusCreated
	StatusDeleted             HttpStatus = http.StatusNoContent
	StatusBadRequest          HttpStatus = http.StatusBadRequest
	StatusUnauthorized        HttpStatus = http.StatusUnauthorized
	StatusForbidden           HttpStatus = http.StatusForbidden
	StatusNotFound            HttpStatus = http.StatusNotFound
	StatusConflict            HttpStatus = http.StatusConflict
	StatusLocked              HttpStatus = http.StatusLocked
	StatusInternalServerError HttpStatus = http.StatusInternalServerError
)
