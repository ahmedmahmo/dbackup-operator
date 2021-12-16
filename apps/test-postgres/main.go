package main

import (
  "database/sql"
  "fmt"

  _ "github.com/lib/pq"
)

const (
	host     = "postgres.postgres.svc.cluster.local"
	port     = 5432
	user     = "postgres"
	password = "1234"
	dbname   = "dvdrental"
  )

  func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
	  "password=%s dbname=%s sslmode=disable",
	  host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
	  panic(err)
	}
	defer db.Close()
  
	err = db.Ping()
	if err != nil {
	  panic(err)
	}
  
	fmt.Println("Successfully connected!")
  }