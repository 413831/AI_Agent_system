package main

import (
	"ai-agent-system/cache"
	"ai-agent-system/graph"
	"ai-agent-system/service"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type GraphQLResponse struct {
	Data   interface{} `json:"data"`
	Errors []string     `json:"errors,omitempty"`
}

type GraphQLServer struct {
	resolver *graph.Resolver
	schema   *ast.Schema
}

func NewGraphQLServer(resolver *graph.Resolver) *GraphQLServer {
	// Parse schema
	schema := gqlparser.MustLoadSchema(&ast.Source{
		Input: `
		type Query {
			askAI(prompt: String!): AIResponse!
		}
		
		type AIResponse {
			prompt: String!
			result: String!
			cached: Boolean!
		}
		`,
	})

	return &GraphQLServer{
		resolver: resolver,
		schema:   schema,
	}
}

func (s *GraphQLServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req GraphQLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Simple query handler
	response := s.handleQuery(req)
	json.NewEncoder(w).Encode(response)
}

func (s *GraphQLServer) handleQuery(req GraphQLRequest) GraphQLResponse {
	// Debug log
	log.Printf("Received query: %s", req.Query)
	
	// Simple parsing for askAI query
	if req.Query == "" {
		return GraphQLResponse{Errors: []string{"Empty query"}}
	}

	// Debug conditions
	log.Printf("Contains askAI: %v", strings.Contains(req.Query, "askAI"))
	log.Printf("Contains prompt:: %v", strings.Contains(req.Query, "prompt:"))

	// Check if it's an askAI query (simple string matching for demo)
	if strings.Contains(req.Query, "askAI") && strings.Contains(req.Query, "prompt:") {
		log.Printf("Entering askAI condition")
		// Extract prompt from query (simple parsing)
		prompt := extractPrompt(req.Query)
		log.Printf("Extracted prompt: '%s'", prompt)
		if prompt == "" {
			return GraphQLResponse{Errors: []string{"Missing prompt parameter"}}
		}

		// Call resolver
		log.Printf("Calling resolver with prompt: '%s'", prompt)
		result, err := s.resolver.AskAI(context.Background(), prompt)
		log.Printf("Resolver returned result: %+v, error: %v", result, err)
		if err != nil {
			return GraphQLResponse{Errors: []string{err.Error()}}
		}

		log.Printf("Creating successful response")
		return GraphQLResponse{
			Data: map[string]interface{}{
				"askAI": map[string]interface{}{
					"prompt": result.Prompt,
					"result": result.Result,
					"cached": result.Cached,
				},
			},
		}
	}

	return GraphQLResponse{Errors: []string{"Unsupported query: " + req.Query}}
}

func extractPrompt(query string) string {
	// Look for prompt: "..." pattern with better parsing
	queryStr := strings.TrimSpace(query)
	log.Printf("Extracting prompt from: '%s'", queryStr)
	
	// Find prompt: pattern
	promptIndex := strings.Index(queryStr, "prompt:")
	log.Printf("Found 'prompt:' at index: %d", promptIndex)
	if promptIndex == -1 {
		log.Printf("No 'prompt:' found in query")
		return ""
	}
	
	// Start after "prompt:"
	start := promptIndex + 7
	for start < len(queryStr) && (queryStr[start] == ' ' || queryStr[start] == '\t') {
		start++
	}
	
	// Skip opening quote
	if start < len(queryStr) && (queryStr[start] == '"' || queryStr[start] == '\'') {
		start++
	}
	
	log.Printf("Start position: %d, char: '%c'", start, queryStr[start])
	
	// Find closing quote
	end := start
	for end < len(queryStr) && queryStr[end] != '"' && queryStr[end] != '\'' {
		end++
	}
	
	log.Printf("End position: %d, char: '%c'", end, queryStr[end])
	
	if start >= end || end > len(queryStr) {
		log.Printf("Invalid start/end positions")
		return ""
	}
	
	result := queryStr[start:end]
	log.Printf("Extracted result: '%s'", result)
	return result
}

func main() {
	redis := cache.NewRedis()
	ai := service.NewAIService()

	resolver := &graph.Resolver{
		Redis: redis,
		AI:    ai,
	}

	graphqlServer := NewGraphQLServer(resolver)

	// GraphQL endpoint
	http.Handle("/graphql", graphqlServer)

	// Simple health check
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok"}`)
	})

	// Simple playground
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>GraphQL Playground</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        textarea { width: 100%%; height: 100px; margin: 10px 0; }
        button { padding: 10px 20px; background: #0066cc; color: white; border: none; cursor: pointer; }
        pre { background: #f5f5f5; padding: 10px; border-radius: 4px; }
    </style>
</head>
<body>
    <h1>GraphQL Playground</h1>
    <div>
        <h3>Query:</h3>
        <textarea id="query">{ askAI(prompt: "Hola, cómo estás?") { prompt result cached } }</textarea>
        <button onclick="executeQuery()">Execute</button>
    </div>
    <div>
        <h3>Result:</h3>
        <pre id="result">Click execute to run query...</pre>
    </div>
    <script>
        function executeQuery() {
            const query = document.getElementById('query').value;
            fetch('/graphql', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ query: query })
            })
            .then(res => res.json())
            .then(data => {
                document.getElementById('result').textContent = JSON.stringify(data, null, 2);
            })
            .catch(err => {
                document.getElementById('result').textContent = 'Error: ' + err.message;
            });
        }
    </script>
</body>
</html>
		`)
	})

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
