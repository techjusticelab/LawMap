package main

import (
    "fmt"
    "os"
    "time"
    "lawmap/internal/app"

    "github.com/joho/godotenv"
)

func main() {
    // Load .env file if it exists (ignore errors - file is optional)
    _ = godotenv.Load()

    var a *app.App
    var err error

    // Check if DATABASE_URL is set (for PostgreSQL/Neon)
    if os.Getenv("DATABASE_URL") != "" {
        fmt.Println("Using PostgreSQL backend (DATABASE_URL detected)")
        a, err = app.NewWithPostgres()
    } else {
        fmt.Println("Using in-memory backend (no DATABASE_URL found)")
        a, err = app.New()
    }

    if err != nil {
        fmt.Fprintf(os.Stderr, "startup error: %v\n", err)
        os.Exit(1)
    }

    // Small banner with timestamp so logs show when the process started.
    fmt.Printf("LawMap API starting at %s\n", time.Now().Format(time.RFC3339))

    if err := a.Start(); err != nil {
        fmt.Fprintf(os.Stderr, "server error: %v\n", err)
        os.Exit(1)
    }
}
