package handler

import (
	"effective-mobile-go/internal/model"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func Test_helper_WriteResponse(t *testing.T) {
	tests := []struct {
		name string
		got  any
		want any
	}{
		{
			"1",
			&struct {
				ID   uint64
				Name string
			}{},
			&struct {
				ID   uint64
				Name string
			}{
				123,
				"Name123",
			},
		},
		// TODO: Add test cases.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "http://localhost/test", nil)
			newHelper("test", w, r).WriteResponse(tt.want)

			// check status code (200 by default)
			if got, want := w.Code, 200; got != want {
				t.Fatalf("status code = %d, want %d", got, want)
			}

			// check headers
			for _, v := range [][2]string{{"content-type", "application/json"}} {
				header, want := v[0], v[1]
				got := w.Header().Get(header)
				if got != want {
					t.Fatalf("%s = %q, want %q", header, got, want)
				}
			}

			// check body
			if err := json.NewDecoder(w.Body).Decode(tt.got); err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(tt.got, tt.want) {
				t.Fatalf("got = %v (%T), want %v (%T)", tt.got, tt.got, tt.want, tt.want)
			}
		})
	}
}

func Test_helper_WriteError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args
		want errorResponse
	}{
		{
			"unknown error",
			args{errors.New("unknown error")},
			errorResponse{
				Error: httpError{
					Code:    http.StatusInternalServerError,
					Message: http.StatusText(http.StatusInternalServerError),
				},
			},
		},
		{
			"model.ErrNotFound",
			args{model.ErrNotFound},
			errorResponse{
				Error: httpError{
					Code:    model.ErrNotFound.Code(),
					Message: model.ErrNotFound.Error(),
				},
			},
		},
		{
			"model.ErrBadRequest",
			args{model.ErrBadRequest},
			errorResponse{
				Error: httpError{
					Code:    model.ErrBadRequest.Code(),
					Message: model.ErrBadRequest.Error(),
				},
			},
		},
		{
			"model.ErrInternalError",
			args{model.ErrInternalError},
			errorResponse{
				Error: httpError{
					Code:    model.ErrInternalError.Code(),
					Message: model.ErrInternalError.Error(),
				},
			},
		},
		// TODO: Add test cases.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "http://localhost/test", nil)
			newHelper("test", w, r).WriteError(tt.args.err)

			// check status code
			if got, want := w.Code, tt.want.Error.Code; got != want {
				t.Fatalf("status code = %d, want %d", got, want)
			}

			// check headers
			for _, v := range [][2]string{{"content-type", "application/json"}} {
				header, want := v[0], v[1]
				got := w.Header().Get(header)
				if got != want {
					t.Fatalf("%s = %q, want %q", header, got, want)
				}
			}

			// check body
			{
				var got errorResponse
				if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
					t.Fatal(err)
				}

				if !reflect.DeepEqual(got, tt.want) {
					t.Fatalf("got = %v (%T), want %v (%T)", got, got, tt.want, tt.want)
				}
			}
		})
	}
}
