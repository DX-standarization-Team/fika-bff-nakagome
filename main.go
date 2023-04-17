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
	"github.com/gin-gonic/gin"
	"google.golang.org/api/idtoken"
)

const Api1Url = "https://fika-api1-nakagome-wsgwmfbvhq-uc.a.run.app"
const Api2Url = "https://fika-api2-nakagome-wsgwmfbvhq-uc.a.run.app"
const workflowUrl = "https://fika-api2-nakagome-wsgwmfbvhq-uc.a.run.app"
const ProjectId = "kaigofika-poc01"
const Location = "us-central1"
const workflowName = "fs-workflow-nakagome"

func main() {

	router := gin.Default()

	router.GET("/workflow", workflowHandler)
	router.GET("/api2", api2Handler)

	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("defaulting to port %s", port)
	}

	router.Run("0.0.0.0:" + port)

}

func workflowHandler(c *gin.Context) {

	// Auth0の認証情報を取り出し、派生コンテキストを返す
	auth0Token := c.Request.Header.Get("X-Forwarded-Authorization")
	ctx := context.WithValue(context.Background(), "auth0-token", auth0Token)
	// ctx := context.Background()

	client, err := executions.NewClient(ctx)
	if err != nil {
		fmt.Printf("executions.NewClient: %v\n", err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	defer client.Close()

	req := &executionspb.CreateExecutionRequest{
		Parent: "projects/" + ProjectId + "/locations/" + Location + "/workflows/" + workflowName,
	}
	resp, err := client.CreateExecution(ctx, req)
	if err != nil {
		fmt.Printf("client.CreateExecution: %v\n", err)
		c.JSON(http.StatusInternalServerError, err)
	}
	log.Println(resp)
	c.JSON(http.StatusOK, resp)

}

func api2Handler(c *gin.Context) {
	// Auth0の認証情報を取り出し、派生コンテキストを返す
	auth0Token := c.Request.Header.Get("X-Forwarded-Authorization")
	ctx := context.WithValue(context.Background(), "auth0-token", auth0Token)
	// ctx := context.Background()

	client, err := idtoken.NewClient(ctx, Api2Url)
	if err != nil {
		fmt.Printf("idtoken.NewClient: %v\n", err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	resp, err := client.Get(Api2Url)
	if err != nil {
		fmt.Printf("client.Get: %v\n", err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	defer resp.Body.Close()
	// 取得したURLの内容を読み込む
	body, _ := io.ReadAll(resp.Body)
	log.Println(string(body))
	c.JSON(resp.StatusCode, string(body))

}
