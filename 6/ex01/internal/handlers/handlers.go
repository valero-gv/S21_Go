package handlers

import (
	"6/ex01/internal/db"
	"6/ex01/internal/middleware"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/russross/blackfriday/v2"
	"html/template"
	"net/http"
	"strconv"
	"time"
)

func dec(x int) int {
	return x - 1
}

func inc(x int) int {
	return x + 1
}

type PageData struct {
	Articles    []db.Article
	CurrentPage int
	TotalPages  int
}

func HandleHome(w http.ResponseWriter, r *http.Request) {
	// Определение текущей страницы
	page := 1
	if r.URL.Query().Get("page") != "" {
		page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	}

	// Лимит на количество статей на одной странице
	limit := 3
	offset := (page - 1) * limit

	var articles []db.Article
	db.Db.Limit(limit).Offset(offset).Find(&articles)

	var totalArticles int64
	db.Db.Model(&db.Article{}).Count(&totalArticles)
	totalPages := int((totalArticles + int64(limit) - 1) / int64(limit))

	// Передаем данные в шаблон
	pageData := PageData{
		Articles:    articles,
		CurrentPage: page,
		TotalPages:  totalPages,
	}

	fmt.Println(articles)

	tmpl := template.Must(template.New("home.html").Funcs(template.FuncMap{
		"dec": dec,
		"inc": inc,
		"slice": func(s string, start, end int) string {
			if start > len(s) {
				return ""
			}
			if end > len(s) {
				end = len(s)
			}
			return s[start:end]
		},
	}).ParseFiles("/home/valero/Desktop/s21_Go/6/ex01/templates/home.html"))
	tmpl.Execute(w, pageData)
}

func HandleArticle(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/article/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}

	var article db.Article
	if err := db.Db.First(&article, id).Error; err != nil {
		http.Error(w, "Article not found", http.StatusNotFound)
		return
	}

	htmlContent := blackfriday.Run([]byte(article.Content))

	tmpl := template.Must(template.ParseFiles("/home/valero/Desktop/s21_Go/6/ex01/templates/article.html"))
	tmpl.Execute(w, map[string]interface{}{
		"Title":   article.Title,
		"Content": template.HTML(htmlContent),
	})
}

func HandleAdmin(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	tokenString := cookie.Value
	claims := &jwt.MapClaims{}

	// Парсим и проверяем токен
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("your-secret-key"), nil
	})

	if err != nil || !token.Valid {
		// Если токен недействителен или не удалось его распарсить
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == "GET" {
		tmpl := template.Must(template.ParseFiles("/home/valero/Desktop/s21_Go/6/ex01/templates/admin.html"))
		tmpl.Execute(w, nil)
		return
	}
}

func HandlePostArticle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")

	if title == "" || content == "" {
		http.Error(w, "Title or content cannot be empty", http.StatusBadRequest)
		return
	}

	newArticle := db.Article{
		Title:   title,
		Content: content,
	}
	result := db.Db.Create(&newArticle)
	if result.Error != nil {
		http.Error(w, "Error saving article", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		tmpl := template.Must(template.ParseFiles("/home/valero/Desktop/s21_Go/6/ex01/templates/login.html"))
		tmpl.Execute(w, nil)
	case "POST":
		r.ParseForm()
		username := r.FormValue("username")
		password := r.FormValue("password")

		if username == db.Creds.AdminUsername && password == db.Creds.AdminPassword {
			token, err := middleware.GenerateJWT(username) // Генерация JWT токена
			if err != nil {
				http.Error(w, "Failed to generate token", http.StatusInternalServerError)
				return
			}

			// Устанавливаем cookie для сессии
			http.SetCookie(w, &http.Cookie{
				Name:     "session_token",
				Value:    token,
				Expires:  time.Now().Add(1 * time.Hour), // Токен действует 1 час
				HttpOnly: true,
			})
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func HandleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "session_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
