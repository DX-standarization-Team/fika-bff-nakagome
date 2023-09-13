package router

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	executions "cloud.google.com/go/workflows/executions/apiv1"
	executionspb "cloud.google.com/go/workflows/executions/apiv1/executionspb"
	"google.golang.org/api/idtoken"

	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/middleware"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

const Api1Url = "https://fika-api1-nakagome-wsgwmfbvhq-uc.a.run.app"
const Api2Url = "https://fika-api2-nakagome-wsgwmfbvhq-uc.a.run.app"
const workflowUrl = "https://fika-api2-nakagome-wsgwmfbvhq-uc.a.run.app"
const ProjectId = "kaigofika-poc01"
const Location = "us-central1"
const workflowName = "fs-workflow-nakagome"

// New sets up our routes and returns a *http.ServeMux
func New() *http.ServeMux {
	router := http.NewServeMux()
	// This route is always accessible.
	router.Handle("/workflow", http.HandlerFunc(workflowHandler))
	// This route is only accessible if the user has a valid access_token.
	router.Handle("/api2", middleware.JWTAuthMiddleware(http.HandlerFunc(api2Handler)))
	return router
}

// BFF → workflow → api1 呼び出し
func workflowHandler(w http.ResponseWriter, r *http.Request) {
	
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
	log.Printf("workflows response: %v", resp)
	if err != nil {
		http.Error(w, fmt.Sprintf("client.CreateExecution failed...: %v", err), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%v\n", resp)

}

// BFF → api2 呼び出し
func api2Handler(w http.ResponseWriter, r *http.Request) {
	
	// Retrieve the token from the request context
	tokenString := r.Context().Value("token").(string)
	fmt.Fprintf(w, "Token retrived from Context: %s\n", tokenString)
	token, err := jwt.Parse([]byte(tokenString))
	org_idClaim, ok := token.Get("org_id")
	if !ok {
		log.Fatal("org_id claim not found in JWT")
	}
	org_id := org_idClaim.(string)
	log.Printf("org_id: %s\n", org_id)
		
	// Auth0の認証情報をそのまま取り出す
	auth0Token := r.Header.Get("X-Forwarded-Authorization")
	// api2へのAuthorization Headerの引き渡し
	ctx := context.Background()
	client, err := idtoken.NewClient(ctx, Api2Url)
	if err != nil {
		http.Error(w, fmt.Sprintf("idtoken.NewClient failed...: %v", err), http.StatusInternalServerError)
	}
	req, _ := http.NewRequest("GET", Api2Url, nil)
	// header追加 → 上手くいく
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
