package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/sse", sseHandler)
	// 启动服务
	fmt.Println("SSE server listening on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Failed to start server: ", err)
	}
}

func sseHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// set cors,nocache,connection,etc header
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	for {
		fmt.Fprintf(w, "data: %s\n\n", time.Now().String())
		flusher.Flush()
		time.Sleep(2 * time.Second)
	}
}
