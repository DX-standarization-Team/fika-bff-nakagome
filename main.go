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
)

const Api1Url = "https://fika-api1-nakagome-wsgwmfbvhq-uc.a.run.app"
const Api2Url = "https://fika-api2-nakagome-wsgwmfbvhq-uc.a.run.app"
const workflowUrl = "https://fika-api2-nakagome-wsgwmfbvhq-uc.a.run.app"
const ProjectId = "kaigofika-poc01"
const Location = "us-central1"
const workflowName = "fs-workflow-nakagome"

func main() {
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

func workflowHandler(w http.ResponseWriter, r *http.Request) {

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
	}
	resp, err := client.CreateExecution(ctx, req)
	if err != nil {
		fmt.Printf("client.CreateExecution: %v\n", err)
		// http.Error(w, fmt.Sprintf("...: %w", err), http.StatusInternalServerError)
	}
	log.Println(resp)
	fmt.Fprintf(w, "%v\n", resp)

}

func api2Handler(w http.ResponseWriter, r *http.Request) {
	// Auth0の認証情報を取り出す
	// auth0Token := r.Header.Get("X-Forwarded-Authorization")
	ctx := context.Background()
	// contextがhttpでうまくいかない
	// ctx := context.WithValue(context.Background(), "auth0-token", auth0Token)
	client, err := idtoken.NewClient(ctx, Api2Url)
	if err != nil {
		fmt.Printf("idtoken.NewClient: %v\n", err)
		return
	}
	// req, _ := http.NewRequestWithContext(ctx, "GET", Api2Url, nil)
	req, _ := http.NewRequest("GET", Api2Url, nil)
	// こちらの方法でcontextを操作→うまくいかない
	// req = req.WithContext(ctx)
	// header追加は上手くいく
	// req.Header.Set("auth0-token", auth0Token)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("%v", err)
	}

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
