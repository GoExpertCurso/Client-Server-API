package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

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
	log.Println("Initializing server...")

	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", CotacaoHandler)

	http.ListenAndServe(":8080", mux)
}

func CotacaoHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Calling economia.awesomeapi.com.br")
	client := http.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

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

	var ctc Cotacao
	if err = json.Unmarshal(body, &ctc); err != nil {
		log.Fatal(err)
	}

	db, err := Conectar()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctxDB, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err = insertCotacao(db, ctc, ctxDB)
	if err != nil {
		log.Fatal(err)
	}

	err = json.NewEncoder(w).Encode(ctc.USDBRL.Bid)
	if err != nil {
		log.Fatal(err)
	}
}

func Conectar() (*sql.DB, error) {
	log.Println("Connecting to database...")
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

func insertCotacao(db *sql.DB, cotacao Cotacao, ctx context.Context) error {
	log.Println("Inserting in the table cotacao...")
	stmt, err := db.PrepareContext(ctx, `INSERT INTO cotacao(code, codein, name, 
		high, low, varBid, pctChange, Bid, Ask, timestamp, create_date) 
		VALUES(?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, cotacao.USDBRL.Code, cotacao.USDBRL.Codein, cotacao.USDBRL.Name, cotacao.USDBRL.High, cotacao.USDBRL.Low,
		cotacao.USDBRL.VarBid, cotacao.USDBRL.PctChange, cotacao.USDBRL.Bid, cotacao.USDBRL.Ask, cotacao.USDBRL.Timestamp,
		cotacao.USDBRL.CreateDate)
	if err != nil {
		return err
	}
	return nil

}
