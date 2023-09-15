package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"

	"os"

	"github.com/joho/godotenv"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

// JWT の有効性チェックを行うミドルウェア
func JwtAuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenHeader := r.Header.Get("X-Forwarded-Authorization")
		// 'Bearer '文字列を取り除きトークン検証
		rep := regexp.MustCompile(`Bearer `)
		tokenStr := rep.ReplaceAllString(tokenHeader, "")
		token, err := verifyToken(tokenStr)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		// リクエスト Context にトークンを格納
		ctx := context.WithValue(r.Context(), "token", token)
		// 更新済み Context で次の http を呼び出す
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// JWT の検証
func verifyToken(tokenString string) (jwt.Token, error) {
	// 環境変数読み込み
	auth0Domain := ""
	auth0Audience := ""
	if err := godotenv.Load(); err != nil {
		// log.Fatalf("Error loading the .env file: %v", err)
		log.Printf("Error loading the .env file: %v", err)
		auth0Domain = "dev-kjqwuq76z8suldgw.us.auth0.com"
		auth0Audience = "https://fs-apigw-bff-nakagome-bi5axj14.uc.gateway.dev/"
	} else {
		auth0Domain = os.Getenv("AUTH0_DOMAIN")
		auth0Audience = os.Getenv("AUTH0_AUDIENCE")
	}
		
	// tenant keysを取得
	tenantKeys, err := jwk.Fetch(context.Background(), fmt.Sprintf("https://%s/.well-known/jwks.json", auth0Domain))
	if err != nil {
		// log.Printf("failed to parse tenant json web keys: err: %v", err)
		return nil, fmt.Errorf("failed to parse tenant json web keys: err: %v", err)
	}
	// log.Printf("tenantKeys: %v", tenantKeys.Len())
	token, err := jwt.Parse(
		[]byte(tokenString),
		jwt.WithKeySet(tenantKeys),
		jwt.WithValidate(true),
		jwt.WithAudience(auth0Audience),
		jwt.WithAcceptableSkew(time.Minute),
	)
	if err != nil {
		// log.Printf("failed to parse the token. err: %v", err)
		return nil, fmt.Errorf("failed to parse the token. err: %v", err)
	}
	
	// // WithAcceptableSkewの検証
	// exp := token.Expiration()	// token有効期限
	// log.Printf("tokenに含まれている有効期限。exp: %v", exp) //UNIXのグリニッジ標準（GMT）時間
	// diff := exp.Sub(time.Now()) // 現時刻との差分
	// token.Set("exp", token.Expiration().Add(-(diff + 10000000000)))
	// log.Printf("有効期限を現時刻の10秒前にセット。 exp: %v", token.Expiration())
	// 1分の猶予時間をセット
	// valid := jwt.Validate(token, jwt.WithAcceptableSkew(time.Minute))
	// if valid != nil {
	// 	return fmt.Errorf("token is expired. err: %v", valid)
	// }
	return token, nil
}
