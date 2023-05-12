package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

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

	router := http.NewServeMux()

	http.HandleFunc("/workflow", workflowHandler)
	http.HandleFunc("/api2", api2Handler)

	port := os.Getenv("PORT")
	if port == "" {
		log.Printf("os.Getenv(PORT) was blank. so we will push push 8080")
		port = "8080"
	}
	log.Print("Server listening on http://localhost:8080")
	if err := http.ListenAndServe(":"+port, router); err != nil {
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
	X5t []string `json:"x5t"`
}

func verifyToken(tokenString string) bool {
	log.Printf("verifyToken entering tokenString: %v", tokenString)

	// トークンを解析
	// token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
	// 	return []byte("AllYourBase"), nil
	// })
	// if err != nil {
	// 	log.Fatal("Failed to retrieve the token")
	// }
	// log.Printf("Succeeded to retrieve the token %v", token)

	// jwt.SigningMethodRS256.Verify()

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// https://github.com/dgrijalva/jwt-go/issues/438 参考
		// JSON Web Key Set取得
		cert := ""
		log.Println("get certificate")
		resp, err := http.Get("https://" + DomainName + "/.well-known/jwks.json")
		if err != nil {
			log.Fatal("failed to get certificate")
		}
		log.Println("succeeded to get certificate")
		defer resp.Body.Close()
		var jwks = Jwks{}
		err = json.NewDecoder(resp.Body).Decode(&jwks)
		if err != nil {
			log.Fatal("feiled to decode the certificate")
		}
		log.Printf("jwks: %v", jwks)
		for k, _ := range jwks.Keys {
			if token.Header["kid"] == jwks.Keys[k].Kid {
				cert = "-----BEGIN CERTIFICATE-----\n" + jwks.Keys[k].X5c[0] + "\n-----END CERTIFICATE-----"
			}
		}
		if cert == "" {
			log.Fatalf("Unable to find appropriate key.")
		}
		result, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
		return result, nil
		// return []byte("SECRET_KEY"), nil
	})
	if err != nil {
		log.Fatalln("Failed to Parse the token")
	}

	log.Printf("token.Valid: %v", token.Valid)
	// confirm each claim
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		log.Printf("aud: %v\n", claims["aud"])
		log.Printf("exp: %v\n", int64(claims["exp"].(float64)))
	} else {
		log.Println(err)
	}

	log.Println("verifyToken exiting")
	return token.Valid

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

	result := verifyToken(auth0Token)
	log.Printf("verifyToekn result: %v", result)

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
