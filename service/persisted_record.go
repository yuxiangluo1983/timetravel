package service

import (
	"context"
	"errors"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/yuxiangluo1983/timetravel/entity"
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
			"id" integer NOT NULL,		
			"version" integer NOT NULL,
			"data" TEXT,
			PRIMARY KEY ("id", "version")
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
	var version string
	if err := s.sqlDb.QueryRow("SELECT MAX(version) from records where id = ? GROUP BY id", id).Scan(&version); err != nil {
		return entity.Record{}
	}
	verNumber, _ := strconv.ParseInt(version, 10, 64)
	return s.ReadRecordWithVersion(id, verNumber)
}

func (s *PersistedRecordService) ReadRecordWithVersion(id int, version int64) entity.Record {
	var jsonString string
	if err := s.sqlDb.QueryRow("SELECT data from records where id = ? and version = ?", id, version).Scan(&jsonString); err == nil {
		var data = make(map[string]string)
		json.Unmarshal([]byte(jsonString), &data)
		return entity.Record{
			ID: id,
			Ver: version,
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

func (s *PersistedRecordService) GetAllRecords(ctx context.Context, id int) []entity.Record {
	row, err := s.sqlDb.Query("SELECT * FROM records where id = ? ORDER BY version", id)
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()
	
	var records []entity.Record

	for row.Next() { // Iterate and fetch the records from result cursor
		var id int
		var version string
		var jsonString string
		row.Scan(&id, &version, &jsonString)

		ver, _ := strconv.ParseInt(version, 10, 64)
		data := make(map[string]string)
		json.Unmarshal([]byte(jsonString), &data)
		records = append(records, entity.Record{
			ID: id,
			Ver: ver,
			Data: data,
		})
	}

	return records
}

func (s *PersistedRecordService) GetRecordsWithVersions(ctx context.Context, id int, versions []int64) []entity.Record {
	var records []entity.Record
	
	for _, v := range versions {
		records = append(records, s.ReadRecordWithVersion(id, v))
	}

	return records
}

func (s *PersistedRecordService) CreateRecord(ctx context.Context, record entity.Record) error {
	id := record.ID
	if id <= 0 {
		return ErrRecordIDInvalid
	}

        jsonString, err := json.Marshal(record.Data)
	version := time.Now().Unix()

	log.Println("Inserting record ...")
	insertSQL := `INSERT INTO records(id, version, data) VALUES (?, ?, ?)`
	statement, err := s.sqlDb.Prepare(insertSQL) // Prepare statement. 
	if err != nil {
		log.Fatalln(err.Error())
	}
	_, err = statement.Exec(id, version, jsonString)
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
	version := time.Now().Unix()

	log.Println("Updating record ...")
	updateSQL := `UPDATE records SET id=?, version=?, data=? WHERE id=? and version=?`
	statement, err := s.sqlDb.Prepare(updateSQL) // Prepare statement. 
	if err != nil {
		log.Fatalln(err.Error())
	}
	_, err = statement.Exec(id, version, jsonString, id, entry.Ver)
	if err != nil {
		log.Fatalln(err.Error())
	}

	return entry.Copy(), nil
}
