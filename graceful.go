package golek

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// listenAndServeGraceful starts an HTTP server and handles graceful shutdown.
func listenAndServeGraceful(addr string, handler http.Handler, tlsConfig ...*tls.Config) error {
	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	if len(tlsConfig) > 0 && tlsConfig[0] != nil {
		server.TLSConfig = tlsConfig[0]
	}

	errChan := make(chan error, 1)
	go func() {
		log.Printf("Server starting on %s\n", addr)
		var err error
		if server.TLSConfig != nil {
			// Certificates should be embedded in TLSConfig or provided
			err = server.ListenAndServeTLS("", "")
		} else {
			err = server.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errChan:
		return err
	case <-quit:
		log.Println("Shutting down gracefully...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			return err
		}
		return nil
	}
}

// ListenAndServeListenAndServe starts an HTTP server with graceful shutdown.
func (e *Engine) ListenAndServe(addr string) error {
	return listenAndServeGraceful(addr, e)
}

// ListenAndServeTLS starts an HTTPS server with graceful shutdown.
func (e *Engine) ListenAndServeTLS(addr, certFile, keyFile string) error {
	// For simplicity, LoadX509KeyPair and pass TLSConfig
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	return listenAndServeGraceful(addr, e, tlsConfig)
}

// ListenAndServe starts an HTTP server with graceful shutdown (package-level).
func ListenAndServe(addr string, handler http.Handler) error {
	return listenAndServeGraceful(addr, handler)
}
