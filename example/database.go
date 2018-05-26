package main

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/greencase/go-gdpr"
)

// dbState is an intermediate type for
// recording request data in our db.
type dbState struct {
	SubjectRequestId       string
	RequestStatus          string
	EncodedRequest         string
	StatusCallbackUrls     []string
	SubmittedTime          time.Time
	ReceivedTime           time.Time
	ExpectedCompletionTime time.Time
}

type Database struct {
	db   *sql.DB
	path string
}

func NewDatabase(path string) (*Database, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	// Allows concurrent access to SQLite
	db.SetMaxOpenConns(1)
	err = db.Ping()
	return &Database{db: db, path: path}, err
}

func (d *Database) Migrate() error {
	stmt := `
    CREATE TABLE request (
		subject_request_id TEXT PRIMARY KEY UNIQUE,
		status TEXT,
		encoded_request TEXT,
		submitted_time DATETIME,
		received_time DATETIME,
		expected_completion_time DATETIME
    );
	CREATE TABLE callback (
		subject_request_id TEXT,
		callback_url TEXT
	)
    `
	_, err := d.db.Exec(stmt)
	return err
}

func (d *Database) callbackUrls(tx *sql.Tx, id string) ([]string, error) {
	var callbacks []string
	rows, err := tx.Query("SELECT (callback_url) FROM callback WHERE subject_request_id = $1", id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var url string
		err = rows.Scan(&url)
		if err != nil {
			return nil, err
		}
		callbacks = append(callbacks, url)
	}
	return callbacks, nil
}

func (d *Database) Read(id string) (*dbState, error) {
	row := d.db.QueryRow("SELECT * FROM request WHERE subject_request_id = $1", id)
	req := &dbState{}
	err := row.Scan(&req.SubjectRequestId, &req.RequestStatus, &req.EncodedRequest,
		&req.SubmittedTime, &req.ReceivedTime, &req.ExpectedCompletionTime)
	return req, err
}

func (d *Database) Pending() ([]*dbState, error) {
	var requests []*dbState
	tx, err := d.db.Begin()
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query("SELECT * FROM request WHERE status = 'pending'")
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	for rows.Next() {
		req := &dbState{}
		err := rows.Scan(&req.SubjectRequestId, &req.RequestStatus, &req.EncodedRequest,
			&req.SubmittedTime, &req.ReceivedTime, &req.ExpectedCompletionTime)
		if err != nil {
			return nil, err
		}
		urls, err := d.callbackUrls(tx, req.SubjectRequestId)
		if err != nil {
			return nil, err
		}
		req.StatusCallbackUrls = urls
		requests = append(requests, req)
	}
	return requests, tx.Commit()
}

func (d *Database) Write(req dbState) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(
		"INSERT INTO request (subject_request_id, status, encoded_request, submitted_time, received_time, expected_completion_time) VALUES ($1,$2,$3,$4,$5,$6)",
		req.SubjectRequestId, req.RequestStatus, req.EncodedRequest,
		req.SubmittedTime, req.ReceivedTime, req.ExpectedCompletionTime,
	)
	if err != nil {
		tx.Rollback()
		return err
	}
	for _, url := range req.StatusCallbackUrls {
		_, err = tx.Exec("INSERT INTO callback (subject_request_id, callback_url) VALUES ($1,$2)", req.SubjectRequestId, url)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (d *Database) SetStatus(id string, status gdpr.RequestStatus) error {
	_, err := d.db.Exec("UPDATE request SET status = $1 WHERE subject_request_id = $2", status, id)
	return err
}
