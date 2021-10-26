package errs

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HttpError struct {
	Code int
	Err  error
	Type gin.ErrorType
}

func (e HttpError) Error() string {
	return e.Err.Error()
}

func (e HttpError) Abort(c *gin.Context) {
	c.AbortWithError(e.Code, e.Err).SetType(e.Type)
}

func NewHttpError(code int, err error, t gin.ErrorType) *HttpError {
	return &HttpError{Code: code, Err: err, Type: t}
}

func NewNotFoundError(err error) *HttpError {
	return NewHttpError(http.StatusNotFound, err, gin.ErrorTypePublic)
}

func NewBadRequestError(err error) *HttpError {
	return NewHttpError(http.StatusBadRequest, err, gin.ErrorTypePublic)
}

func NewUnauthorizedError(err error) *HttpError {
	return NewHttpError(http.StatusUnauthorized, err, gin.ErrorTypePublic)
}

func NewForbiddenError(err error) *HttpError {
	return NewHttpError(http.StatusForbidden, err, gin.ErrorTypePublic)
}

func NewInternalError(err error) *HttpError {
	return NewHttpError(http.StatusInternalServerError, err, gin.ErrorTypePrivate)
}
