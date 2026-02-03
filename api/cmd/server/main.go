package main

import (
	"ai-agent-system/cache"
	"ai-agent-system/graph"
	"ai-agent-system/handler"
	"ai-agent-system/service"

	"log"
	"net/http"
)

func main() {
	redis := cache.NewRedis()
	ai := service.NewAIService()

	srv := handler.NewDefaultServer(
		graph.NewExecutableSchema(graph.Config{
			Resolvers: &graph.Resolver{
				Redis: redis,
				AI:    ai,
			},
		}),
	)

	http.Handle("/graphql", srv)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
