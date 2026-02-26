package templates

import (
	"log"

	"httpfromtcp/internal/response"
)

type ErrResponse struct {
	StatusCode response.StatusCode
	Err        error
}

func Panic(status response.StatusCode, err error) {
	panic(ErrResponse{
		StatusCode: status,
		Err:        err,
	})
}

func Recover(w *response.Writer) {
	if r := recover(); r != nil {
		switch v := r.(type) {
		case ErrResponse:
			log.Println(v.Err)
			handlerErr := response.NewHandlerError(v.StatusCode, v.Err.Error())
			_ = w.WriteError(*handlerErr)
		default:
			log.Printf("unexpected panic: %v", r)
			handlerErr := response.NewHandlerError(response.InternalServerError, "Internal Server Error")
			_ = w.WriteError(*handlerErr)
		}
	}
}

func ErrorOnlyMust(err error) {
	if err != nil {
		Panic(response.InternalServerError, err)
	}
}

func Must[T any](val T, err error) T {
	if err != nil {
		Panic(response.InternalServerError, err)
	}
	return val
}

func MustWithStatus[T any](statusCode response.StatusCode, val T, err error) T {
	if err != nil {
		Panic(statusCode, err)
	}
	return val
}
