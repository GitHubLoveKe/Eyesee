package response

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type PageData struct {
	Items    interface{} `json:"items"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

func JSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func OK(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, Response{Code: 0, Message: "success", Data: data})
}

func OKPage(w http.ResponseWriter, items interface{}, total int64, page, pageSize int) {
	OK(w, PageData{Items: items, Total: total, Page: page, PageSize: pageSize})
}

func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, Response{Code: 0, Message: "created", Data: data})
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func Error(w http.ResponseWriter, httpStatus int, code int, message string) {
	JSON(w, httpStatus, Response{Code: code, Message: message})
}

func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, 40001, message)
}

func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, 40002, message)
}

func Conflict(w http.ResponseWriter, message string) {
	Error(w, http.StatusConflict, 40901, message)
}

func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, 40101, message)
}

func Forbidden(w http.ResponseWriter, message string) {
	Error(w, http.StatusForbidden, 40301, message)
}

func InternalError(w http.ResponseWriter, message string) {
	Error(w, http.StatusInternalServerError, 50001, message)
}

func Unprocessable(w http.ResponseWriter, code int, message string) {
	Error(w, http.StatusUnprocessableEntity, code, message)
}
