package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	executions "cloud.google.com/go/workflows/executions/apiv1"
	executionspb "cloud.google.com/go/workflows/executions/apiv1/executionspb"
	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/form3tech-oss/jwt-go"
	"google.golang.org/api/idtoken"

	"encoding/json"
)

const Api1Url = "https://fika-api1-nakagome-wsgwmfbvhq-uc.a.run.app"
const Api2Url = "https://fika-api2-nakagome-wsgwmfbvhq-uc.a.run.app"
const workflowUrl = "https://fika-api2-nakagome-wsgwmfbvhq-uc.a.run.app"
const ProjectId = "kaigofika-poc01"
const Location = "us-central1"
const workflowName = "fs-workflow-nakagome"

const Audience = "https://fs-apigw-bff-nakagome-bi5axj14.uc.gateway.dev/"
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

func main() {
	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: verifyToken,
		SigningMethod:       jwt.SigningMethodRS256,
	})
	log.Println(jwtMiddleware)

	http.HandleFunc("/workflow", workflowHandler)
	http.HandleFunc("/api2", api2Handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func verifyToken(token *jwt.Token) (interface{}, error) {
	log.Println("verifyToken entering")
	log.Println("token")
	log.Println(token)
	// Verify 'aud' claim
	aud := Audience
	checkAud := token.Claims.(jwt.MapClaims).VerifyAudience(aud, false)
	if !checkAud {
		fmt.Printf("Invalid audience.\n")
		return token, errors.New("Invalid audience.")
	}
	// Verify 'iss' claim
	iss := "https://" + DomainName + "/"
	checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(iss, false)
	if !checkIss {
		return token, errors.New("Invalid isssuer.")
	}

	cert, err := getPemCert(token)
	if err != nil {
		panic(err.Error())
	}
	result, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
	log.Println("verifyToken exiting")
	return result, nil
}

func getPemCert(token *jwt.Token) (string, error) {
	log.Println("getPemCert entering")
	// Auth0の公開鍵を取得
	cert := ""
	resp, err := http.Get("https://" + DomainName + "/.well-known/jwks.json")

	if err != nil {
		return cert, err
	}
	defer resp.Body.Close()
	log.Println("resp %v", resp.Body)

	var jwks = Jwks{}
	err = json.NewDecoder(resp.Body).Decode(&jwks)

	if err != nil {
		return cert, err
	}

	for k, _ := range jwks.Keys {
		if token.Header["kid"] == jwks.Keys[k].Kid {
			cert = "-----BEGIN CERTIFICATE-----\n" + jwks.Keys[k].X5c[0] + "\n-----END CERTIFICATE-----"
		}
	}

	if cert == "" {
		err := errors.New("Unable to find appropriate key.")
		return cert, err
	}

	log.Println("getPemCert exiting")
	return cert, nil
}

// BFF → workflow → api2 呼び出し
func workflowHandler(w http.ResponseWriter, r *http.Request) {

	// Auth0の認証情報を取り出す
	auth0Token := r.Header.Get("X-Forwarded-Authorization")
	ctx := context.Background()
	client, err := executions.NewClient(ctx)
	if err != nil {
		fmt.Printf("executions.NewClient: %v\n", err)
		// http.Error(w, fmt.Sprintf("...: %w", err), http.StatusInternalServerError)
		return
	}
	defer client.Close()

	req := &executionspb.CreateExecutionRequest{
		Parent: "projects/" + ProjectId + "/locations/" + Location + "/workflows/" + workflowName,
		Execution: &executionspb.Execution{
			Argument: `{"auth0-token":"` + auth0Token + `"}`,
		},
	}

	resp, err := client.CreateExecution(ctx, req)
	if err != nil {
		fmt.Printf("client.CreateExecution: %v\n", err)
		// http.Error(w, fmt.Sprintf("...: %w", err), http.StatusInternalServerError)
	}
	log.Println(resp)
	fmt.Fprintf(w, "%v\n", resp)

}

// BFF → api2 呼び出し
func api2Handler(w http.ResponseWriter, r *http.Request) {
	// Auth0の認証情報を取り出す
	auth0Token := r.Header.Get("X-Forwarded-Authorization")

	token, err := jwt.Parse(r.Header.Get("X-Forwarded-Authorization"), func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodRSA)
		if !ok {
			//    writer.WriteHeader(http.StatusUnauthorized)
			//    _, err := writer.Write([]byte("You're Unauthorized!"))
			//    if err != nil {
			// 	  return nil, err
			//    }
			log.Fatal("token not ok")
		}
		return "", nil
	})
	// 認証情報の検証
	result, err := verifyToken(token)
	log.Println(result)

	// api2へのAuthorization Headerの引き渡し
	ctx := context.Background()
	// TEST - contextがhttpでうまくいかない
	// ctx := context.WithValue(context.Background(), "auth0-token", auth0Token)
	client, err := idtoken.NewClient(ctx, Api2Url)
	// TEST - idtoken.NewTokenSource : NG
	// ts, err := idtoken.NewTokenSource(ctx, Api2Url)
	if err != nil {
		fmt.Printf("idtoken.NewClient: %v\n", err)
		return
	}
	// TEST - NewTokenSource : NG
	// token, err := ts.Token()
	if err != nil {
		// TODO: Handle error.
	}
	// req, _ := http.NewRequestWithContext(ctx, "GET", Api2Url, nil)
	req, _ := http.NewRequest("GET", Api2Url, nil)
	// TEST - NewTokenSource : NG
	// token.SetAuthHeader(req)
	// TEST - WithContextでcontextを操作 → うまくいかない
	// req = req.WithContext(ctx)
	// header追加 → 上手くいく
	req.Header.Set("auth0-token", auth0Token)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("%v", err)
	}

	// TEST - NewTokenSource
	// resp, err := client.Get(Api2Url)
	// if err != nil {
	// 	fmt.Printf("client.Get: %v\n", err)
	// 	// http.Error(w, fmt.Sprintf("...: %w", err), http.StatusInternalServerError)
	// 	return
	// }
	defer resp.Body.Close()
	// 取得したURLの内容を読み込む
	body, _ := io.ReadAll(resp.Body)
	log.Println(string(body))
	fmt.Fprintf(w, "%s\n", string(body))
}
