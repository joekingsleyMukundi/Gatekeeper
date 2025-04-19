package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/joekingsleyMukundi/Gatekeeper/utils"
	_ "github.com/lib/pq"
)

var testQueries *Queries
var testDb *sql.DB

func TestMain(m *testing.M) {
	config, err := utils.LoadConfig("../..")
	if err != nil {
		log.Fatalf("ERROR: cannot load config: %s", err)
	}
	dbDriver, dbSource := config.DBdriver, config.DBsource
	testDb, err = sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatalf("Test DB conn error: %s", err)
	}
	testQueries = New(testDb)
	os.Exit(m.Run())
}
