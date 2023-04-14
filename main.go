package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	log.Print("starting server...")
	http.HandleFunc("/", handler)
	http.HandleFunc("/api1", api1Handler)
	http.HandleFunc("/api2", api2Handler)

	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("defaulting to port %s", port)
	}

	// Start HTTP server.
	log.Printf("listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	name := os.Getenv("NAME")
	if name == "" {
		name = "bff"
	}
	res1, _ := http.Get("https://go-codescanning-githubactions-cloudrun-api1-wsgwmfbvhq-uc.a.run.app")
	// 取得したURLの内容を読み込む
	body1, _ := io.ReadAll(res1.Body)
	// 取得した情報は[]byteなのでstringに型変換
	fmt.Fprintf(w, "Body %s\n", string(body1))
	fmt.Fprintf(w, "Status %s\n", res1.StatusCode)
	fmt.Fprintf(w, "URL %s\n", res1.Request.URL)
	fmt.Fprintf(w, "Protocol %s\n", res1.Proto)
	fmt.Fprintf(w, "Date %s\n", res1.Header["Date"])
	fmt.Fprintf(w, "Content-Type %s\n", res1.Header["Content-Type"])
	fmt.Fprintf(w, "Method %s\n", res1.Request.Method)

	res2, _ := http.Get("https://go-codescanning-githubactions-cloudrun-api2-wsgwmfbvhq-uc.a.run.app")
	// 取得したURLの内容を読み込む
	body2, _ := io.ReadAll(res2.Body)
	// 取得した情報は[]byteなのでstringに型変換
	fmt.Fprintf(w, "Body %s\n", string(body2))
	fmt.Fprintf(w, "Status %s\n", res2.StatusCode)
	fmt.Fprintf(w, "URL %s\n", res2.Request.URL)
	fmt.Fprintf(w, "Protocol %s\n", res2.Proto)
	fmt.Fprintf(w, "Date %s\n", res2.Header["Date"])
	fmt.Fprintf(w, "Content-Type %s\n", res2.Header["Content-Type"])
	fmt.Fprintf(w, "Method %s\n", res2.Request.Method)
}

func api1Handler(w http.ResponseWriter, r *http.Request) {
	name := os.Getenv("NAME")
	if name == "" {
		name = "bff"
	}
	res1, _ := http.Get("https://go-codescanning-githubactions-cloudrun-api1-wsgwmfbvhq-uc.a.run.app")
	// 取得したURLの内容を読み込む
	body1, _ := io.ReadAll(res1.Body)
	// 取得した情報は[]byteなのでstringに型変換
	fmt.Fprintf(w, "Body %s\n", string(body1))
	fmt.Fprintf(w, "Status %s\n", res1.StatusCode)
	fmt.Fprintf(w, "URL %s\n", res1.Request.URL)
	fmt.Fprintf(w, "Protocol %s\n", res1.Proto)
	fmt.Fprintf(w, "Date %s\n", res1.Header["Date"])
	fmt.Fprintf(w, "Content-Type %s\n", res1.Header["Content-Type"])
	fmt.Fprintf(w, "Method %s\n", res1.Request.Method)

}

func api2Handler(w http.ResponseWriter, r *http.Request) {
	name := os.Getenv("NAME")
	if name == "" {
		name = "bff"
	}

	res2, _ := http.Get("https://go-codescanning-githubactions-cloudrun-api2-wsgwmfbvhq-uc.a.run.app")
	// 取得したURLの内容を読み込む
	body2, _ := io.ReadAll(res2.Body)
	// 取得した情報は[]byteなのでstringに型変換
	fmt.Fprintf(w, "Body %s\n", string(body2))
	fmt.Fprintf(w, "Status %s\n", res2.StatusCode)
	fmt.Fprintf(w, "URL %s\n", res2.Request.URL)
	fmt.Fprintf(w, "Protocol %s\n", res2.Proto)
	fmt.Fprintf(w, "Date %s\n", res2.Header["Date"])
	fmt.Fprintf(w, "Content-Type %s\n", res2.Header["Content-Type"])
	fmt.Fprintf(w, "Method %s\n", res2.Request.Method)
}
