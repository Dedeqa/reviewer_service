package handlers

import (
	"encoding/json"
	"net/http"
)

// Error codes used by API (as strings from spec)
const (
	ErrCodeTeamExists  = "TEAM_EXISTS"
	ErrCodePRExists    = "PR_EXISTS"
	ErrCodePRMerged    = "PR_MERGED"
	ErrCodeNotAssigned = "NOT_ASSIGNED"
	ErrCodeNoCandidate = "NO_CANDIDATE"
	ErrCodeNotFound    = "NOT_FOUND"
)

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
type errorResponse struct {
	Error apiError `json:"error"`
}

func writeErr(w http.ResponseWriter, code int, errCode, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(errorResponse{Error: apiError{Code: errCode, Message: msg}})
}
