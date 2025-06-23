package internal

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

type TransientSessionStorage struct {
	db    *sql.DB
	encde *Encde
	quit  chan int
}

func NewTransientSessionStorage(encde *Encde) (*TransientSessionStorage, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}
	quit := make(chan int)
	tss := &TransientSessionStorage{
		db,
		encde,
		quit,
	}

	return tss, nil
}

func (tss *TransientSessionStorage) Initialize() error {
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS session_handover (
    	token TEXT NOT NULL PRIMARY KEY,
    	cookie TEXT NOT NULL,
    	expires_at INTEGER NOT NULL
	);
	`
	_, err := tss.db.Exec(sqlStmt)

	if err != nil {
		return err
	}

	go func() {
		tick := time.Tick(30 * time.Second)

		for {
			select {
			case <-tick:
				_, err := tss.db.Exec(`
				DELETE FROM session_handover
				WHERE expires_at <= strftime('%s', 'now');
				`)
				if err != nil {
					log.Println(err)
				}
			case <-tss.quit:
				return
				// default:
				// 	rows, err := tss.db.Query(`
				// 	SELECT * FROM session_handover;
				// 	`)
				// 	if err != nil {
				// 		log.Println(err)
				// 	}
				// 	log.Println("| token                            | cookie                                                                                                                                                                                                                                                   | expires_at |")
				// 	log.Println("| -------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------- |")
				// 	for rows.Next() {
				// 		var token string
				// 		var cookie string
				// 		var expires_at int
				// 		err = rows.Scan(&token, &cookie, &expires_at)
				// 		if err != nil {
				// 			log.Println(err)
				// 		}
				// 		log.Printf("| %s | %s | %d |\n", token, cookie, expires_at)
				// 	}
				// 	err = rows.Err()
				// 	if err != nil {
				// 		log.Println(err)
				// 	}
				// 	rows.Close()
				// 	time.Sleep(5 * time.Second)
			}
		}
	}()

	return nil
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
	tss.quit <- 0

	return tss.db.Close()
}
