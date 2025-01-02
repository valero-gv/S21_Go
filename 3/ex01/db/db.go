package db

import (
	"3/ex01/db/types"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Store interface {
	GetPlaces(limit int, offset int) ([]types.Place, int, error)
	GetClosestPlaces(ctx context.Context, lat float64, lon float64, limit int) ([]types.Place, error)
}

type ElasticStore struct {
	client *elasticsearch.Client
}

func NewElasticStore(client *elasticsearch.Client) *ElasticStore {
	return &ElasticStore{client: client}
}

func (es *ElasticStore) GetPlaces(limit int, offset int) ([]types.Place, int, error) {
	query := `{
        "from": ` + strconv.Itoa(offset) + `,
        "size": ` + strconv.Itoa(limit) + `,
        "query": {
            "match_all": {}
        }
    }`

	req := esapi.SearchRequest{
		Index: []string{"places"},
		Body:  strings.NewReader(query),
	}

	res, err := req.Do(context.Background(), es.client)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, 0, errors.New(res.String())
	}

	var searchResult struct {
		Hits struct {
			Total struct {
				Value int `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source types.Place `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&searchResult); err != nil {
		return nil, 0, err
	}

	places := make([]types.Place, 0)
	for _, hit := range searchResult.Hits.Hits {
		places = append(places, hit.Source)
	}

	return places, searchResult.Hits.Total.Value, nil
}

func (es *ElasticStore) GetClosestPlaces(ctx context.Context, lat float64, lon float64, limit int) ([]types.Place, error) {
	var buf bytes.Buffer
	query := map[string]interface{}{
		"sort": []map[string]interface{}{
			{
				"_geo_distance": map[string]interface{}{
					"location": map[string]float64{
						"lat": lat,
						"lon": lon,
					},
					"order":           "asc",
					"unit":            "km",
					"mode":            "min",
					"distance_type":   "arc",
					"ignore_unmapped": true,
				},
			},
		},
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
		"size": 3,
	}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Printf("Error encoding query: %s", err)
		return nil, err
	}

	res, err := es.client.Search(
		es.client.Search.WithContext(ctx),
		es.client.Search.WithIndex("places"),
		es.client.Search.WithBody(&buf),
	)
	if err != nil {
		log.Printf("Error getting response: %s", err)
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			log.Printf("Error parsing the response body: %s", err)
			return nil, err
		}
		errMsg := fmt.Sprintf("[%s] %s: %s", res.Status(), e["error"].(map[string]interface{})["type"], e["error"].(map[string]interface{})["reason"])
		log.Print(errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	var r struct {
		Hits struct {
			Hits []struct {
				Source types.Place `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Printf("Error parsing the response body: %s", err)
		return nil, err
	}

	places := make([]types.Place, len(r.Hits.Hits))
	for i, hit := range r.Hits.Hits {
		places[i] = hit.Source
	}

	return places, nil
}

func PlacesAPIHandler(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pageParam := r.URL.Query().Get("page")
		if pageParam == "" {
			pageParam = "1"
		}
		page, err := strconv.Atoi(pageParam)
		if err != nil || page < 1 {
			http.Error(w, fmt.Sprintf(`{"error": "Invalid 'page' value: '%s'"}`, pageParam), http.StatusBadRequest)
			return
		}
		limit := 20
		offset := (page - 1) * limit
		places, total, err := store.GetPlaces(limit, offset)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		lastPage := (total + limit - 1) / limit
		if page > lastPage {
			http.Error(w, fmt.Sprintf(`{"error": "Invalid 'page' value: '%s'"}`, pageParam), http.StatusBadRequest)
			return
		}

		response := struct {
			Name     string        `json:"name"`
			Total    int           `json:"total"`
			Places   []types.Place `json:"places"`
			PrevPage int           `json:"prev_page,omitempty"`
			NextPage int           `json:"next_page,omitempty"`
			LastPage int           `json:"last_page"`
		}{
			Name:     "Places",
			Total:    total,
			Places:   places,
			LastPage: lastPage,
		}

		if page > 1 {
			response.PrevPage = page - 1
		}
		if page < lastPage {
			response.NextPage = page + 1
		}

		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	}
}
