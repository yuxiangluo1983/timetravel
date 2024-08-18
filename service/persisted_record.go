package service

import (
	"context"
	"errors"
	"database/sql"
	"encoding/json"
	"log"
	"os"

	"github.com/rainbowmga/timetravel/entity"
	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)

// PersistedRecordService is an implementation of RecordService based on sqlite for persistence.
type PersistedRecordService struct {
	sqlDb *sql.DB
}

func NewPersistedRecordService(dbName string) PersistedRecordService {
	_, err := os.Stat(dbName)
	if errors.Is(err, os.ErrNotExist) {
		log.Println("Creating " + dbName)
		file, err := os.Create(dbName) // Create SQLite file
		if err != nil {
			log.Fatal(err.Error())
		}
		file.Close()
	}

	sqliteDb, _ := sql.Open("sqlite3", dbName) // Open the created SQLite File

	if errors.Is(err, os.ErrNotExist) {
		createTableSQL := `CREATE TABLE records (
			"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,		
			"data" TEXT
		  );` // SQL Statement for Create Table

		log.Println("Create records table...")
		statement, err := sqliteDb.Prepare(createTableSQL) // Prepare SQL Statement
		if err != nil {
			log.Fatal(err.Error())
		}
		statement.Exec() // Execute SQL Statements
		log.Println("records table created")
	}

	return PersistedRecordService {
		sqlDb: sqliteDb,
	}
}

func (s * PersistedRecordService) Close() {
	s.sqlDb.Close()
}

func (s *PersistedRecordService) ReadRecord(id int) entity.Record {
	var jsonString string
	if err := s.sqlDb.QueryRow("SELECT data from records where id = ?", id).Scan(&jsonString); err == nil {
		var data = make(map[string]string)
		json.Unmarshal([]byte(jsonString), &data)
		return entity.Record{
			ID: id,
			Data: data,
		}
	}
	return entity.Record{}
}

func (s *PersistedRecordService) GetRecord(ctx context.Context, id int) (entity.Record, error) {
	record := s.ReadRecord(id)
	if record.ID == 0 {
		return entity.Record{}, ErrRecordDoesNotExist
	}

	record = record.Copy() // copy is necessary so modifations to the record don't change the stored record
	return record, nil
}

func (s *PersistedRecordService) CreateRecord(ctx context.Context, record entity.Record) error {
	id := record.ID
	if id <= 0 {
		return ErrRecordIDInvalid
	}

	existingRecord := s.ReadRecord(id)
	if existingRecord.ID != 0 {
		return ErrRecordAlreadyExists
	}

        jsonString, err := json.Marshal(record.Data)

	log.Println("Inserting record ...")
	insertSQL := `INSERT INTO records(id, data) VALUES (?, ?)`
	statement, err := s.sqlDb.Prepare(insertSQL) // Prepare statement. 
	if err != nil {
		log.Fatalln(err.Error())
	}
	_, err = statement.Exec(id, jsonString)
	if err != nil {
		log.Fatalln(err.Error())
	}

	return nil
}

func (s *PersistedRecordService) UpdateRecord(ctx context.Context, id int, updates map[string]*string) (entity.Record, error) {
	entry := s.ReadRecord(id)
	if entry.ID == 0 {
		return entity.Record{}, ErrRecordDoesNotExist
	}

	for key, value := range updates {
		if value == nil { // deletion update
			delete(entry.Data, key)
		} else {
			entry.Data[key] = *value
		}
	}

        jsonString, err := json.Marshal(entry.Data)

	log.Println("Updating record ...")
	updateSQL := `UPDATE records SET id=?, data=? WHERE id=?`
	statement, err := s.sqlDb.Prepare(updateSQL) // Prepare statement. 
	if err != nil {
		log.Fatalln(err.Error())
	}
	_, err = statement.Exec(id, jsonString, id)
	if err != nil {
		log.Fatalln(err.Error())
	}

	return entry.Copy(), nil
}
