package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Fatalf("%v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req = req.WithContext(ctx)

	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("%v", err)
	}

	log.Printf("%v\n", res.Status)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("%v", err)
	}

	var bid string
	err = json.Unmarshal(body, &bid)
	handleFile(bid)

}

func handleFile(bid string) {
	file, err := os.OpenFile("cotacao.txt", os.O_RDWR|os.O_APPEND, 0755)
	if err != nil {
		log.Printf("%v", err)
		file, err = os.Create("cotacao.txt")
	}

	_, err = file.WriteString("Dólar: " + bid + "\n")
	if err != nil {
		log.Fatalf("%v", err)
	}
}
