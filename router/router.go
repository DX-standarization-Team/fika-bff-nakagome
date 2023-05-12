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
<<<<<<< HEAD
	router.Handle("/api2", middleware.EnsureValidToken()(
		http.HandlerFunc(api2Handler),
	))
=======
	// router.Handle("/api2", middleware.VerifyToekn()(
	// 	http.HandlerFunc(api2Handler),
	// ))
	router.Handle("/api2", middleware.JWTAuthMiddleware(http.HandlerFunc(api2Handler)))
>>>>>>> 03ef2c790aaff52849867f06dc32e57e7228966a

	return router
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
