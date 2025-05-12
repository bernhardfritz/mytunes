package internal

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type TransientSessionStorage struct {
	db    *sql.DB
	encde *Encde
}

func NewTransientSessionStorage(encde *Encde) (*TransientSessionStorage, error) {
	db, err := sql.Open("sqlite3", "/var/lib/mytunes/mytunes.db")
	if err != nil {
		return nil, err
	}
	tss := &TransientSessionStorage{
		db,
		encde,
	}

	err = tss.initialize()
	if err != nil {
		return nil, err
	}

	return tss, nil
}

func (tss *TransientSessionStorage) initialize() error {
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS session_handover (
    	token TEXT NOT NULL PRIMARY KEY,
    	cookie TEXT NOT NULL,
    	expires_at INTEGER NOT NULL
	);
	`
	_, err := tss.db.Exec(sqlStmt)

	return err
}

func generate_token() (string, error) {
	token := make([]byte, 16)
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", token), nil
}

func (tss *TransientSessionStorage) StoreCookie(cookie string) (string, error) {
	stmt, err := tss.db.Prepare(`
	INSERT INTO session_handover(token, cookie, expires_at)
	VALUES (?, ?, ?);
	`)
	if err != nil {
		return "", err
	}
	defer stmt.Close()
	token, err := generate_token()
	if err != nil {
		return "", err
	}
	encryptedCookie, err := tss.encde.Encrypt(cookie)
	if err != nil {
		return "", err
	}
	expiresAt := time.Now().Local().Add(time.Second * time.Duration(30)).Unix()
	_, err = stmt.Exec(token, encryptedCookie, expiresAt)

	return token, err
}

func (tss *TransientSessionStorage) FindCookie(token string) (string, error) {
	stmt, err := tss.db.Prepare(`
	SELECT cookie, expires_at
	FROM session_handover
	WHERE token = ?;
	`)
	if err != nil {
		return "", err
	}
	defer stmt.Close()
	var encryptedCookie string
	var expiresAt int64
	err = stmt.QueryRow(token).Scan(&encryptedCookie, &expiresAt)
	if err != nil {
		return "", err
	}
	if time.Now().Local().After(time.Unix(expiresAt, 0)) {
		return "", err
	}
	decryptedCookie, err := tss.encde.Decrypt(encryptedCookie)
	if err != nil {
		return "", err
	}

	return decryptedCookie, nil
}

func (tss *TransientSessionStorage) DeleteCookie(token string) error {
	stmt, err := tss.db.Prepare(`
	DELETE FROM session_handover
	WHERE token = ?;
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(token)

	return err
}

func (tss *TransientSessionStorage) Close() error {
	return tss.db.Close()
}
