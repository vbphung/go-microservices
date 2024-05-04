package main

import (
	"authentication/data"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgconn"

	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const webPort = "80"

var count int64

type Config struct {
	DB     *sql.DB
	Models data.Models
}

func main() {
	log.Println("Start authentication service")

	conn := connectToDB()
	if conn == nil {
		log.Panic("Cannot connect to DB")
	}

	app := Config{
		DB:     conn,
		Models: data.New(conn),
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}

}

func opneDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func connectToDB() *sql.DB {
	dsn := os.Getenv("DSN")
	for {
		conn, err := opneDB(dsn)
		if err != nil {
			log.Println("Potgres is not ready yet")
			count++
		} else {
			log.Println("Connected to Potgres")
			return conn
		}

		if count > 0 {
			log.Println(err)
			return nil
		}

		log.Println("Back off for 2 seconds")
		time.Sleep(2 * time.Second)
		continue
	}
}
