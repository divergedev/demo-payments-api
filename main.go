package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "baseline"
	}
	accountsURL := os.Getenv("ACCOUNTS_API_URL")
	if accountsURL == "" {
		accountsURL = "http://accounts-api:8080"
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "payments-api", "version": version})
	})

	http.HandleFunc("/api/payments", func(w http.ResponseWriter, r *http.Request) {
		client := &http.Client{Timeout: 5 * time.Second}
		req, _ := http.NewRequestWithContext(r.Context(), "GET", accountsURL+"/api/accounts/balance", nil)
		if pid := r.Header.Get("x-preview-id"); pid != "" {
			req.Header.Set("x-preview-id", pid)
		}

		var balance string
		resp, err := client.Do(req)
		if err != nil {
			balance = "unavailable"
		} else {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			balance = string(body)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"service":  "payments-api",
			"version":  version,
			"preview":  r.Header.Get("x-preview-id"),
			"balance":  balance,
			"payments": []map[string]interface{}{
				{"id": "pay-001", "amount": 150.00, "status": "completed"},
				{"id": "pay-002", "amount": 75.50, "status": "pending"},
			},
		})
	})

	log.Printf("payments-api %s listening on :%s", version, port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
