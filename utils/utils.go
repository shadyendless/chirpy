package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

func RespondWithError(res http.ResponseWriter, err error, statusCode int) {
	type returnVal struct {
		Error string `json:"error"`
	}

	if statusCode == 0 {
		statusCode = 200
	}

	val := returnVal{
		Error: err.Error(),
	}

	data, err := json.Marshal(val)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		res.WriteHeader(500)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(statusCode)
	res.Write(data)
}
