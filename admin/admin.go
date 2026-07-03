package admin

import (
	"fmt"
	"net/http"

	"github.com/shadyendless/chirpy/config"
	"github.com/shadyendless/chirpy/utils"
)

func GetMetricsHandler(cfg *config.Config) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, _ *http.Request) {
		res.Header().Add("Content-Type", "text/html")
		res.WriteHeader(http.StatusOK)
		res.Write([]byte(fmt.Sprintf(`<html>
		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>
	</html>`, cfg.FileserverHits.Load())))
	}
}

func ResetUsersHandler(cfg *config.Config) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		if cfg.Platform != "dev" {
			res.WriteHeader(http.StatusForbidden)
			return
		}

		err := cfg.DbQueries.DeleteUsers(req.Context())
		if err != nil {
			utils.RespondWithServerError(res, err)
			return
		}

		res.WriteHeader(http.StatusOK)
	}
}
