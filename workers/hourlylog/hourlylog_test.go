package hourlylog

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"gopkg.in/mgo.v2/dbtest"

	"github.com/exekias/metric-collector/queue"
)

var server dbtest.DBServer

func TestMain(m *testing.M) {
	if _, err := exec.LookPath("mongod"); err != nil {
		fmt.Println("mongod not in $PATH, skipping this test")
		return
	}

	// Get a temp dir
	dir, err := ioutil.TempDir("", "mongo")
	if err != nil {
		panic(err)
	}
	server.SetPath(dir)

	// clean up
	defer func() {
		server.Stop()
		os.RemoveAll(dir)
	}()

	flag.Parse()
	os.Exit(m.Run())
}

func TestHourlyLog(t *testing.T) {
	session := server.Session()
	defer session.Close()
	processor, err := NewHourlyLog(session, "test", "test")
	if err != nil {
		t.Error(err)
	}

	if err = processor.Process(queue.MetricData{"user1", 0, "metric"}); err != nil {
		t.Error("Processing a metric", err)
	}

	if err = processor.Process(queue.MetricData{"user2", 1, "metric"}); err != nil {
		t.Error("Processing a metric", err)
	}

	count, err := session.DB("test").C("test").Count()
	if err != nil {
		t.Error("Getting collection count", err)
	}

	if count != 2 {
		t.Errorf("Metrics uncorrectly inserted, expected 2, got %d", count)
	}
}
