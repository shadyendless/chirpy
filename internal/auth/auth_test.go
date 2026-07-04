package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestValidateJWT(t *testing.T) {
	id := uuid.New()
	jwt, err := MakeJWT(id, "secret", time.Hour)
	if err != nil {
		t.Error(err)
	}

	t.Logf("\nGenerated JWT: %v\n", jwt)

	userID, err := ValidateJWT(jwt, "secret")
	if err != nil {
		t.Error(err)
	}

	if userID != id {
		t.Logf("Decoded ID '%v' does not match '%v'", userID, id)
		t.Fail()
	}

	t.Logf("JWT successfully generates and is validated.\n")
}

func TestGetBearerToken(t *testing.T) {
	t.Run("No header returns error", func(t *testing.T) {
		_, err := GetBearerToken(http.Header{})
		if err == nil || err.Error() != "No Authorization header received" {
			t.Fail()
		}
	})

	t.Run("Malformed header returns error", func(t *testing.T) {
		_, err := GetBearerToken(http.Header{
			"Authorization": []string{" Bearer: Applesauce "},
		})

		if err == nil || err.Error() != "Authorization header must be in the format: Bearer <token>" {
			t.Fail()
		}
	})

	t.Run("Returns authorization token", func(t *testing.T) {
		authToken := "TestAuthToken"

		bearer, err := GetBearerToken(http.Header{
			"Authorization": []string{"Bearer TestAuthToken"},
		})

		if err != nil {
			t.Fail()
		}

		if authToken != bearer {
			t.Fail()
		}
	})
}
