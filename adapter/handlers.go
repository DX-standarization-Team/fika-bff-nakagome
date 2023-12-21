package router

import (
	"context"
	"os"

	"fmt"
	"io"
	"log"
	"net/http"

	executions "cloud.google.com/go/workflows/executions/apiv1"
	executionspb "cloud.google.com/go/workflows/executions/apiv1/executionspb"
	"gopkg.in/yaml.v3"

	// authorization "github.com/DX-standarization-Team/common-service-v2/middleware/authorization"
	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/config"
	"google.golang.org/api/idtoken"

	"cloud.google.com/go/logging"
)

var openAPIFile = "/util/openapi.yaml"

type OpenAPISpec struct {
	Paths map[string]PathItem `yaml:"paths"`
}

type PathItem map[string]Operation

type Operation struct {
	OperationID string `yaml:"operationId"`
}

func getOperationID(path, method string, spec *OpenAPISpec) (string, error) {
	pathItem, ok := spec.Paths[path]
	if !ok {
		return "", fmt.Errorf("Path not found: %s", path)
	}

	operation, ok := pathItem[method]
	if !ok {
		return "", fmt.Errorf("Method not found for path %s: %s", path, method)
	}

	return operation.OperationID, nil
}

func loadOpenAPIFile(filePath string) (*OpenAPISpec, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var spec OpenAPISpec
	err = yaml.Unmarshal(content, &spec)
	if err != nil {
		return nil, err
	}

	return &spec, nil
}

// BFF → workflow → api1 呼び出し
func workflowHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("workflowHandler entering")
	// log.Println(r.Header)
	// log.Printf("Request URL: %s", r.URL)
	// log.Printf("Request URL Path: %s", r.URL.Path)
	// log.Printf("Request Method: %s", r.Method)
	// log.Printf("Request URL User: %s", r.URL.User)
	// log.Printf("Request URL RawQuery: %s", r.URL.RawQuery)

	// ------------------- Open API specification --------------------------
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting working directory:", err)
		return
	}

	fmt.Println("Working Directory:", wd)

	_, err = os.Stat(openAPIFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("File does not exist:", openAPIFile)
		} else {
			fmt.Println("Error checking file:", err)
		}
		return
	}
	fmt.Println("File exists:", openAPIFile)

	openAPISpec, err := loadOpenAPIFile(openAPIFile)
	if err != nil {
		log.Println("Error reading OpenAPI file:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	path := r.URL.Path
	method := r.Method

	operationID, err := getOperationID(path, method, openAPISpec)
	if err != nil {
		log.Println("Error getting OperationID:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Printf("OperationID: %s\n", operationID)

	// ------------------- cloud logging --------------------------
	log.Println("cloud logging entering")
	ctx := context.Background()

	// Sets your Google Cloud Platform project ID.
	projectID := "kaigofika-poc01"

	// Creates a client.
	loggingclient, err := logging.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer loggingclient.Close()

	// Sets the name of the log to write to.
	logName := "my-log"

	// Selects the log to write to.
	logger := loggingclient.Logger(logName)

	// Add key-value pairs to the map
	labels := make(map[string]string)
	labels["fikaid"] = "01234"
	labels["tenantId"] = "test-tenant-01"
	labels["ipAddress"] = r.Header.Get("X-Fowarded-For")

	entry := logging.Entry{
		Payload:  fmt.Sprintf("create user succeeded. userId: %s", "testUserId"),
		Severity: logging.Debug,
		HTTPRequest: &logging.HTTPRequest{
			Request: r,
		},
		Labels: labels,
	}
	type Body struct {
		text string
	}
	body := Body{
		text: "test",
	}

	entry.Payload = fmt.Sprintf("%s\nRequest Body: %s", entry.Payload, body)
	// entry.Payload = fmt.Sprintf("%s\nRequest Body: %s", entry.Payload, r.GetBody)
	logger.Log(entry)
	// --------------------------------------------------------------

	log.Printf("Authorization: %s", r.Header.Get("Authorization"))
	log.Printf("X-Forwarded-Authorization: %s", r.Header.Get("X-Forwarded-Authorization"))
	token := r.Header.Get("X-Forwarded-Authorization")
	// ctx := context.Background()

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
	// log.Printf("api2Handler was called")
	// orgId := authorization.GetOrgId(r)
	// log.Printf("orgId: %s", orgId)
	config := config.GetConfig()
	auth0aud := config.Auth0.AUTH0_AUDIENCE
	log.Printf("auth0aud: %s", auth0aud)

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
