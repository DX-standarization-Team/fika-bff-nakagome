// package main

// import (
// 	"context"
// 	"errors"
// 	"fmt"
// 	"io"
// 	"log"
// 	"net/http"
// 	"os"

// 	executions "cloud.google.com/go/workflows/executions/apiv1"
// 	executionspb "cloud.google.com/go/workflows/executions/apiv1/executionspb"
// 	jwtmiddleware "github.com/auth0/go-jwt-middleware"
// 	"github.com/form3tech-oss/jwt-go"
// 	"google.golang.org/api/idtoken"

// 	"encoding/json"
// )

// const Api1Url = "https://fika-api1-nakagome-wsgwmfbvhq-uc.a.run.app"
// const Api2Url = "https://fika-api2-nakagome-wsgwmfbvhq-uc.a.run.app"
// const workflowUrl = "https://fika-api2-nakagome-wsgwmfbvhq-uc.a.run.app"
// const ProjectId = "kaigofika-poc01"
// const Location = "us-central1"
// const workflowName = "fs-workflow-nakagome"

// const Audience = "https://fs-apigw-bff-nakagome-bi5axj14.uc.gateway.dev/"
// const DomainName = "dev-kjqwuq76z8suldgw.us.auth0.com"

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
// 	X5t []string `json:"x5t"`
// }

// func main() {
// 	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
// 		ValidationKeyGetter: verifyToken,
// 		SigningMethod:       jwt.SigningMethodRS256,
// 	})
// 	log.Println(jwtMiddleware)

// 	http.HandleFunc("/workflow", workflowHandler)
// 	http.HandleFunc("/api2", api2Handler)

// 	port := os.Getenv("PORT")
// 	if port == "" {
// 		port = "8080"
// 	}
// 	if err := http.ListenAndServe(":"+port, nil); err != nil {
// 		log.Fatal(err)
// 	}
// }

// func verifyToken(token *jwt.Token) (interface{}, error) {
// 	log.Println("verifyToken entering")
// 	log.Println("token")
// 	log.Println(token)

// 	// claimsが正しい形式であるか確認
// 	log.Println("claimsが正しい形式であるか確認")
// 	err := token.Claims.(jwt.MapClaims).Valid()
// 	if err != nil {
// 		return token, errors.New("invalid claims type")
// 	}

// 	// Verify 'aud' claim
// 	log.Println("Verify 'aud' claim")
// 	checkAud := token.Claims.(jwt.MapClaims).VerifyAudience(Audience, true)
// 	if !checkAud {
// 		fmt.Printf("Invalid audience.\n")
// 		return token, errors.New("Invalid audience.")
// 	}
// 	// issフィールドを見て、正しいトークン発行者か確認する
// 	log.Println("issフィールドを見て、正しいトークン発行者か確認する")
// 	iss := "https://" + DomainName + "/"
// 	checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(iss, true)
// 	if !checkIss {
// 		return token, errors.New("Invalid isssuer.")
// 	}

// 	log.Println("getPemCert")
// 	cert, err := getPemCert(token)
// 	if err != nil {
// 		panic(err.Error())
// 	}
// 	result, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
// 	log.Println("verifyToken exiting")
// 	return result, nil
// }

// func getPemCert(token *jwt.Token) (string, error) {
// 	log.Println("getPemCert entering")
// 	// Auth0の公開鍵を取得
// 	cert := ""
// 	resp, err := http.Get("https://" + DomainName + "/.well-known/jwks.json")
// 	if err != nil {
// 		log.Fatal("公開鍵を取得失敗")
// 		return cert, err
// 	}
// 	log.Println("公開鍵を取得成功")
// 	log.Println(resp)
// 	defer resp.Body.Close()

// 	var jwks = Jwks{}
// 	err = json.NewDecoder(resp.Body).Decode(&jwks)
// 	if err != nil {
// 		return cert, err
// 	}

// 	// token.Header["kid"] が nil となり、certもうまく取れない。
// 	for k, _ := range jwks.Keys {
// 		if token.Header["kid"] == jwks.Keys[k].Kid {
// 			cert = "-----BEGIN CERTIFICATE-----\n" + jwks.Keys[k].X5c[0] + "\n-----END CERTIFICATE-----"
// 		}
// 	}
// 	if cert == "" {
// 		log.Println("Unable to find appropriate key.")
// 		err := errors.New("Unable to find appropriate key.")
// 		return cert, err
// 	}

// 	log.Println("getPemCert exiting")
// 	return cert, nil
// }

// // BFF → workflow → api2 呼び出し
// func workflowHandler(w http.ResponseWriter, r *http.Request) {

// 	// Auth0の認証情報を取り出す
// 	auth0Token := r.Header.Get("X-Forwarded-Authorization")
// 	ctx := context.Background()
// 	client, err := executions.NewClient(ctx)
// 	if err != nil {
// 		fmt.Printf("executions.NewClient: %v\n", err)
// 		return
// 	}
// 	defer client.Close()

// 	req := &executionspb.CreateExecutionRequest{
// 		Parent: "projects/" + ProjectId + "/locations/" + Location + "/workflows/" + workflowName,
// 		Execution: &executionspb.Execution{
// 			Argument: `{"auth0-token":"` + auth0Token + `"}`,
// 		},
// 	}

// 	resp, err := client.CreateExecution(ctx, req)
// 	if err != nil {
// 		fmt.Printf("client.CreateExecution: %v\n", err)
// 	}
// 	log.Println(resp)
// 	fmt.Fprintf(w, "%v\n", resp)

// }

// // BFF → api2 呼び出し
// func api2Handler(w http.ResponseWriter, r *http.Request) {
// 	// Auth0の認証情報をそのまま取り出す
// 	auth0Token := r.Header.Get("X-Forwarded-Authorization")

// 	// tokenをParseする
// 	token, err := jwt.Parse(auth0Token, func(token *jwt.Token) (interface{}, error) {
// 		// check signing method
// 		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
// 			log.Fatal("You're Unauthorized!")
// 			w.WriteHeader(http.StatusUnauthorized)
// 			_, err := w.Write([]byte("You're Unauthorized!"))
// 			if err != nil {
// 				return nil, err
// 			}
// 		}
// 		return "", nil
// 	})
// 	cert, err := getPemCert(token)
// 	log.Printf(cert)

// 	// log.Println("Raw: ", token.Raw)
// 	// log.Println("Method: ", token.Method)
// 	// log.Println("Header: ", token.Header)
// 	// log.Println("Claims: ", token.Claims)
// 	// log.Println("Signature: ", token.Signature)
// 	// log.Println("Valid: ", token.Valid)

// 	// 認証情報の検証
// 	result, err := verifyToken(token)
// 	if err != nil {
// 		log.Fatalf("検証に失敗。原因：%v", err)
// 	} else {
// 		log.Println("Got rsa.PublicKey! ", result)
// 	}

// 	// api2へのAuthorization Headerの引き渡し
// 	ctx := context.Background()
// 	client, err := idtoken.NewClient(ctx, Api2Url)
// 	if err != nil {
// 		fmt.Printf("idtoken.NewClient: %v\n", err)
// 		return
// 	}
// 	if err != nil {
// 		// TODO: Handle error.
// 	}
// 	req, _ := http.NewRequest("GET", Api2Url, nil)
// 	// header追加 → 上手くいく
// 	req.Header.Set("auth0-token", auth0Token)
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		log.Fatalf("%v", err)
// 	}

// 	defer resp.Body.Close()
// 	// 取得したURLの内容を読み込む
// 	body, _ := io.ReadAll(resp.Body)
// 	log.Println(string(body))
// 	fmt.Fprintf(w, "%s\n", string(body))
// }