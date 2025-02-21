package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// Estrutura para capturar a resposta do servidor
type Response struct {
	Bid string `json:"bid"`
}

func main() {
	// Criar contexto para requisição ao servidor (timeout de 300ms)
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	bid, err := getExchangeRate(ctx)
	if err != nil {
		log.Fatal("Erro ao obter cotação:", err)
	}

	// Salvar no arquivo "cotacao.txt"
	err = saveToFile("cotacao.txt", fmt.Sprintf("Dólar: %s", bid))
	if err != nil {
		log.Fatal("Erro ao salvar cotação:", err)
	}

	fmt.Println("Cotação salva com sucesso!")
}

// Obtém a cotação do dólar do servidor
func getExchangeRate(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var exchange Response
	if err := json.NewDecoder(resp.Body).Decode(&exchange); err != nil {
		return "", err
	}

	return exchange.Bid, nil
}

// Salva a cotação no arquivo
func saveToFile(filename, content string) error {
	return ioutil.WriteFile(filename, []byte(content), 0644)
}
