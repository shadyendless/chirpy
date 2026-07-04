package auth

import (
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
