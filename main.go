package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/idtoken"
)

const Api1Url = "https://fika-api1-nakagome-wsgwmfbvhq-uc.a.run.app"
const Api2Url = "https://fika-api2-nakagome-wsgwmfbvhq-uc.a.run.app"
const ProjectId = "kaigofika-poc01"
const Location = "us-central1"

func main() {

	router := gin.Default()

	router.GET("/", handler)
	router.GET("/api1", api1Handler)
	router.GET("/api2", api2Handler)

	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("defaulting to port %s", port)
	}

	router.Run("0.0.0.0:" + port)

}

func handler(c *gin.Context) {

	ctx := context.Background()
	client, err := idtoken.NewClient(ctx, Api1Url)
	if err != nil {
		fmt.Printf("idtoken.NewClient: %v\n", err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	resp, err := client.Get(Api1Url)
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

	client2, err2 := idtoken.NewClient(ctx, Api2Url)
	if err2 != nil {
		fmt.Printf("idtoken.NewClient: %v\n", err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	resp2, err2 := client2.Get(Api2Url)
	if err2 != nil {
		fmt.Printf("client.Get: %v\n", err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	defer resp2.Body.Close()
	// 取得したURLの内容を読み込む
	body2, _ := io.ReadAll(resp2.Body)
	log.Println(string(body2))
	c.JSON(resp2.StatusCode, string(body2))

}

func api1Handler(c *gin.Context) {

	ctx := context.Background()
	client, err := idtoken.NewClient(ctx, Api1Url)
	if err != nil {
		fmt.Printf("idtoken.NewClient: %v\n", err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	resp, err := client.Get(Api1Url)
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

func api2Handler(c *gin.Context) {

	ctx := context.Background()
	client2, err2 := idtoken.NewClient(ctx, Api2Url)
	if err2 != nil {
		fmt.Printf("idtoken.NewClient: %v\n", err2)
		c.JSON(http.StatusInternalServerError, err2)
		return
	}
	resp2, err2 := client2.Get(Api2Url)
	if err2 != nil {
		fmt.Printf("client.Get: %v\n", err2)
		c.JSON(http.StatusInternalServerError, err2)
		return
	}
	defer resp2.Body.Close()
	// 取得したURLの内容を読み込む
	body2, _ := io.ReadAll(resp2.Body)
	log.Println(string(body2))
	c.JSON(resp2.StatusCode, string(body2))

}
