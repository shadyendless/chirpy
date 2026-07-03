package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"

	"github.com/google/uuid"
	"github.com/shadyendless/chirpy/config"
	"github.com/shadyendless/chirpy/internal/database"
	"github.com/shadyendless/chirpy/utils"
)

func HealthHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusOK)
	res.Write([]byte("OK"))
}

func ValidateChirpHandler(res http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		utils.RespondWithServerError(res, err)
		return
	}

	if len(params.Body) > 140 {
		utils.RespondWithErrorStatus(res, errors.New("Chirp is too long"), http.StatusBadRequest)
		return
	}

	profanityRegex := regexp.MustCompile("(?i)kerfuffle|sharbert|fornax")

	val := struct {
		CleanedBody string `json:"cleaned_body"`
	}{
		CleanedBody: profanityRegex.ReplaceAllString(params.Body, "****"),
	}
	jsonVal, err := json.Marshal(val)
	if err != nil {
		utils.RespondWithServerError(res, err)
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write(jsonVal)
}

func CreateUserHandler(cfg *config.Config) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		type reqBody struct {
			Email string `json:"email"`
		}

		decoder := json.NewDecoder(req.Body)
		payload := reqBody{}

		if err := decoder.Decode(&payload); err != nil {
			utils.RespondWithServerError(res, err)
			return
		}

		user, err := cfg.DbQueries.CreateUser(req.Context(), payload.Email)
		if err != nil {
			utils.RespondWithServerError(res, err)
			return
		}

		resJson, err := json.Marshal(user)
		if err != nil {
			utils.RespondWithServerError(res, err)
			return
		}

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated)
		res.Write(resJson)
	}
}

func CreateChirpHandler(cfg *config.Config) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		type reqBody struct {
			Body   string    `json:"body"`
			UserID uuid.UUID `json:"user_id"`
		}

		decoder := json.NewDecoder(req.Body)
		payload := reqBody{}

		if err := decoder.Decode(&payload); err != nil {
			utils.RespondWithServerError(res, err)
			return
		}

		if len(payload.Body) > 140 {
			utils.RespondWithErrorStatus(res, errors.New("Chirp is too long"), http.StatusBadRequest)
			return
		}

		chirp, err := cfg.DbQueries.CreateChirp(req.Context(), database.CreateChirpParams{
			Body:   payload.Body,
			UserID: payload.UserID,
		})
		if err != nil {
			utils.RespondWithServerError(res, err)
			return
		}

		resJson, err := json.Marshal(chirp)
		if err != nil {
			utils.RespondWithServerError(res, err)
			return
		}

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated)
		res.Write(resJson)
	}
}
