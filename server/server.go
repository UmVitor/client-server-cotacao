package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

type ExchangeRate struct {
	USDBRL struct {
		Code       string `json:"code"`
		CodeIn     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

const URLExchange = "https://economia.awesomeapi.com.br/json/last/USD-BRL"

type ExchangeResponse struct {
	Bid float64 `json:"bid"`
}

func InitServer() {
	initDB()
	defer DB.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /cotacao", ContacaoHandler)
	mux.HandleFunc("/cotacao/get", GetExchangeByIDHandler)

	log.Println("Initializing server at localhost:8080...")
	http.ListenAndServe("localhost:8080", mux)
}

func ContacaoHandler(w http.ResponseWriter, r *http.Request) {

	req, err := http.NewRequest(http.MethodGet, URLExchange, nil)
	if err != nil {
		log.Print(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Millisecond*200))
	defer cancel()

	req = req.WithContext(ctx)
	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		log.Print(err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	var response *ExchangeRate
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Fatalln(err)
	}

	err = insertExchangeDatabase(*response)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error to save in database: %v", err.Error()), http.StatusInternalServerError)
	}

	bid, _ := strconv.ParseFloat(response.USDBRL.Bid, 64)
	exchangeResponseObject := ExchangeResponse{Bid: bid}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(exchangeResponseObject)

}

func initDB() {
	var err error
	DB, err = sql.Open("sqlite3", "./app.db")
	if err != nil {
		log.Fatal(err)
	}

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS cotacao (
    	id INTEGER PRIMARY KEY AUTOINCREMENT,
    	code TEXT NOT NULL,
    	code_in TEXT NOT NULL,
    	name TEXT NOT NULL,
    	high TEXT NOT NULL,
    	low TEXT NOT NULL,
    	var_bid TEXT NOT NULL,
    	pct_change TEXT NOT NULL,
    	bid TEXT NOT NULL,
    	ask TEXT NOT NULL,
    	timestamp TEXT NOT NULL,
    	create_date TEXT NOT NULL
 	);`

	_, err = DB.Exec(sqlStmt)
	if err != nil {
		log.Fatalf("Error creating table: %q: %s\n", err, sqlStmt)
	}
}

func insertExchangeDatabase(cotacao ExchangeRate) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Millisecond*10))
	defer cancel()
	_, err := DB.ExecContext(ctx, `INSERT INTO cotacao (code, code_in, name, high, low, var_bid, pct_change, bid, ask, timestamp, create_date) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`,
		cotacao.USDBRL.Code,
		cotacao.USDBRL.CodeIn,
		cotacao.USDBRL.Name,
		cotacao.USDBRL.High,
		cotacao.USDBRL.Low,
		cotacao.USDBRL.VarBid,
		cotacao.USDBRL.PctChange,
		cotacao.USDBRL.Bid,
		cotacao.USDBRL.Ask,
		cotacao.USDBRL.Timestamp,
		cotacao.USDBRL.CreateDate)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("Database save - timed out")
		} else {
			log.Println("Database Error to persist:", err)
		}
		return err
	}

	return nil
}
func GetExchangeByIDHandler(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		http.Error(w, "Missing 'id' parameter", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "Invalid 'id' parameter", http.StatusBadRequest)
		return
	}

	row := DB.QueryRow(`SELECT code, code_in, name, high, low, var_bid, pct_change, bid, ask, timestamp, create_date FROM cotacao WHERE id = ?`, id)

	var cotacao ExchangeRate
	err = row.Scan(
		&cotacao.USDBRL.Code,
		&cotacao.USDBRL.CodeIn,
		&cotacao.USDBRL.Name,
		&cotacao.USDBRL.High,
		&cotacao.USDBRL.Low,
		&cotacao.USDBRL.VarBid,
		&cotacao.USDBRL.PctChange,
		&cotacao.USDBRL.Bid,
		&cotacao.USDBRL.Ask,
		&cotacao.USDBRL.Timestamp,
		&cotacao.USDBRL.CreateDate,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Record not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(cotacao)
	if err != nil {
		http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
	}
}
