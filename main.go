package main

import (
	"Term-api/closer"
	"Term-api/config"
	middleware "Term-api/middleware/handlers"
	"Term-api/router"
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := runServer(ctx); err != nil {
		log.Fatal(err)
	}
}

func runServer(ctx context.Context) error {
	h := middleware.New(config.New().String())
	r := router.Router(h)
	c := &closer.Closer{}
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}
	c.Add(srv.Shutdown)
	c.Add(h.LogHand.Close)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen and serve: %v", err)
		}
	}()

	log.Printf("listening on %s", srv.Addr)
	<-ctx.Done()
	log.Println("shutting down server gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := c.Close(shutdownCtx); err != nil {
		return fmt.Errorf("Closer: %v", err)
	}
	return nil
}
