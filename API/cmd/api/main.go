package main

import (
    "fmt"
    "os"
    "time"
    "lawmap/internal/app"
)

func main() {
    a, err := app.New()
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
