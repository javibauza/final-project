package transport

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kit/log"
	"github.com/gorilla/mux"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/javibauza/final-project/gbfp-httpservice/endpoints"
	erro "github.com/javibauza/final-project/gbfp-httpservice/errors"
)

type ErrorResponse struct {
	Message string
}

func NewHTTPServer(endpoints endpoints.Endpoints, logger log.Logger) http.Handler {
	r := mux.NewRouter()

	r.Use(commonMiddleware)
	options := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(encodeError),
	}

	r.Methods("POST").Path("/api/auth").Handler(
		httptransport.NewServer(
			endpoints.Authenticate,
			decodeAuthRequest,
			encodeAuthResponse,
			options...,
		),
	)

	r.Methods("POST").Path("/api").Handler(
		httptransport.NewServer(
			endpoints.CreateUser,
			decodeCreateUserRequest,
			encodeCreateUserResponse,
			options...,
		),
	)

	r.Methods("PUT").Path("/api/{userId}").Handler(
		httptransport.NewServer(
			endpoints.UpdateUser,
			decodeUpdateUserRequest,
			encodeUpdateUserResponse,
			options...,
		),
	)

	r.Methods("GET").Path("/api/{userId}").Handler(
		httptransport.NewServer(
			endpoints.GetUser,
			decodeGetUserRequest,
			encodeGetUserResponse,
			options...,
		),
	)

	return r
}

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func decodeAuthRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req endpoints.AuthRequest
	if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
		return nil, e
	}
	return req, nil
}

func encodeAuthResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(response)
}

func decodeCreateUserRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req endpoints.CreateUserRequest
	if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
		return nil, e
	}
	return req, nil
}

func encodeCreateUserResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(response)
}

func decodeUpdateUserRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req endpoints.UpdateUserRequest
	if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
		return nil, e
	}

	params := mux.Vars(r)
	req.UserId = params["userId"]

	return req, nil
}

func encodeUpdateUserResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	return nil
}

func decodeGetUserRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req endpoints.GetUserRequest

	params := mux.Vars(r)
	req.UserId = params["userId"]

	return req, nil
}

func encodeGetUserResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(response)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	if err == nil {
		panic("encodeError with nil error")
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(codeFrom(err))
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func codeFrom(err error) int {
	switch err.(type) {
	case erro.ErrNotFound:
		return http.StatusNotFound
	case erro.ErrBadRequest:
		return http.StatusBadRequest
	case erro.ErrForbidden:
		return http.StatusForbidden
	case erro.ErrInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
