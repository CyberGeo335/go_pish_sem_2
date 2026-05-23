package graph

//go:generate go run github.com/99designs/gqlgen@v0.17.49 generate

import (
	"github.com/CyberGeo335/pz11-graphql/internal/repository"
	"github.com/CyberGeo335/pz11-graphql/internal/service"
)

// Resolver is the root dependency container for GraphQL resolvers.
// It is initialized once in server.go.
type Resolver struct {
	TaskService *service.TaskService
}

func NewResolver() *Resolver {
	taskRepo := repository.NewMemoryTaskRepository()
	return &Resolver{
		TaskService: service.NewTaskService(taskRepo),
	}
}
