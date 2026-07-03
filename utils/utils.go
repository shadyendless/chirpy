package utils

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

func RespondWithError(res http.ResponseWriter, err error) {
	RespondWithErrorStatus(res, err, http.StatusInternalServerError)
}

func RespondWithServerError(res http.ResponseWriter, err error) {
	log.Println(err)
	RespondWithErrorStatus(res, errors.New("Something went wrong"), http.StatusInternalServerError)
}

func RespondWithErrorStatus(res http.ResponseWriter, err error, statusCode int) {
	if err == nil {
		res.WriteHeader(statusCode)
		return
	}

	type returnVal struct {
		Error string `json:"error"`
	}

	val := returnVal{
		Error: err.Error(),
	}

	data, err := json.Marshal(val)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(statusCode)
	res.Write(data)
}
