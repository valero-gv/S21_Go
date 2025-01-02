package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type Place struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Address  string `json:"address"`
	Phone    string `json:"phone"`
	Location struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	} `json:"location"`
}

func main() {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	cfg := elasticsearch.Config{
		Addresses: []string{
			"https://localhost:9200",
		},
		Username:  "elastic",
		Password:  "GbU9-8IatX7etp6twbz1",
		Transport: tr,
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}

	res, err := es.Info()
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Fatalf("Error: %s", res.String())
	}

	fmt.Println(res)

	indexName := "places"
	existsReq := esapi.IndicesExistsRequest{
		Index: []string{indexName},
	}

	existsRes, err := existsReq.Do(context.Background(), es)
	if err != nil {
		log.Fatalf("Error checking if index exists: %s", err)
	}
	defer existsRes.Body.Close()

	if existsRes.StatusCode == 200 {
		deleteReq := esapi.IndicesDeleteRequest{
			Index: []string{indexName},
		}
		deleteRes, err := deleteReq.Do(context.Background(), es)
		if err != nil {
			log.Fatalf("Error deleting existing index: %s", err)
		}
		defer deleteRes.Body.Close()

		if deleteRes.IsError() {
			log.Fatalf("Error: %s", deleteRes.String())
		}
		fmt.Println("Existing index deleted")
	}

	req := esapi.IndicesCreateRequest{
		Index: "places",
	}

	res, err = req.Do(context.Background(), es)
	if err != nil {
		log.Fatalf("Cannot create index: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Fatalf("Error: %s", res.String())
	}

	fmt.Println(res)

	mapping := `{
        "properties": {
            "name": {
                "type": "text"
            },
            "address": {
                "type": "text"
            },
            "phone": {
                "type": "text"
            },
            "location": {
                "type": "geo_point"
            }
        }
    }`
	putMappingRequest := esapi.IndicesPutMappingRequest{
		Index: []string{"places"},
		Body:  strings.NewReader(mapping),
	}
	res, err = putMappingRequest.Do(context.Background(), es)
	if err != nil {
		log.Fatalf("Error putting mapping: %s", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		log.Fatalf("Error response from server: %s", res.String())
	}
	fmt.Println("Mapping applied")

	var m map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&m); err == nil {
		fmt.Println(m)
	} else {
		log.Fatalf("Error parsing the response body: %s", err)
	}

	jsonData, err := os.ReadFile("data.json")
	if err != nil {
		log.Fatalf("Error reading JSON file: %s", err)
	}

	var places []Place
	err = json.Unmarshal(jsonData, &places)
	if err != nil {
		log.Fatalf("Error parsing JSON data: %s", err)
	}
	var bulkData strings.Builder
	for _, place := range places {
		meta := fmt.Sprintf(`{ "index" : { "_index" : "places" } }%s`, "\n")
		data, err := json.Marshal(place)
		if err != nil {
			log.Fatalf("Error marshaling place data: %s", err)
		}
		bulkData.WriteString(meta)
		bulkData.Write(data)
		bulkData.WriteString("\n")
	}

	BulkRequest := esapi.BulkRequest{
		Body: strings.NewReader(bulkData.String()),
	}

	res, err = BulkRequest.Do(context.Background(), es)
	if err != nil {
		log.Fatalf("Error performing bulk request: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Fatalf("Error response from server: %s", res.String())
	}

	// Вывод результата
	io.Copy(os.Stdout, res.Body)
	fmt.Println("Data successfully indexed")
}
