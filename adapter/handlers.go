package router

import (
	"context"
	// "encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	executions "cloud.google.com/go/workflows/executions/apiv1"
	executionspb "cloud.google.com/go/workflows/executions/apiv1/executionspb"
	"go.uber.org/zap"

	// authorization "github.com/DX-standarization-Team/common-service-v2/middleware/authorization"
	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/config"
	// "github.com/sirupsen/logrus"
	"google.golang.org/api/idtoken"

	"cloud.google.com/go/logging"
	// logger "github.com/GoogleCloudPlatform/golang-samples/run/helloworld/lib"
	// "go.opentelemetry.io/otel/sdk/trace"
)

type LogContent struct {
	Message     string `json:"message"`
	Severity    string `json:"severity"`
	Trace       string `json:"logging.googleapis.com/trace"`
	OperationId string `json:"operationId"`
}

// BFF → workflow → api1 呼び出し
func workflowHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("workflowHandler entering")
	log.Printf("Header: %v", r.Header)
	operationId := r.Header.Get("traceparent")
	log.Printf("operationId: %v", operationId)
	projectID := "kaigofika-poc01"

	var traceId string
	// Use Sscanf to extract values
	cloudTraceContext := r.Header.Get("X-Cloud-Trace-Context")
	parts := strings.Split(cloudTraceContext, "/")
	log.Printf("parts: %v", parts)
	if len(parts) >= 1 {
		traceId = parts[0]
	}

	// "projects/[プロジェクトID]/traces/[トレースID]" の形式
	trace := "projects/" + projectID + "/traces/" + traceId
	fmt.Println("Trace:", trace)

	// ------------------- zap logger --------------------------
	// conf := zap.Config{
	// 	Level: zap.NewAtomicLevel(),
	// 	// Development: false,
	// 	Encoding: "json",
	// 	EncoderConfig: zapcore.EncoderConfig{
	// 		TimeKey:        "timestamp",
	// 		LevelKey:       "severity",
	// 		NameKey:        "name",
	// 		CallerKey:      "caller",
	// 		MessageKey:     "message",
	// 		StacktraceKey:  "stacktrace",
	// 		EncodeLevel:    zapcore.LowercaseLevelEncoder,
	// 		EncodeTime:     zapcore.ISO8601TimeEncoder,
	// 		EncodeDuration: zapcore.StringDurationEncoder,
	// 		EncodeCaller:   zapcore.ShortCallerEncoder,
	// 	},
	// 	OutputPaths:      []string{"stdout"},
	// 	ErrorOutputPaths: []string{"stderr"},
	// 	// OutputPaths:      []string{"stdout", "./log/development.out.log"},
	// 	// ErrorOutputPaths: []string{"stderr", "./log/development.err.log"},
	// }
	// zaplogger, err := conf.Build()
	// if err != nil {
	// 	log.Fatalf("Failed to create zap client: %v", err)
	// }
	// defer zaplogger.Debug(
	// 	"Zap logging test",
	// 	zap.String("trace", trace),
	// 	zap.String("operationId", operationId),
	// )

	// ------------------- zap logger 2 --------------------------
	zaplogger, err := zap.NewProduction()
	if err != nil {
		panic("Failed to initialize Zap logger")
	}
	defer zaplogger.Sync() // Flushes buffer, if any

	// Log with fields
	zaplogger.Info("### Zap logging test",
		zap.String("logging.googleapis.com/trace", trace),
		zap.String("operationId", operationId),
	)

	// ------------------- log package → 構造体ログ出力できない --------------------------
	// logMessage := LogContent{
	// 	Message:     "### log package test",
	// 	Severity:    "DEBUG",
	// 	Trace:       trace,
	// 	OperationId: operationId,
	// }
	// log.Println(json.Marshal(logMessage))

	// ------------------- cloud logging --------------------------
	log.Println("cloud logging entering")
	ctx := context.Background()

	// Sets your Google Cloud Platform project ID.
	// projectID := "kaigofika-poc01"

	// Creates a client.
	loggingclient, err := logging.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer loggingclient.Close()
	log.Printf("client: %v", loggingclient)

	// Sets the name of the log to write to.
	logName := "my-log"
	// Selects the log to write to.
	logger := loggingclient.Logger(logName)
	log.Printf("logger: %v", logger)
	// --------------------------------------------------------------
	// logger := logger.NewLogger()
	logger.Log(logging.Entry{
		Payload: struct{ Message string }{
			Message: "workflowHandler entering",
		},
		Severity: logging.Debug,
		HTTPRequest: &logging.HTTPRequest{
			Request: r,
		},
	})
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
