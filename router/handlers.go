package router

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	executions "cloud.google.com/go/workflows/executions/apiv1"
	executionspb "cloud.google.com/go/workflows/executions/apiv1/executionspb"

	authorization "github.com/DX-standarization-Team/common-service-v2/middleware/authorization"
	"google.golang.org/api/idtoken"
)

// BFF → workflow → api1 呼び出し
func workflowHandler(w http.ResponseWriter, r *http.Request) {

	token := r.Header.Get("X-Forwarded-Authorization")
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
			// Argument: `{"auth0-token":"` + token + `"}`,
			Argument: `{"X-Forwarded-Authorization":"` + token + `"}`,
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

	log.Printf("api2Handler was called")

	orgId := authorization.GetOrgId(r)
	log.Printf("orgId: %s", orgId)
	// context に含まれる jwt から org_id を抽出テスト
	// ctx := context.Background() ⇒ r.Context() ではうまくいくのに context.Backgroud()なぜか token が取り出せなかった

	// サービス間認証できる client の作成
	client, err := idtoken.NewClient(r.Context(), Api2Url)
	if err != nil {
		log.Fatalf("%v", err)
	}
	req, err := http.NewRequest(http.MethodGet, Api2Url, nil)
	if err != nil {
		log.Fatalf("%v", err)
	}

	// トークンヘッダ追加
	token := r.Header.Get("X-Forwarded-Authorization")
	// req.Header.Add("auth0-token", token)
	req.Header.Add("X-Forwarded-Authorization", token)
	log.Printf("X-Forwarded-Authorization: %s", token)

	// 冪等キーヘッダ追加
	idempotencyKey := r.Header.Get("Idempotency-Key")
	req.Header.Add("Idempotency-Key", idempotencyKey)
	log.Printf("Idempotency-Key: %s", idempotencyKey)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer resp.Body.Close()
	log.Printf("Call API2. statuscode: %v, body: %v", resp.StatusCode, resp.Body)

	// API2の処理結果をレスポンスに格納
	w.WriteHeader(resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	fmt.Fprintf(w, "%s\n", string(body))
}
