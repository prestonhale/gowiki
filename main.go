package main

import (
	"net/http"
	"log"
	"strings"
	"fmt"
	"database/sql"
	"io/ioutil"
	_ "github.com/mattn/go-sqlite3"
)

var (
	db *sql.DB
)

func logFatalIfErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func endpoint(w http.ResponseWriter, r *http.Request) {
	page := strings.TrimLeft(r.URL.Path, "/")
	if r.Method == "POST" {
		content, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
			return
		}
		pageContent := string(content)
		fmt.Println(pageContent)
		postPage(page, pageContent)
		w.Write([]byte(pageContent))
	} else {
		pageContent, pageFound := getPage(page)
		if pageFound == false {
			w.Write([]byte("Not found."))
		}
		w.Write([]byte(pageContent))
	}
}

func postPage(name string, content string) {
	tx, err := db.Begin()
	logFatalIfErr(err)

	stmt, err := tx.Prepare(`
		INSERT INTO pages (name, content) VALUES(?, ?)
		ON CONFLICT(name) DO UPDATE SET content=?
	`)
	logFatalIfErr(err)
	defer stmt.Close()

	_, err = stmt.Exec(name, content, content)
	logFatalIfErr(err)
	tx.Commit()

}

func getPage(name string) (string, bool) {
	stmt, err := db.Prepare("select content from pages where name=?")
	logFatalIfErr(err)
	defer stmt.Close()

	var content string
	err = stmt.QueryRow(name).Scan(&content)
	if err == sql.ErrNoRows {
		return "", false
	}
	logFatalIfErr(err)
	return content, true
}

func main() {
	dbRef, err := sql.Open("sqlite3", "./wikiDB.db")
	db = dbRef
	logFatalIfErr(err)
	defer db.Close()

	http.HandleFunc("/", endpoint)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}