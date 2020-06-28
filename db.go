package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

///////////////////////////
// META AND DB FUNCTIONS //
///////////////////////////

func openDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./cody.db")
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func resetDB() {
	os.Remove("./hackernet.db")
	rawExecDB(`DROP TABLE IF EXISTS events;`)
	// distributed comps (cyber conquest, ccdc) with multiple vapp isnstances use multiple ids and combine /display separately in the frontend
	rawExecDB(`CREATE TABLE events (
                id          INTEGER PRIMARY KEY AUTOINCREMENT,
                type        INTEGER,
                kind        VARCHAR(255),
                title       VARCHAR(255),
                vapp        VARCHAR(255),
                field1      VARCHAR(2048),
                field2      VARCHAR(2048),
                field3      VARCHAR(2048),
                switch      BOOL
            );`)
	rawExecDB(`CREATE TABLE scores (
                id          INTEGER PRIMARY KEY AUTOINCREMENT,
                time        DATETIME DEFAULT CURRENT_TIMESTAMP,
                event       VARCHAR(255),
                user        VARCHAR(255),
                points      INTEGER,
                desc        VARCHAR(2048)
            );`)
}

func rawExecDB(sqlStatement string) {
	db := openDB()
	defer db.Close()
	_, err := db.Exec(sqlStatement)
	if err != nil {
		log.Printf("SQL Error: %q: %s\n", err, sqlStatement)
	}
}

func rawQueryDB(sqlQuery string) (*sql.Rows, error) {
	db := openDB()
	defer db.Close()
	rows, err := db.Query(sqlQuery)
	return rows, err
}

/////////////////////////
// HELPER/META STRUCTS //
/////////////////////////

type vappData struct {
	Id     string
	Images []imageData
}

type imageData struct {
	Id   string
	Name string
	IP   string
}

////////////////////////
// EVENT DB FUNCTIONS //
////////////////////////

type Event struct {
	Id     int
	Type   int
	Kind   string
	Title  string
	Vapp   string
	Field1 string
	Field2 string
	Field3 string
	Switch bool
}

func newEvent() {

}

func getEvents(eventType int) []Event {

	// smells like sql injection
	events, err := rawQueryDB(fmt.Sprintf("SELECT * from `events` WHERE type='%d'", eventType))
	if err != nil {
		log.Fatal(err)
	}
	defer events.Close()

	eventSlice := []Event{}
	for events.Next() {
		ev := Event{}
		err = events.Scan(&ev.Id, &ev.Type, &ev.Kind, &ev.Title, &ev.Vapp, &ev.Field1, &ev.Field2, &ev.Field3, &ev.Switch)
		if err != nil {
			log.Fatal(err)
		}
		eventSlice = append(eventSlice, ev)
	}
	err = events.Err()
	if err != nil {
		log.Fatal(err)
	}
	return eventSlice
}

func getEvent(key string, value string) (Event, error) {

	ev := Event{}
	if !validateName(key) || !validateName(value) {
		return Event{}, errors.New("Invalid input")
	}

	// lmao... use a parameterized query at least
	event, err := rawQueryDB(fmt.Sprintf("SELECT * from `events` WHERE %s='%s'", key, value))
	if err != nil {
		log.Fatal(err)
	}
	defer event.Close()

	for event.Next() {
		err = event.Scan(&ev.Id, &ev.Type, &ev.Kind, &ev.Title, &ev.Vapp, &ev.Field1, &ev.Field2, &ev.Field3, &ev.Switch)
		if err != nil {
			log.Fatal(err)
		}
	}
	err = event.Err()
	return ev, err
}

func addEvent(ev Event) error {

	fmt.Println("titel  is", ev.Title)
	testEvent, err := getEvent("title", ev.Title)
	fmt.Println("testEvent is", testEvent)
	fmt.Println("err is", err)
	if err != nil {
		return err
	}
	if testEvent.Title != "" {
		return errors.New("Event already exists")
	}

	db := openDB()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT INTO `events` ('type', 'kind', 'title', 'vapp', 'field1', 'field2', 'field3', 'switch') VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(ev.Type, ev.Kind, ev.Title, ev.Vapp, ev.Field1, ev.Field2, ev.Field3, ev.Switch)
	if err != nil {
		return err
	}
	tx.Commit()
	return err

}

/////////////////
// LEADERBOARD //
/////////////////

type Player struct {
	Name  string
	Score int
}

func getLeaderboard() []Player {
	leaderboardList := []Player{}
	leaderboardList = append(leaderboardList, Player{"hi", 1})
	return leaderboardList
}
