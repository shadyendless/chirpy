package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"

	"github.com/google/uuid"
	"github.com/shadyendless/chirpy/config"
	"github.com/shadyendless/chirpy/internal/auth"
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
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		decoder := json.NewDecoder(req.Body)
		payload := reqBody{}

		if err := decoder.Decode(&payload); err != nil {
			utils.RespondWithServerError(res, err)
			return
		}

		if payload.Email == "" || payload.Password == "" {
			utils.RespondWithErrorStatus(res, errors.New("email and password are required"), http.StatusBadRequest)
			return
		}

		hashedPassword, err := auth.HashPassword(payload.Password)
		if err != nil {
			utils.RespondWithServerError(res, err)
			return
		}

		user, err := cfg.DbQueries.CreateUser(req.Context(), database.CreateUserParams{
			Email:          payload.Email,
			HashedPassword: hashedPassword,
		})
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

func LoginHandler(cfg *config.Config) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		type reqBody struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		decoder := json.NewDecoder(req.Body)
		payload := reqBody{}

		if err := decoder.Decode(&payload); err != nil {
			utils.RespondWithServerError(res, err)
			return
		}

		user, err := cfg.DbQueries.GetUserByEmail(req.Context(), payload.Email)
		if err != nil {
			utils.RespondWithErrorStatus(res, nil, http.StatusUnauthorized)
			return
		}

		matches, err := auth.CheckPasswordHash(payload.Password, user.HashedPassword)
		if err != nil {
			utils.RespondWithServerError(res, err)
			return
		}

		if !matches {
			utils.RespondWithErrorStatus(res, nil, http.StatusUnauthorized)
			return
		}

		resJson := struct {
			ID        string `json:"id"`
			CreatedAt string `json:"created_at"`
			UpdatedAt string `json:"updated_at"`
			Email     string `json:"email"`
		}{
			ID:        user.ID.String(),
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			Email:     user.Email,
		}

		resBody, err := json.Marshal(resJson)
		if err != nil {
			utils.RespondWithServerError(res, err)
			return
		}

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(200)
		res.Write(resBody)
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

func GetChirpsHandler(cfg *config.Config) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		chirps, err := cfg.DbQueries.GetChirps(req.Context())
		if err != nil {
			utils.RespondWithServerError(res, err)
			return
		}

		body, err := json.Marshal(chirps)
		if err != nil {
			utils.RespondWithServerError(res, err)
			return
		}

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		res.Write(body)
	}
}

func GetChirp(cfg *config.Config) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		chirpID, err := uuid.Parse(req.PathValue("chirpID"))
		if err != nil {
			utils.RespondWithServerError(res, err)
			return
		}

		chirp, err := cfg.DbQueries.GetChirp(req.Context(), chirpID)
		if err != nil {
			if err == sql.ErrNoRows {
				utils.RespondWithErrorStatus(res, nil, 404)
			} else {
				utils.RespondWithServerError(res, err)
			}

			return
		}

		body, err := json.Marshal(chirp)
		if err != nil {
			utils.RespondWithServerError(res, err)
			return
		}

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		res.Write(body)
	}
}
