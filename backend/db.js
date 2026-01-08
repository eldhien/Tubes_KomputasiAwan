const sqlite3 = require('sqlite3').verbose();
const path = require('path');

const dbFile = path.join(__dirname, 'tickets.db');
const db = new sqlite3.Database(dbFile);

db.serialize(() => {
  db.run(
    `CREATE TABLE IF NOT EXISTS events (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT NOT NULL,
      date TEXT NOT NULL,
      price INTEGER NOT NULL,
      available_tickets INTEGER NOT NULL
    )`
  );

  db.run(
    `CREATE TABLE IF NOT EXISTS purchases (
      id TEXT PRIMARY KEY,
      event_id INTEGER,
      buyer_name TEXT,
      qty INTEGER,
      created_at TEXT,
      FOREIGN KEY(event_id) REFERENCES events(id)
    )`
  );

  // Seed sample events if empty
  db.get('SELECT COUNT(*) AS cnt FROM events', (err, row) => {
    if (err) return console.error(err);
    if (row.cnt === 0) {
      const stmt = db.prepare('INSERT INTO events (name,date,price,available_tickets) VALUES (?,?,?,?)');
      stmt.run('Konser A - Pop Night', '2026-03-20', 150000, 100);
      stmt.run('Konser B - Rock Live', '2026-04-10', 200000, 80);
      stmt.run('Konser C - Jazz Evening', '2026-05-05', 120000, 50);
      stmt.finalize();
    }
  });
});

module.exports = db;
