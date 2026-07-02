package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/shadyendless/chirpy/utils"
)

func HealthHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(200)
	res.Write([]byte("OK"))
}

func ValidateChirpHandler(res http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		utils.RespondWithError(res, errors.New("Something went wrong"), 500)
		return
	}

	if len(params.Body) > 140 {
		utils.RespondWithError(res, errors.New("Chirp is too long"), 400)
		return
	}

	val := struct {
		Valid bool `json:"valid"`
	}{
		Valid: true,
	}
	jsonVal, err := json.Marshal(val)
	if err != nil {
		utils.RespondWithError(res, errors.New("Something went wrong"), 500)
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(200)
	res.Write(jsonVal)
}
