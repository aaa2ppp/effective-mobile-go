package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"effective-mobile-go/internal/logger"
	"effective-mobile-go/internal/model"
)

const loggerGroup = "handler"

type helper struct {
	op  string
	w   http.ResponseWriter
	r   *http.Request
	log *slog.Logger
}

func newHelper(op string, w http.ResponseWriter, r *http.Request) helper {
	return helper{
		op: op,
		w:  w,
		r:  r,
	}
}

func (x *helper) Log() *slog.Logger {

	if x.log == nil {
		x.log = logger.GetLoggerFromContextOrDefault(x.r.Context()).
			WithGroup(loggerGroup).
			With(slog.String("op", x.op))
	}

	return x.log
}

func (x helper) Ctx() context.Context {
	return x.r.Context()
}

type errorResponse struct {
	Error httpError `json:"error,omitempty"`
}

type httpError struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func (x helper) WriteError(err error) {

	var httpErr *model.Error
	var resp errorResponse

	if errors.As(err, &httpErr) {

		resp = errorResponse{
			Error: httpError{
				Code:    httpErr.Code(),
				Message: httpErr.Error(),
			},
		}
	} else {

		x.Log().Error("unhandled error has been detected", "error", err, "errorType", fmt.Sprintf("%T", err), "errorDetail", fmt.Sprintf("%+v", err))
		resp = errorResponse{
			httpError{
				Code:    http.StatusInternalServerError,
				Message: http.StatusText(http.StatusInternalServerError),
			},
		}
	}

	x.w.WriteHeader(resp.Error.Code)
	x.WriteResponse(&resp)
}

func (x helper) WriteResponse(resp any) {

	x.w.Header().Add("content-type", "application/json")

	if err := json.NewEncoder(x.w).Encode(&resp); err != nil {

		x.Log().Error("can't write response", "error", err)
	}
}

func (x helper) GetID() (uint64, error) {

	s := x.r.PathValue("id")
	if s == "" {

		x.Log().Error("no ID in the path", "path", x.r.URL.Path)
		return 0, ErrInternalError
	}

	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {

		x.Log().Debug("can't parse ID", "error", err, "path", x.r.URL.Path)
		return 0, ErrBadRequest
	}

	return v, nil
}

func (x helper) DecodeBody(req any) error {

	body, err := io.ReadAll(x.r.Body)
	if err != nil {

		x.Log().Error("can't read request body", "error", err)
		return ErrInternalError
	}

	if err := json.Unmarshal(body, &req); err != nil {

		x.Log().Debug("can't parse body", "error", err, "body", body)
		return ErrBadRequest
	}

	return nil
}
