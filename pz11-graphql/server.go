package main

import (
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/CyberGeo335/pz11-graphql/graph"
	"github.com/CyberGeo335/pz11-graphql/graph/generated"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	resolver := graph.NewResolver()
	server := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))

	http.Handle("/", playground.Handler("Practical 11 GraphQL Playground", "/query"))
	http.Handle("/query", server)

	log.Printf("GraphQL Playground: http://localhost:%s/", port)
	log.Printf("GraphQL endpoint:   http://localhost:%s/query", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
