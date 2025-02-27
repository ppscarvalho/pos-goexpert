package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Estrutura para mapear a resposta da API de câmbio
type ExchangeRate struct {
	USDBRL struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

// Banco de dados SQLite
const dbFile = "cotacao.db"

func main() {
	// Criar banco de dados e tabela
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal("Erro ao abrir o banco:", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS cotacoes (id INTEGER PRIMARY KEY AUTOINCREMENT, bid TEXT, timestamp DATETIME DEFAULT CURRENT_TIMESTAMP)")
	if err != nil {
		log.Fatal("Erro ao criar tabela:", err)
	}

	// Configurar rota HTTP
	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
		defer cancel()

		rate, err := getDollarExchangeRate(ctx)
		if err != nil {
			http.Error(w, "Erro ao obter cotação", http.StatusInternalServerError)
			log.Println("Erro ao obter cotação:", err)
			return
		}

		// Criar contexto para inserção no banco (timeout de 10ms)
		dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer dbCancel()

		err = saveExchangeRate(dbCtx, db, rate)
		if err != nil {
			log.Println("Erro ao salvar cotação no banco:", err)
		}

		// Retornar JSON com a cotação
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"bid": rate})
	})

	log.Println("Servidor rodando na porta 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Obtém a cotação do dólar na API externa
func getDollarExchangeRate(ctx context.Context) (string, error) {
	// Criar uma requisição HTTP com timeout maior
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond) // Timeout maior
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return "", err
	}

	start := time.Now()
	resp, err := http.DefaultClient.Do(req)
	elapsed := time.Since(start)

	if err != nil {
		log.Println("Erro ao chamar API:", err)
		return "", err
	}
	defer resp.Body.Close()

	log.Printf("API respondeu em %s\n", elapsed)

	var exchange ExchangeRate
	if err := json.NewDecoder(resp.Body).Decode(&exchange); err != nil {
		return "", err
	}

	return exchange.USDBRL.Bid, nil
}

// Salva a cotação no banco de dados SQLite
func saveExchangeRate(ctx context.Context, db *sql.DB, bid string) error {
	_, err := db.ExecContext(ctx, "INSERT INTO cotacoes (bid) VALUES (?)", bid)
	return err
}
