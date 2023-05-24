package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

// golang では slice, array, mapは定数として使用できない(https://tech.notti.link/02b52f2)
const Audience = "https://fs-apigw-bff-nakagome-bi5axj14.uc.gateway.dev/"

// const Audience2 = "https://dev-kjqwuq76z8suldgw.us.auth0.com/userinfo"
const DomainName = "dev-kjqwuq76z8suldgw.us.auth0.com"

// type Jwks struct {
// 	Keys []JSONWebKeys `json:"keys"`
// }

// type JSONWebKeys struct {
// 	Kty string   `json:"kty"`
// 	Kid string   `json:"kid"`
// 	Use string   `json:"use"`
// 	N   string   `json:"n"`
// 	E   string   `json:"e"`
// 	X5c []string `json:"x5c"`
// }

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
	log.Printf("tenantKeys: %v", tenantKeys.Len())
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

	// WithAcceptableSkewの検証
	log.Printf("WithAcceptableSkewの検証")
	exp := token.Expiration()                   //tokenに含まれている有効期限
	log.Printf("tokenに含まれている有効期限。exp: %v", exp) //UNIXのグリニッジ標準（GMT）時間
	diff := exp.Sub(time.Now())
	log.Printf("現時刻: %v", time.Now())
	log.Printf("現時刻との差分。 diff: %v", diff)
	// token.Set("exp", token.Expiration().Add(-(diff + 10000000000)))
	// log.Printf("有効期限を現時刻の10秒前にセット。 exp: %v", token.Expiration())
	token.Set("exp", token.Expiration().Add(-(diff + time.Duration(2)*time.Minute)))
	log.Printf("有効期限を現時刻の2分前にセット。 exp: %v", token.Expiration())
	token2, err := jwt.Parse(
		[]byte(tokenString),
		jwt.WithKeySet(tenantKeys),
		jwt.WithValidate(true),
		jwt.WithAudience(Audience),
		jwt.WithAcceptableSkew(time.Minute),
	)
	if token2 != nil && err != nil {
		log.Printf("failed to parse the token. err: %v", err)
		return false, err
	}

	return true, nil

}
