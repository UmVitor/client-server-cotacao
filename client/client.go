package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type CotacaoResponse struct {
	Bid float64 `json:"bid"`
}

var URLCotacao = "http://localhost:8080/cotacao"

func InitClient() {
	log.Println("Initializing client...")
	req, err := http.NewRequest(http.MethodGet, URLCotacao, nil)
	if err != nil {
		log.Print(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Millisecond*300))
	defer cancel()

	req = req.WithContext(ctx)
	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("Request timed out")
		} else {
			log.Println("Request failed:", err)
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	var cotacaoResponse CotacaoResponse
	err = json.Unmarshal(body, &cotacaoResponse)
	if err != nil {
		log.Println(err)
	}

	writeInFile := fmt.Sprintf("Dolar: %v\n", cotacaoResponse.Bid)

	f, err := os.OpenFile("cotacao.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	if _, err = f.WriteString(writeInFile); err != nil {
		log.Println(err)
	}
}
