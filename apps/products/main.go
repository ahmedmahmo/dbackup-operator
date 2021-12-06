package main

import (
	"log"
	"net/http"
	"os"
	"github.com/ahmedmahmo/products/handlers"
	"time"
	"os/signal"
	"context"
)
func main()  {
	// register logging with name product-api
	l := log.New(os.Stdout, "product-api", log.LstdFlags)
	l.Println("Starting server on port 8080")

	// Handle Product request
	ph := handlers.NewProduct(l)

	// ServeMux interface for function registration in HTTP 
	sm := http.NewServeMux()
	sm.Handle("/", ph)

	// Custom Server with Timeout
	addr := ":8080"
	server := &http.Server{
		Addr: addr,
		Handler: sm,
		IdleTimeout: 120*time.Second,
		ReadTimeout: 1*time.Second,
		WriteTimeout: 1*time.Second,
	}

	// Async Server Start
	go func(){
		err := server.ListenAndServe()
		if err != nil{
			l.Fatal(err)
		}
	}()
	
	// Create signal for interupting
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt)
	signal.Notify(sigChan, os.Kill)

	sig := <- sigChan
	l.Println("Recived terminate, graceful shutdown", sig)

	tc, _ := context.WithTimeout(context.Background(), 30*time.Second)
	server.Shutdown(tc)

}