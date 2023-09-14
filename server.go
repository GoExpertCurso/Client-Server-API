package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

type Cotacao struct {
	USDBRL USDBRL
}

type USDBRL struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

func main() {
	println("Initializing server...")

	ctx := context.Background()
	fmt.Printf("ctx: %v\n", ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", CotacaoHandler)

	http.ListenAndServe(":8080", mux)
}

func CotacaoHandler(w http.ResponseWriter, r *http.Request) {
	client := http.Client{}
	ctx := r.Context()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	w.Write([]byte(string(body)))

	var ctc Cotacao
	if err = json.Unmarshal(body, &ctc); err != nil {
		log.Fatal(err)
	}

	db, err := Conectar()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = insertCotacao(db, ctc)
	if err != nil {
		panic(err)
	}
}

func Conectar() (*sql.DB, error) {
	db, err := sql.Open("mysql", "root:root@tcp(localhost:3306)/goexpert")
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func insertCotacao(db *sql.DB, cotacao Cotacao) error {
	stmt, err := db.Prepare(`INSERT INTO cotacao(code, codein, name, 
		high, low, varBid, pctChange, Bid, Ask, timestamp, create_date) 
		VALUES(?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(cotacao.USDBRL.Code, cotacao.USDBRL.Codein, cotacao.USDBRL.Name, cotacao.USDBRL.High, cotacao.USDBRL.Low,
		cotacao.USDBRL.VarBid, cotacao.USDBRL.PctChange, cotacao.USDBRL.Bid, cotacao.USDBRL.Ask, cotacao.USDBRL.Timestamp,
		cotacao.USDBRL.CreateDate)
	if err != nil {
		return err
	}
	return nil
}
