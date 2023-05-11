package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	executions "cloud.google.com/go/workflows/executions/apiv1"
	executionspb "cloud.google.com/go/workflows/executions/apiv1/executionspb"
	"google.golang.org/api/idtoken"

	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/middleware"
	"github.com/joho/godotenv"
)

const Api1Url = "https://fika-api1-nakagome-wsgwmfbvhq-uc.a.run.app"
const Api2Url = "https://fika-api2-nakagome-wsgwmfbvhq-uc.a.run.app"
const workflowUrl = "https://fika-api2-nakagome-wsgwmfbvhq-uc.a.run.app"
const ProjectId = "kaigofika-poc01"
const Location = "us-central1"
const workflowName = "fs-workflow-nakagome"

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading the .env file: %v", err)
	}

	router := http.NewServeMux()

	// http.HandleFunc("/workflow", workflowHandler)
	// http.HandleFunc("/api2", api2Handler)
	router.Handle("/workflow", http.HandlerFunc(workflowHandler))
	router.Handle("/api2", middleware.EnsureValidToken()(http.HandlerFunc(api2Handler)))

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

	// api2へのAuthorization Headerの引き渡し
	ctx := context.Background()
	// TEST - contextがhttpでうまくいかない
	// ctx := context.WithValue(context.Background(), "auth0-token", auth0Token)
	client, err := idtoken.NewClient(ctx, Api2Url)
	// TEST - idtoken.NewTokenSource : NG
	// ts, err := idtoken.NewTokenSource(ctx, Api2Url)
	// TEST - NewTokenSource : NG
	// token, err := ts.Token()
	if err != nil {
		http.Error(w, fmt.Sprintf("idtoken.NewClient failed...: %v", err), http.StatusInternalServerError)
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
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s\n", string(body))
}
