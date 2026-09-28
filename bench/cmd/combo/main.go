// Command combo measures the combined cost of the two heaviest planned
// dependencies: the shoutrrr fork (notifications) and modernc.org/sqlite
// (pure-Go storage). It links both and does nothing else.
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/nicholas-fedor/shoutrrr"
	_ "modernc.org/sqlite"
)

func main() {
	dbPath := flag.String("db", "", "path to a sqlite file (default: temp dir)")
	addr := flag.String("addr", "127.0.0.1:18080", "listen address")
	flag.Parse()

	path := *dbPath
	if path == "" {
		path = os.TempDir() + "/octopulse-probe.db"
	}

	db, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		log.Fatalf("sqlite open: %v", err)
	}
	defer db.Close()
	if _, err := db.Exec("create table if not exists probe (id integer primary key, v text)"); err != nil {
		log.Fatalf("sqlite exec: %v", err)
	}

	if _, err := shoutrrr.CreateSender(); err != nil {
		log.Fatalf("shoutrrr: %v", err)
	}
	fmt.Println("combo probe ready: shoutrrr service map + sqlite loaded")
	log.Fatal(http.ListenAndServe(*addr, nil))
}
