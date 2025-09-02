package main

import (
	"jwt_go/config"
	"jwt_go/graph"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/rs/cors"
	gqlast "github.com/vektah/gqlparser/v2/ast"
)

const defaultPort = "8080"

func main() {
	// Load environment variables first
	if err := config.LoadEnv(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Initialize database
	db, err := config.InitDb()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Create GraphQL server with directives wired
	es := graph.NewExecutableSchema(graph.Config{
		Resolvers: &graph.Resolver{DB: db},
		Directives: graph.DirectiveRoot{
			Auth: graph.AuthDirective,
		},
	})
	srv := handler.New(es)

	// Configure transports
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})

	// Configure caching and extensions
	srv.SetQueryCache(lru.New[*gqlast.QueryDocument](1000))
	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	// CORS configuration
	corsOptions := cors.Options{
		AllowedOrigins:   strings.Split(os.Getenv("ALLOWED_ORIGINS"), ","),
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		Debug:            os.Getenv("ENV") == "development",
	}

	// Create router and setup routes
	mux := http.NewServeMux()
	mux.Handle("/", playground.Handler("GraphQL playground", "/query"))

	// Apply auth middleware to the GraphQL endpoint
	graphQLHandler := graph.AuthMiddleware(srv)
	mux.Handle("/query", graphQLHandler)

	// Serve static files
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Start server with CORS
	log.Printf("Server starting on http://localhost:%s (environment: %s)", port, os.Getenv("ENV"))
	log.Fatal(http.ListenAndServe(":"+port, cors.New(corsOptions).Handler(mux)))
}
