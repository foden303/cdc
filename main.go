package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	data := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "pimlico_getUserOperationGasPrice",
		"id":      "c933d10a-7c09-4e90-90a6-223a0d659eb9",
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Lỗi khi encode JSON: %s\n", err)
		return
	}

	url := "https://api.pimlico.io/v2/56/rpc?apikey=pim_i9T2hZbvYg5SH7McFy6XyV"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("error request: %s\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error request: %s\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("err response: %s\n", err)
		return
	}

	// In kết quả
	fmt.Println("Status Code:", resp.Status)
	fmt.Println("Response Body:", string(body))
}
