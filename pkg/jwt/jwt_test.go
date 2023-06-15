package jwt

import (
	"testing"
	"time"
)

func TestJwtToken(t *testing.T) {
	k := GenKey()
	token, err := GenerateToken("app_1", k, "a", Userinfo{
		Id:   "1",
		Name: "Test",
	}, nil, 5*time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	_, err = ValidateToken("app_1", "a", token, func(claims *ZyClaims) ([]byte, error) {
		return k, nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
