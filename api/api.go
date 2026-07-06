package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/shadyendless/chirpy/internal/auth"
	"github.com/shadyendless/chirpy/internal/config"
	"github.com/shadyendless/chirpy/internal/database"
	"github.com/shadyendless/chirpy/internal/utils"
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

		jwt, err := auth.MakeJWT(user.ID, cfg.JWTSecret)
		if err != nil {
			utils.RespondWithServerError(res, err)
			return
		}
		token := auth.MakeRefreshToken()

		refreshToken, err := cfg.DbQueries.CreateRefreshToken(req.Context(), database.CreateRefreshTokenParams{
			Token:     token,
			UserID:    user.ID,
			ExpiresAt: time.Now().Add(time.Hour * 24 * 60), // 60 days
		})

		resJson := struct {
			ID           string `json:"id"`
			CreatedAt    string `json:"created_at"`
			UpdatedAt    string `json:"updated_at"`
			Email        string `json:"email"`
			Token        string `json:"token"`
			RefreshToken string `json:"refresh_token"`
			IsChirpyRed  bool   `json:"is_chirpy_red"`
		}{
			ID:           user.ID.String(),
			CreatedAt:    user.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:    user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			Email:        user.Email,
			Token:        jwt,
			RefreshToken: refreshToken.Token,
			IsChirpyRed:  user.IsChirpyRed,
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
			Body string `json:"body"`
		}

		decoder := json.NewDecoder(req.Body)
		payload := reqBody{}

		if err := decoder.Decode(&payload); err != nil {
			utils.RespondWithServerError(res, err)
			return
		}

		token, err := auth.GetBearerToken(req.Header)
		if err != nil {
			utils.RespondWithErrorStatus(res, err, http.StatusUnauthorized)
			return
		}

		id, err := auth.ValidateJWT(token, cfg.JWTSecret)
		if err != nil {
			utils.RespondWithErrorStatus(res, err, http.StatusUnauthorized)
			return
		}

		if len(payload.Body) > 140 {
			utils.RespondWithErrorStatus(res, errors.New("Chirp is too long"), http.StatusBadRequest)
			return
		}

		chirp, err := cfg.DbQueries.CreateChirp(req.Context(), database.CreateChirpParams{
			Body:   payload.Body,
			UserID: id,
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
		authorId := req.URL.Query().Get("author_id")
		sort := req.URL.Query().Get("sort")

		chirps, err := cfg.DbQueries.GetChirps(req.Context(), database.GetChirpsParams{
			AuthorID: authorId,
			SortDir:  sort,
		})
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

func GetChirpHandler(cfg *config.Config) func(http.ResponseWriter, *http.Request) {
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

func RefreshHandler(cfg *config.Config) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		token, err := auth.GetBearerToken(req.Header)
		if err != nil {
			utils.RespondWithErrorStatus(res, err, http.StatusUnauthorized)
			return
		}

		refreshToken, err := cfg.DbQueries.GetRefreshToken(req.Context(), token)
		if err != nil {
			utils.RespondWithErrorStatus(res, nil, http.StatusUnauthorized)
			return
		}

		jwt, err := auth.MakeJWT(refreshToken.UserID, cfg.JWTSecret)
		if err != nil {
			utils.RespondWithServerError(res, err)
			return
		}

		resJson := struct {
			Token string `json:"token"`
		}{
			Token: jwt,
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

func RevokeHandler(cfg *config.Config) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		token, err := auth.GetBearerToken(req.Header)
		if err != nil {
			utils.RespondWithErrorStatus(res, err, http.StatusUnauthorized)
			return
		}

		err = cfg.DbQueries.RevokeRefreshToken(req.Context(), token)
		if err != nil {
			utils.RespondWithErrorStatus(res, err, http.StatusUnauthorized)
			return
		}

		res.WriteHeader(http.StatusNoContent)
	}
}

func UpdateUserHandler(cfg *config.Config) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		token, err := auth.GetBearerToken(req.Header)
		if err != nil {
			utils.RespondWithErrorStatus(res, err, http.StatusUnauthorized)
			return
		}

		uuid, err := auth.ValidateJWT(token, cfg.JWTSecret)
		if err != nil {
			utils.RespondWithErrorStatus(res, err, http.StatusUnauthorized)
			return
		}

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

		_, err = cfg.DbQueries.GetUser(req.Context(), uuid)
		if err != nil {
			utils.RespondWithErrorStatus(res, err, http.StatusUnauthorized)
			return
		}

		if payload.Password == "" {
			utils.RespondWithErrorStatus(res, errors.New("You must provide a password"), http.StatusBadRequest)
			return
		}

		hashedPassword, err := auth.HashPassword(payload.Password)
		if err != nil {
			utils.RespondWithServerError(res, err)
			return
		}

		updatedUser, err := cfg.DbQueries.UpdateUser(req.Context(), database.UpdateUserParams{
			ID:             uuid,
			Email:          payload.Email,
			HashedPassword: hashedPassword,
		})
		if err != nil {
			utils.RespondWithServerError(res, err)
			return
		}

		resJson, err := json.Marshal(updatedUser)
		if err != nil {
			utils.RespondWithServerError(res, err)
			return
		}

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		res.Write(resJson)
	}
}

func DeleteChirpHandler(cfg *config.Config) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		chirpID, err := uuid.Parse(req.PathValue("chirpID"))
		if err != nil {
			utils.RespondWithServerError(res, err)
			return
		}

		token, err := auth.GetBearerToken(req.Header)
		if err != nil {
			utils.RespondWithErrorStatus(res, err, http.StatusUnauthorized)
			return
		}

		userID, err := auth.ValidateJWT(token, cfg.JWTSecret)
		if err != nil {
			utils.RespondWithErrorStatus(res, err, http.StatusUnauthorized)
			return
		}

		chirp, err := cfg.DbQueries.GetChirp(req.Context(), chirpID)
		if err != nil {
			utils.RespondWithErrorStatus(res, err, http.StatusNotFound)
			return
		}

		if chirp.UserID != userID {
			utils.RespondWithErrorStatus(res, nil, http.StatusForbidden)
			return
		}

		if err = cfg.DbQueries.DeleteChirp(req.Context(), chirpID); err != nil {
			utils.RespondWithServerError(res, err)
			return
		}

		res.WriteHeader(204)
	}
}

func PolkaWebhookHandler(cfg *config.Config) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		if apiKey, err := auth.GetAPIKey(req.Header); err != nil || apiKey != os.Getenv("POLKA_KEY") {
			utils.RespondWithErrorStatus(res, err, http.StatusUnauthorized)
			return
		}

		body := struct {
			Event string `json:"event"`
			Data  struct {
				UserID string `json:"user_id"`
			} `json:"data"`
		}{}

		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&body); err != nil {
			utils.RespondWithServerError(res, err)
			return
		}

		if body.Event != "user.upgraded" {
			utils.RespondWithErrorStatus(res, nil, http.StatusNoContent)
			return
		}

		if _, err := cfg.DbQueries.UpdateUserSubscription(req.Context(), database.UpdateUserSubscriptionParams{
			IsChirpyRed: true,
			ID:          uuid.MustParse(body.Data.UserID),
		}); err != nil {
			utils.RespondWithErrorStatus(res, err, http.StatusNotFound)
			return
		}

		res.WriteHeader(http.StatusNoContent)
	}
}
