package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth"
	"github.com/henriquedessen/goexpert/api/configs"
	_ "github.com/henriquedessen/goexpert/api/docs"
	"github.com/henriquedessen/goexpert/api/internal/entity"
	"github.com/henriquedessen/goexpert/api/internal/infra/database"
	"github.com/henriquedessen/goexpert/api/internal/infra/webserver/handlers"
	httpSwagger "github.com/swaggo/http-swagger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// @title           Go Expert API Example
// @version         1.0
// @description     Product API with auhtentication
// @termsOfService  http://swagger.io/terms/

// @contact.name   Henrique Dessen

// @host      localhost:8000
// @BasePath  /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	configs, err := configs.LoadConfig(".")
	if err != nil {
		panic(err)
	}
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&entity.Product{}, &entity.User{})
	ProductDB := database.NewProduct(db)
	ProductHandler := handlers.NewProductHandler(ProductDB)

	UserDB := database.NewUser(db)
	UserHandler := handlers.NewUserHandler(UserDB)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(LogRequest)
	r.Use(middleware.Recoverer)
	r.Use(middleware.WithValue("jwt", configs.TokenAut))
	r.Use(middleware.WithValue("jwtExperesIn", configs.JWTExperesIn))

	r.Route("/products", func(r chi.Router) {
		r.Use(jwtauth.Verifier(configs.TokenAut))
		r.Use(jwtauth.Authenticator)
		r.Post("/", ProductHandler.CreateProduct)
		r.Get("/", ProductHandler.GetProducts)
		r.Get("/{id}", ProductHandler.GetProduct)
		r.Put("/{id}", ProductHandler.UpdateProdut)
		r.Delete("/{id}", ProductHandler.DeleteProduct)
	})

	r.Post("/users", UserHandler.Create)
	r.Post("/users/generate_token", UserHandler.GetJWT)

	r.Get("/docs/*", httpSwagger.Handler(httpSwagger.URL("http://localhost:8000/docs/doc.json")))

	http.ListenAndServe(":8000", r)
}

func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
