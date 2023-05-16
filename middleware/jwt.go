package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/form3tech-oss/jwt-go"
)

// golang では slice, array, mapは定数として使用できない(https://tech.notti.link/02b52f2)
const Audience = "https://fs-apigw-bff-nakagome-bi5axj14.uc.gateway.dev/"
const Audience2 = "https://dev-kjqwuq76z8suldgw.us.auth0.com/userinfo"
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

	// Parse and validate, and returns a token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// https://github.com/dgrijalva/jwt-go/issues/438 参考
		// get certificate from JSON Web Key Set from Auth0
		cert := ""
		resp, err := http.Get("https://" + DomainName + "/.well-known/jwks.json")
		if err != nil {
			log.Printf("failed to get certificate: resp.Status: %v, err: %v", resp.Status, err)
		}
		log.Println("succeeded to get certificate")
		defer resp.Body.Close()
		// convert response into Jwks structure
		var jwks = Jwks{}
		err = json.NewDecoder(resp.Body).Decode(&jwks)
		if err != nil {
			log.Printf("failed to decode the certificate: %v", err)
		}
		log.Printf("jwks: %v", jwks)
		// find an appropriate certificate
		for k, _ := range jwks.Keys {
			if token.Header["kid"] == jwks.Keys[k].Kid {
				cert = "-----BEGIN CERTIFICATE-----\n" + jwks.Keys[k].X5c[0] + "\n-----END CERTIFICATE-----"
			}
		}
		log.Printf("cert: %v", cert)
		if cert == "" {
			log.Printf("Unable to find appropriate key.")
		}
		// get a RSA public key from the certificate
		result, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
		log.Printf("result: %v", result)
		// returns *rsa.publicKey in case of rsa
		return result, nil
	})
	if err != nil {
		return false, fmt.Errorf("ailed to Parse the token: %w", err)
	}

	// 取得したトークンで検証
	log.Printf("token.Valid: %v", token.Valid)
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// confirm audience
		log.Printf("aud: %v", claims["aud"])
	} else {
		fmt.Println(err)
	}

	if !token.Valid {
		return false, fmt.Errorf("Invalid token.")
	} else {
		// check each claim
		iss := "https://" + DomainName + "/"
		checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(iss, true)
		if !checkIss {
			return false, fmt.Errorf("Invalid isssuer.")
		}
		log.Printf("Check isssuer: %v", checkIss)
		// checkAud := token.Claims.(jwt.MapClaims).VerifyAudience(Audience, true)
		// if !checkAud {
		// 	return false, fmt.Errorf("Invalid audience.")
		// }
		// log.Printf("Check audience: %v", checkAud)
	}

	return token.Valid, nil

}
