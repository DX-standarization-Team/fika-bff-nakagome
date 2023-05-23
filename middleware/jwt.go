package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

// golang では slice, array, mapは定数として使用できない(https://tech.notti.link/02b52f2)
const Audience = "https://fs-apigw-bff-nakagome-bi5axj14.uc.gateway.dev/"

// const Audience2 = "https://dev-kjqwuq76z8suldgw.us.auth0.com/userinfo"
const DomainName = "dev-kjqwuq76z8suldgw.us.auth0.com"

type Jwks struct {
	Keys []JSONWebKeys `json:"keys"`
}

type JSONWebKeys struct {
	Kty string   `json:"kty"`
	Kid string   `json:"kid"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

// https://github.com/dgrijalva/jwt-go/pull/308 参考
type multiString string
type KeycloakClaims struct {
	Audience multiString `json:"aud,omitempty"`
}

func (ms *multiString) UnmarshalJSON(data []byte) error {
	if len(data) > 0 {
		switch data[0] {
		case '"':
			var s string
			if err := json.Unmarshal(data, &s); err != nil {
				return err
			}
			*ms = multiString(s)
		case '[':
			var s []string
			if err := json.Unmarshal(data, &s); err != nil {
				return err
			}
			*ms = multiString(strings.Join(s, ","))
		}
	}
	return nil
}

// EnsureValidToken is a middleware that will check the validity of our JWT.
func JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Auth0のトークンを取得
		auth0Token := r.Header.Get("X-Forwarded-Authorization")
		// トークンから'Beaere '文字列を取り除く
		rep := regexp.MustCompile(`Bearer `)
		auth0Token = rep.ReplaceAllString(auth0Token, "")
		// トークンの検証
		_, err := verifyToken(auth0Token)
		if err != nil {
			log.Fatalf("verifyToekn Failed. %v", err)
			w.Write([]byte("JWT Verification failed."))
			return
		}
		// Our middleware logic goes here...
		next.ServeHTTP(w, r)
	})
}

func verifyToken(tokenString string) (bool, error) {
	// fetch tenant keys
	tenantKeys, err := jwk.Fetch(context.Background(), fmt.Sprintf("https://%s/.well-known/jwks.json", DomainName))
	if err != nil {
		log.Printf("failed to parse tenant json web keys: err: %v", err)
	}
	log.Printf("tenantKeys: %v", tenantKeys)
	token, err := jwt.Parse(
		[]byte(tokenString),
		jwt.WithKeySet(tenantKeys),
		jwt.WithAudience(Audience),
		jwt.WithAcceptableSkew(time.Minute),
	)
	if token != nil && err != nil {
		log.Printf("failed to parse the token. err: %v", err)
		return false, err
	}
	return true, nil

	// if !token.Valid {
	// 	return false, fmt.Errorf("Invalid token.")
	// } else {
	// 	// check each claim
	// 	iss := "https://" + DomainName + "/"
	// 	checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(iss, true)
	// 	if !checkIss {
	// 		return false, fmt.Errorf("Invalid isssuer.")
	// 	}
	// 	log.Printf("Check isssuer: %v", checkIss)
	// 	// checkAud := token.Claims.(jwt.MapClaims).VerifyAudience(Audience, true)
	// 	// if !checkAud {
	// 	// 	return false, fmt.Errorf("Invalid audience.")
	// 	// }
	// 	// log.Printf("Check audience: %v", checkAud)
	// }

	// return token.Valid, nil

}
