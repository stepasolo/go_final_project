package api

import (
	"net/http"
	"time"
)

func ValidateDate(w http.ResponseWriter, dateStr string) bool {
	_, err := time.Parse("20060102", dateStr)
	if err != nil {
		http.Error(w, `{"error": "Неверный формат даты"}`, http.StatusBadRequest)
		return false
	}
	return true
}

func validateTitle(w http.ResponseWriter, title string) bool {
	if title == "" {
		http.Error(w, `{"error": "Поле Title обязательно"}`, http.StatusBadRequest)
		return false
	}
	return true
}

func validateIdParam(w http.ResponseWriter, id string) bool {
	if id == "" {
		http.Error(w, `{"error": "Не указан идентификатор"}`, http.StatusBadRequest)
		return false
	}
	return true
}
