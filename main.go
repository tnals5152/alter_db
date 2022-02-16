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

type MailinfoStruct struct {
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
	//SELECT mi_type, mi_hash, mi_hashtext, mi_id
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
		insertMailinfo(db, rows, ctxBackground)
	}

}

func insertMailinfo(db *sql.DB, rows *sql.Rows, ctxBackground context.Context) {
	var mailinfo MailinfoStruct

	err := rows.Scan(&mailinfo.content, &mailinfo.text, &mailinfo.email, &mailinfo.id)
	CheckError(err)
	mailinfo.text.String = changeBlob(mailinfo.text.String)
	if mailinfo.text.String == "" {
		fmt.Println("decoding err text : ", mailinfo.text.String)
		return
	}
	ctx, cancel := ContextTimeout(ctxBackground)
	defer cancel()
	stmt, err := db.PrepareContext(ctx, `
	INSERT INTO new_mailinfo(content, text, email, id)
	VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		CheckError(err)
		fmt.Println("decoding err text : ", mailinfo.text.String)
		return
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, mailinfo.content.String, mailinfo.text.String, mailinfo.email.String, mailinfo.id.Int64)
	if err != nil {
		CheckError(err)
		fmt.Println("decoding err text : ", mailinfo.text.String)
	}

}

// 바이너리로 바꾸기 sqlite3.바이너리리리리리리리릴어카메라디ㅇㅗ
func changeBlob(str string) string {
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
