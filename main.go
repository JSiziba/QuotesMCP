package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const quotesURL = "https://api.api-ninjas.com/v2/quotes"

var httpClient = &http.Client{Timeout: 10 * time.Second}

type Quote struct {
	Quote      string   `json:"quote"`
	Author     string   `json:"author"`
	Work       string   `json:"work"`
	Categories []string `json:"categories"`
}

func fetchQuote(apiKey, category string) ([]Quote, error) {
	params := url.Values{}
	if category != "" {
		params.Set("categories", category)
	}

	req, err := http.NewRequest("GET", quotesURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Api-Key", apiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var quotes []Quote
	if err := json.Unmarshal(body, &quotes); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return quotes, nil
}

func buildMCPServer(apiKey string) *server.MCPServer {
	s := server.NewMCPServer(
		"quotes-server",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	quoteTool := mcp.NewTool("get_quote",
		mcp.WithDescription("Fetches a random quote from API Ninjas, optionally filtered by category"),
		mcp.WithString("category",
			mcp.Description("Category of quote to fetch, e.g. 'success', 'wisdom', 'happiness', 'love'. Leave empty for any category."),
		),
	)

	s.AddTool(quoteTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		category := req.GetString("category", "")

		quotes, err := fetchQuote(apiKey, category)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if len(quotes) == 0 {
			return mcp.NewToolResultError("no quotes returned"), nil
		}

		q := quotes[0]
		result := fmt.Sprintf("\"%s\"\n— %s", q.Quote, q.Author)
		if q.Work != "" {
			result += fmt.Sprintf(", %s", q.Work)
		}

		return mcp.NewToolResultText(result), nil
	})

	return s
}

func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("required environment variable %s is not set", key)
	}
	return val
}

func main() {
	_ = godotenv.Load()

	apiKey := mustEnv("API_NINJAS_KEY")

	mcpServer := buildMCPServer(apiKey)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	baseURL := mustEnv("BASE_URL")

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	})

	r.Get("/.well-known/oauth-authorization-server", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"issuer":                   baseURL,
			"authorization_endpoint":   baseURL + "/oauth/authorize",
			"token_endpoint":           baseURL + "/oauth/token",
			"response_types_supported": []string{"code"},
		})
	})

	r.Get("/.well-known/oauth-protected-resource", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"resource":              baseURL,
			"authorization_servers": []string{baseURL},
		})
	})

	streamServer := server.NewStreamableHTTPServer(mcpServer)
	r.Mount("/", streamServer)

	log.Printf("quotes MCP server running on :%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}
