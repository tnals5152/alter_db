package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type emailStruct struct {
	content     sql.NullString
	text     sql.NullString
	email sql.NullString
	id       sql.NullInt64
}

func main() {
	ctxBackground := context.Background()
	allPath := os.Args[1]//dbs path
	fmt.Println(allPath)
	if _, err := os.Stat(allPath); os.IsNotExist(err) {
		fmt.Println("path error : " + allPath)
		return
	}
	db, err := sql.Open("sqlite3", allPath)
	CheckError(err)
	defer db.Close()
	ctx, cancel := ContextTimeout(ctxBackground)
	defer cancel()

	stmt, err := db.PrepareContext(ctx, `
	SELECT *
	FROM test_db
	WHERE id IN (
		SELECT id FROM test_db 
		EXCEPT
		SELECT id FROM new_test_db
	)
	`)

	CheckError(err)
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx)
	CheckError(err)

	for rows.Next() {
		insertMail(db, rows, ctxBackground)
	}

}

func insertMail(db *sql.DB, rows *sql.Rows, ctxBackground context.Context) {
	var email emailStruct

	err := rows.Scan(&email.content, &email.text, &email.email, &email.id)
	CheckError(err)
	email.text.String = decoding(email.text.String)
	if email.text.String == "" {
		fmt.Println("decoding err text : ", email.text.String)
		return
	}
	ctx, cancel := ContextTimeout(ctxBackground)
	defer cancel()
	stmt, err := db.PrepareContext(ctx, `
	INSERT INTO new_test_db(content, text, email, id)
	VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		CheckError(err)
		fmt.Println("decoding err text : ", email.text.String)
		return
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, email.content.String, email.text.String, email.email.String, email.id.Int64)
	if err != nil {
		CheckError(err)
		fmt.Println("decoding err text : ", email.text.String)
	}

}

func decoding(str string) string {
	decodingStr, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		CheckError(err)
		return ""
	}

	return string(decodingStr)
}

func CheckError(err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		text := []interface{}{file, "line-" + strconv.Itoa(line), err.Error()}
		log.Println(text, err)
	}
}

func ContextTimeout(ctxBackground context.Context) (ctx context.Context, cancel context.CancelFunc) {
	ctx, cancel = context.WithTimeout(ctxBackground, time.Duration(3000000)*time.Second)
	return
}
