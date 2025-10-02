package app

import (
    "fmt"
    "os"
    httpapi "lawmap/internal/http"
    graphrepo "lawmap/internal/repo/graph"
    conf "lawmap/internal/config"
)

type App struct {
    Server *httpapi.Server
}

func New() (*App, error) {
    store := graphrepo.NewMemoryStore()
    // Load example data by default to make the API immediately useful.
    examples := os.Getenv("EXAMPLES_FILE")
    if examples == "" {
        // Assume the working dir is API/; if not, allow override via EXAMPLES_FILE
        examples = "docs/EXAMPLES.graph.jsonl"
    }
    if err := store.LoadJSONL(examples); err != nil {
        return nil, fmt.Errorf("load examples: %w", err)
    }
    // Load sources config if available
    var sources []conf.SourceDescriptor
    spath := os.Getenv("SOURCES_FILE")
    if spath == "" {
        // prefer project config if present
        if _, err := os.Stat("configs/sources.json"); err == nil {
            spath = "configs/sources.json"
        } else if _, err := os.Stat("configs/sources.example.json"); err == nil {
            spath = "configs/sources.example.json"
        }
    }
    if spath != "" {
        if ss, err := conf.LoadSources(spath); err == nil {
            sources = ss
            fmt.Printf("Loaded %d sources from %s\n", len(sources), spath)
        } else {
            fmt.Printf("warn: could not load sources from %s: %v\n", spath, err)
        }
    }
    server := httpapi.NewServer(store, sources)
    return &App{Server: server}, nil
}

func (a *App) Start() error {
    return a.Server.Start()
}
