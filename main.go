package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"

	executions "cloud.google.com/go/workflows/executions/apiv1"
	executionspb "cloud.google.com/go/workflows/executions/apiv1/executionspb"
	"google.golang.org/api/idtoken"

	"github.com/dgrijalva/jwt-go"
)

const Api1Url = "https://fika-api1-nakagome-wsgwmfbvhq-uc.a.run.app"
const Api2Url = "https://fika-api2-nakagome-wsgwmfbvhq-uc.a.run.app"
const workflowUrl = "https://fika-api2-nakagome-wsgwmfbvhq-uc.a.run.app"
const ProjectId = "kaigofika-poc01"
const Location = "us-central1"
const workflowName = "fs-workflow-nakagome"

const Audience = "https://fs-apigw-bff-nakagome-bi5axj14.uc.gateway.dev/"
const DomainName = "dev-kjqwuq76z8suldgw.us.auth0.com"

func main() {

	http.HandleFunc("/workflow", workflowHandler)
	http.HandleFunc("/api2", api2Handler)

	port := os.Getenv("PORT")
	if port == "" {
		log.Printf("os.Getenv(PORT) was blank. so we will push push 8080")
		port = "8080"
	}
	log.Print("Server listening on http://localhost:8080")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("There was an error with the http server: %v", err)
	}
}

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

func verifyToken(tokenString string) bool {

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
			log.Fatalf("failed to get certificate: resp.Status: %v, err: %v", resp.Status, err)
		}
		log.Println("succeeded to get certificate")
		defer resp.Body.Close()
		// convert response into Jwks structure
		var jwks = Jwks{}
		err = json.NewDecoder(resp.Body).Decode(&jwks)
		if err != nil {
			log.Fatalf("failed to decode the certificate: %v", err)
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
			log.Fatalf("Unable to find appropriate key.")
		}
		// get a RSA public key from the certificate
		result, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
		log.Printf("result: %v", result)
		// returns *rsa.publicKey in case of rsa
		return result, nil
	})
	if err != nil {
		log.Fatalf("Failed to Parse the token: %v", err)
	}
	// confirm each claim
	iss := "https://" + DomainName + "/"
	checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(iss, true)
	if !checkIss {
		log.Fatalf("Invalid isssuer.")
	}
	log.Printf("Check isssuer: %v", checkIss)
	checkAud := token.Claims.(jwt.MapClaims).VerifyAudience(Audience, true)
	if !checkAud {
		log.Fatalf("Invalid audience.")
	}
	log.Printf("Check audience: %v", checkAud)

	log.Printf("verifyToken exiting. token.Valid: %v", token.Valid)
	return token.Valid && checkIss && checkAud

	// if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
	// 	fmt.Printf("claims %v\n", claims)
	// 	// fmt.Printf("user_id: %v\n", int64(claims["user_id"].(float64)))
	// 	fmt.Printf("exp: %v\n", int64(claims["exp"].(float64)))
	// } else {
	// 	fmt.Println(err)
	// }
}

// BFF → workflow → api2 呼び出し
func workflowHandler(w http.ResponseWriter, r *http.Request) {

	// Auth0の認証情報を取り出す
	auth0Token := r.Header.Get("X-Forwarded-Authorization")
	ctx := context.Background()

	// Workflowアクセス用のクライアントライブラリを準備
	client, err := executions.NewClient(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("executions.NewClient failed...: %v", err), http.StatusInternalServerError)
		return
	}
	defer client.Close()

	// Workflowへ引数のauth0-tokenを指定してアクセス
	req := &executionspb.CreateExecutionRequest{
		Parent: "projects/" + ProjectId + "/locations/" + Location + "/workflows/" + workflowName,
		Execution: &executionspb.Execution{
			Argument: `{"auth0-token":"` + auth0Token + `"}`,
		},
	}
	resp, err := client.CreateExecution(ctx, req)
	if err != nil {
		http.Error(w, fmt.Sprintf("client.CreateExecution failed...: %v", err), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%v\n", resp)

}

// BFF → api2 呼び出し
func api2Handler(w http.ResponseWriter, r *http.Request) {
	// Auth0の認証情報をそのまま取り出す
	auth0Token := r.Header.Get("X-Forwarded-Authorization")
	// token から 'Beaere '文字列を取り除く
	rep := regexp.MustCompile(`Bearer `)
	auth0Token = rep.ReplaceAllString(auth0Token, "")
	// トークンの検証
	result := verifyToken(auth0Token)
	log.Printf("verifyToekn result: %v", result)
	if !result {
		log.Fatal("Token verification failed.")
	}

	// api2へのAuthorization Headerの引き渡し
	ctx := context.Background()
	client, err := idtoken.NewClient(ctx, Api2Url)
	if err != nil {
		http.Error(w, fmt.Sprintf("idtoken.NewClient failed...: %v", err), http.StatusInternalServerError)
	}
	req, _ := http.NewRequest("GET", Api2Url, nil)
	req.Header.Set("auth0-token", auth0Token)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer resp.Body.Close()

	// 取得したURLの内容を読み込む
	body, _ := io.ReadAll(resp.Body)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s\n", string(body))
}
