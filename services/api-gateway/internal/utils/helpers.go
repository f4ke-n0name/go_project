package utils

import (
	"encoding/json"
	"net/http"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func WriteJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, code int, msg string) {
	WriteJSON(w, code, ErrorResponse{Error: msg})
}

func WriteGRPCError(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}

	switch st.Code() {
	case codes.NotFound:
		WriteError(w, http.StatusNotFound, st.Message())
	case codes.AlreadyExists:
		WriteError(w, http.StatusConflict, st.Message())
	case codes.InvalidArgument:
		WriteError(w, http.StatusBadRequest, st.Message())
	case codes.Unauthenticated:
		WriteError(w, http.StatusUnauthorized, st.Message())
	case codes.PermissionDenied:
		WriteError(w, http.StatusForbidden, st.Message())
	default:
		WriteError(w, http.StatusInternalServerError, "internal error")
	}
}

func ParseInt(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}
