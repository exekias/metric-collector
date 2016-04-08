package accountname

import (
	"database/sql"
	"fmt"

	// PostgreSQL driver
	_ "github.com/lib/pq"

	"github.com/exekias/metric-collector/logging"
	"github.com/exekias/metric-collector/queue"
)

var log = logging.MustGetLogger("accountname")

// AccountName collects all the account names that sent metrics, with
// their first occurrence datetime (UTC) into PostgreSQL
type AccountName struct {
	db   *sql.DB
	stmt *sql.Stmt
}

// NewAccountName intializes and returns a new account name processor
func NewAccountName(url string) (*AccountName, error) {
	var processor AccountName

	log.Debug(fmt.Sprintf("Connecting to PostgreSQL (%s)", url))
	db, err := sql.Open("postgres", url)
	if err != nil {
		log.Error("Error connecting to PostgreSQL:", err)
		return nil, err
	}
	processor.db = db

	// Create table if not present
	log.Debug("Creating metrics table (if not exists)")
	_, err = processor.db.Exec(`
        CREATE TABLE IF NOT EXISTS metrics (
            username  varchar(100) CONSTRAINT firstkey PRIMARY KEY,
            time      timestamp without time zone default (now() at time zone 'utc')
        );
    `)
	if err != nil {
		log.Error("Could not create metrics table:", err)
		return nil, err
	}

	// Insert if new, do nothing if username is already inserted
	stmt, err := db.Prepare("INSERT INTO metrics (username) VALUES ($1) ON CONFLICT DO NOTHING")
	if err != nil {
		log.Error("Error preparing statement")
		return nil, err
	}
	processor.stmt = stmt

	return &processor, nil
}

// Process data from the queue
func (a AccountName) Process(d queue.MetricData) error {
	_, err := a.stmt.Exec(d.Username)
	if err != nil {
		log.Error("Error inserting user in the database:", err)
		return err
	}
	return nil
}
