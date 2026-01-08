package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	_ "github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type Event struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	Date             string `json:"date"`
	Price            int    `json:"price"`
	AvailableTickets int    `json:"available_tickets"`
}

type Purchase struct {
	ID        string `json:"id"`
	EventID   int    `json:"event_id"`
	BuyerName string `json:"buyer_name"`
	Qty       int    `json:"qty"`
	CreatedAt string `json:"created_at"`
}

func main() {
	dbPath := filepath.Join(".", "tickets.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	if err := initDB(db); err != nil {
		log.Fatalf("failed to init db: %v", err)
	}

	r := mux.NewRouter()
	r.Use(corsMiddleware)

	r.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		events, err := listEvents(db)
		if err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}
		writeJSON(w, events)
	}).Methods("GET", "OPTIONS")

	r.HandleFunc("/events/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		idStr := vars["id"]
		id, _ := strconv.Atoi(idStr)
		e, err := getEvent(db, id)
		if err == sql.ErrNoRows {
			httpError(w, fmt.Errorf("event not found"), http.StatusNotFound)
			return
		}
		if err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}
		writeJSON(w, e)
	}).Methods("GET", "OPTIONS")

	r.HandleFunc("/purchase", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			EventId   int    `json:"eventId"`
			BuyerName string `json:"buyerName"`
			Qty       int    `json:"qty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpError(w, err, http.StatusBadRequest)
			return
		}
		if req.EventId == 0 || req.BuyerName == "" || req.Qty <= 0 {
			httpError(w, fmt.Errorf("invalid purchase data"), http.StatusBadRequest)
			return
		}

		purchase, err := createPurchase(db, req.EventId, req.BuyerName, req.Qty)
		if err != nil {
			httpError(w, err, http.StatusBadRequest)
			return
		}
		writeJSON(w, purchase)
	}).Methods("POST", "OPTIONS")

	r.HandleFunc("/purchases", func(w http.ResponseWriter, r *http.Request) {
		rows, err := listPurchases(db)
		if err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}
		writeJSON(w, rows)
	}).Methods("GET", "OPTIONS")

	// Serve frontend static files from ./frontend (copied into image)
	frontendDir := filepath.Join(".", "frontend")
	fs := http.FileServer(http.Dir(frontendDir))
	r.PathPrefix("/").Handler(fs)

	addr := ":3000"
	log.Printf("Go server running on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}

func initDB(db *sql.DB) error {
	createEvents := `CREATE TABLE IF NOT EXISTS events (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT NOT NULL,
      date TEXT NOT NULL,
      price INTEGER NOT NULL,
      available_tickets INTEGER NOT NULL
    )`

	createPurchases := `CREATE TABLE IF NOT EXISTS purchases (
      id TEXT PRIMARY KEY,
      event_id INTEGER,
      buyer_name TEXT,
      qty INTEGER,
      created_at TEXT,
      FOREIGN KEY(event_id) REFERENCES events(id)
    )`

	if _, err := db.Exec(createEvents); err != nil {
		return err
	}
	if _, err := db.Exec(createPurchases); err != nil {
		return err
	}

	// seed if empty
	var cnt int
	err := db.QueryRow("SELECT COUNT(*) FROM events").Scan(&cnt)
	if err != nil {
		return err
	}
	if cnt == 0 {
		tx, _ := db.Begin()
		stmt, _ := tx.Prepare("INSERT INTO events (name,date,price,available_tickets) VALUES (?,?,?,?)")
		defer stmt.Close()
		stmt.Exec("Konser A - Pop Night", "2026-03-20", 150000, 100)
		stmt.Exec("Konser B - Rock Live", "2026-04-10", 200000, 80)
		stmt.Exec("Konser C - Jazz Evening", "2026-05-05", 120000, 50)
		tx.Commit()
	}
	return nil
}

func listEvents(db *sql.DB) ([]Event, error) {
	rows, err := db.Query("SELECT id,name,date,price,available_tickets FROM events")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.ID, &e.Name, &e.Date, &e.Price, &e.AvailableTickets); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, nil
}

func getEvent(db *sql.DB, id int) (Event, error) {
	var e Event
	err := db.QueryRow("SELECT id,name,date,price,available_tickets FROM events WHERE id = ?", id).Scan(&e.ID, &e.Name, &e.Date, &e.Price, &e.AvailableTickets)
	return e, err
}

func createPurchase(db *sql.DB, eventId int, buyer string, qty int) (Purchase, error) {
	tx, err := db.Begin()
	if err != nil {
		return Purchase{}, err
	}
	defer tx.Rollback()

	var available int
	err = tx.QueryRow("SELECT available_tickets FROM events WHERE id = ?", eventId).Scan(&available)
	if err == sql.ErrNoRows {
		return Purchase{}, fmt.Errorf("event not found")
	}
	if err != nil {
		return Purchase{}, err
	}
	if available < qty {
		return Purchase{}, fmt.Errorf("not enough tickets available")
	}

	id := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)

	_, err = tx.Exec("INSERT INTO purchases (id,event_id,buyer_name,qty,created_at) VALUES (?,?,?,?,?)", id, eventId, buyer, qty, now)
	if err != nil {
		return Purchase{}, err
	}
	_, err = tx.Exec("UPDATE events SET available_tickets = available_tickets - ? WHERE id = ?", qty, eventId)
	if err != nil {
		return Purchase{}, err
	}

	if err := tx.Commit(); err != nil {
		return Purchase{}, err
	}

	return Purchase{ID: id, EventID: eventId, BuyerName: buyer, Qty: qty, CreatedAt: now}, nil
}

func listPurchases(db *sql.DB) ([]map[string]interface{}, error) {
	rows, err := db.Query(`SELECT p.id, p.event_id, e.name AS event_name, p.buyer_name, p.qty, p.created_at FROM purchases p LEFT JOIN events e ON p.event_id = e.id ORDER BY p.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]interface{}
	for rows.Next() {
		var id, eventName, buyer string
		var eventId, qty int
		var createdAt string
		if err := rows.Scan(&id, &eventId, &eventName, &buyer, &qty, &createdAt); err != nil {
			return nil, err
		}
		m := map[string]interface{}{"id": id, "event_id": eventId, "event_name": eventName, "buyer_name": buyer, "qty": qty, "created_at": createdAt}
		out = append(out, m)
	}
	return out, nil
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func httpError(w http.ResponseWriter, err error, code int) {
	w.WriteHeader(code)
	writeJSON(w, map[string]string{"error": err.Error()})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
