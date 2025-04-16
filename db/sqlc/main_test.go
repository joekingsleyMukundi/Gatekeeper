package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

var testQueries *Queries
var testDb *sql.DB

func TestMain(m *testing.M) {
	dbDriver, dbSource := "postgres", "postgresql://root:secret@localhost:5432/gate_keeper?sslmode=disable"
	testDb, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatalf("Test DB conn error: %s", err)
	}
	testQueries = New(testDb)
	os.Exit(m.Run())
}
