package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

type CandyOrder struct {
	Money      int    `json:"money"`
	CandyType  string `json:"candyType"`
	CandyCount int    `json:"candyCount"`
}

func main() {
	candyType := flag.String("k", "", "Type of candy")
	candyCount := flag.Int("c", 0, "Count of candy")
	money := flag.Int("m", 0, "Amount of money")
	flag.Parse()

	order := CandyOrder{
		Money:      *money,
		CandyType:  *candyType,
		CandyCount: *candyCount,
	}

	orderJSON, err := json.Marshal(order)
	if err != nil {
		log.Fatalf("Error marshaling order: %v", err)
	}

	caCert, err := os.ReadFile("/home/valero/Desktop/s21_Go/4/ex01/ca/minica.pem")

	if err != nil {
		log.Fatalf("Error loading CA file: %v", err)
	}

	caCertPool, _ := x509.SystemCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	log.Println("RootCA loaded")

	cert, err := tls.LoadX509KeyPair("/home/valero/Desktop/s21_Go/4/ex01/ca/client/cert.pem",
		"/home/valero/Desktop/s21_Go/4/ex01/ca/client/key.pem")
	if err != nil {
		log.Fatalf("Error loading client certificate: %v", err)
	}

	tlsConfig := &tls.Config{
		RootCAs:      caCertPool,
		Certificates: []tls.Certificate{cert},
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	resp, err := client.Post("https://candy.tld:3333/buy_candy", "application/json", bytes.NewBuffer(orderJSON))
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}
	defer resp.Body.Close()
	resp.Header.Set("Content-Type", "application/json")

	if resp.StatusCode != http.StatusCreated {
		log.Fatalf("Error response from server: %s", resp.Status)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Fatalf("Error decoding response: %v", err)
	}

	fmt.Printf("Thank you! Your change is %v\n", response["change"])
}
