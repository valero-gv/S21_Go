package main

import (
	"3/ex01/db"
	"3/ex01/db/types"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/elastic/go-elasticsearch/v8"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var store db.Store

func main() {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // Отключает проверку сертификата (не рекомендуется для production)
		},
	}

	// Настройка клиента Elasticsearch с использованием логина и пароля
	cfg := elasticsearch.Config{
		Addresses: []string{
			"https://localhost:9200",
		},
		Username:  "elastic",
		Password:  "GbU9-8IatX7etp6twbz1",
		Transport: tr,
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}
	store = db.NewElasticStore(client)
	http.HandleFunc("/", handleRequest)
	http.HandleFunc("/api/places", db.PlacesAPIHandler(store))
	http.Handle("/api/recommend", jwtMiddleware(handleRecommendPlaces(store)))
	http.HandleFunc("/api/get_token", getTokenHandler)
	fmt.Println("Server is running on port 8888")
	http.ListenAndServe(":8888", nil)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	prettyJSON, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		log.Printf("Error marshaling JSON response: %s", err)
		return
	}

	_, err = w.Write(prettyJSON)
	if err != nil {
		log.Printf("Error writing JSON response: %s", err)
	}
}

func handleRecommendPlaces(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		latStr := r.URL.Query().Get("lat")
		lonStr := r.URL.Query().Get("lon")

		lat, err := strconv.ParseFloat(latStr, 64)
		if err != nil {
			http.Error(w, "Invalid 'lat' parameter", http.StatusBadRequest)
			return
		}

		lon, err := strconv.ParseFloat(lonStr, 64)
		if err != nil {
			http.Error(w, "Invalid 'lon' parameter", http.StatusBadRequest)
			return
		}
		const limit = 3
		places, err := store.GetClosestPlaces(ctx, lat, lon, limit)
		respondWithJSON(w, http.StatusOK, places)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	if pageStr == "" {
		pageStr = "1"
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		http.Error(w, fmt.Sprintf("Invalid 'page' value: '%s'", pageStr), http.StatusBadRequest)
		return
	}

	limit := 10
	offset := (page - 1) * limit

	places, total, err := store.GetPlaces(limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	totalPages := (total + limit - 1) / limit

	if page > totalPages {
		http.Error(w, fmt.Sprintf("Invalid 'page' value: '%s'", pageStr), http.StatusBadRequest)
		return
	}

	tmpl := `
<!doctype html>
<html>
<head>
    <meta charset="utf-8">
    <title>Places</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
</head>
<body>
<h5>Total: {{.Total}}</h5>
<ul>
    {{range .Places}}
    <li>
        <div>{{.Name}}</div>
        <div>{{.Address}}</div>
        <div>{{.Phone}}</div>
    </li>
    {{end}}
</ul>
<div>
    {{if gt .Page 1}}
    <a href="/?page={{.PrevPage}}">Previous</a>
    {{end}}
    {{if lt .Page .TotalPages}}
    <a href="/?page={{.NextPage}}">Next</a>
    {{end}}
</div>
</body>
</html>`

	data := struct {
		Total      int
		Places     []types.Place
		Page       int
		PrevPage   int
		NextPage   int
		TotalPages int
	}{
		Total:      total,
		Places:     places,
		Page:       page,
		PrevPage:   page - 1,
		NextPage:   page + 1,
		TotalPages: totalPages,
	}

	t := template.Must(template.New("api/places").Parse(tmpl))
	t.Execute(w, data)
}

var jwtKey = []byte("my_secret_key")

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func getTokenHandler(w http.ResponseWriter, r *http.Request) {
	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &Claims{
		Username: "Username",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		//http.Error(w, err.Error(), http.StatusInternalServerError)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error in generating token"))
		return
	}

	response := struct {
		Token string `json:"token"`
	}{
		Token: tokenString,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func jwtMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil || !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), "username", claims.Username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
