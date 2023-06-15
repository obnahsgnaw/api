package jwt

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/obnahsgnaw/application/pkg/utils"
	"strconv"
	"time"
)

var SignMethod = jwt.SigningMethodHS256

type Userinfo struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Attrs map[string]string
}
type ZyClaims struct {
	jwt.RegisteredClaims
	Userinfo Userinfo `json:"userinfo"`
}

func (c ZyClaims) Valid() error {
	now := jwt.TimeFunc()

	// The claims below are optional, by default, so if they are set to the
	// default value in Go, let's not fail the verification for them.
	if !c.VerifyExpiresAt(now, true) {
		return jwt.ErrTokenExpired
	}

	if !c.VerifyIssuedAt(now, true) {
		return jwt.ErrTokenUsedBeforeIssued
	}

	if !c.VerifyNotBefore(now, true) {
		return jwt.ErrTokenNotValidYet
	}

	return nil
}

func GenerateToken(subject string, key []byte, issuer string, userinfo Userinfo, notBefore *jwt.NumericDate, ttl time.Duration) (string, error) {
	now := time.Now()
	if notBefore == nil {
		notBefore = jwt.NewNumericDate(now)
	}
	c := ZyClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   subject,
			Audience:  []string{userinfo.Id},
			ExpiresAt: jwt.NewNumericDate(notBefore.Add(ttl)),
			NotBefore: notBefore,
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        strconv.FormatInt(now.UnixMicro(), 10),
		},
		Userinfo: userinfo,
	}
	token := jwt.NewWithClaims(SignMethod, c)

	return token.SignedString(key)
}

func ValidateToken(subject, issuer, tokenString string, keyProvider func(claims *ZyClaims) ([]byte, error)) (Userinfo, error) {
	var claims *ZyClaims
	var ok bool
	token, err := jwt.ParseWithClaims(tokenString, &ZyClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok = token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		if claims, ok = token.Claims.(*ZyClaims); ok {
			if key, err := keyProvider(claims); err == nil {
				return key, nil
			} else {
				return nil, err
			}
		}

		return nil, errors.New("unexpected payload")
	})

	if err != nil {
		return Userinfo{}, err
	}

	if token.Valid {
		if !claims.RegisteredClaims.VerifyIssuer(issuer, true) {
			return Userinfo{}, errors.New("token issuer error")
		}
		if !claims.RegisteredClaims.VerifyAudience(claims.Userinfo.Id, true) {
			return Userinfo{}, jwt.ErrTokenInvalidAudience
		}
		if !verifySub(claims.RegisteredClaims.Subject, subject, true) {
			return Userinfo{}, errors.New("token not for this application")
		}
		return claims.Userinfo, nil
	} else {
		return Userinfo{}, err
	}
}

func verifySub(sub string, cmp string, required bool) bool {
	if sub == "" {
		return !required
	}
	if subtle.ConstantTimeCompare([]byte(sub), []byte(cmp)) != 0 {
		return true
	} else {
		return false
	}
}

func GenKey() []byte {
	return []byte(utils.RandAlphaNum(SignMethod.Hash.Size()))
}
