const express = require('express');
const cors = require('cors');
const bodyParser = require('body-parser');
const db = require('./db');
const { nanoid } = require('nanoid');

const app = express();
const PORT = process.env.PORT || 3000;

app.use(cors());
app.use(bodyParser.json());

app.get('/events', (req, res) => {
  db.all('SELECT * FROM events', (err, rows) => {
    if (err) return res.status(500).json({ error: err.message });
    res.json(rows);
  });
});

app.get('/events/:id', (req, res) => {
  const id = req.params.id;
  db.get('SELECT * FROM events WHERE id = ?', [id], (err, row) => {
    if (err) return res.status(500).json({ error: err.message });
    if (!row) return res.status(404).json({ error: 'Event not found' });
    res.json(row);
  });
});

app.post('/purchase', (req, res) => {
  const { eventId, buyerName, qty } = req.body;
  if (!eventId || !buyerName || !qty || qty <= 0) {
    return res.status(400).json({ error: 'Invalid purchase data' });
  }

  db.get('SELECT * FROM events WHERE id = ?', [eventId], (err, event) => {
    if (err) return res.status(500).json({ error: err.message });
    if (!event) return res.status(404).json({ error: 'Event not found' });
    if (event.available_tickets < qty) {
      return res.status(400).json({ error: 'Not enough tickets available' });
    }

    const purchaseId = nanoid();
    const now = new Date().toISOString();

    const insert = db.prepare('INSERT INTO purchases (id,event_id,buyer_name,qty,created_at) VALUES (?,?,?,?,?)');
    insert.run(purchaseId, eventId, buyerName, qty, now, function (err) {
      if (err) return res.status(500).json({ error: err.message });

      db.run('UPDATE events SET available_tickets = available_tickets - ? WHERE id = ?', [qty, eventId], function (err2) {
        if (err2) return res.status(500).json({ error: err2.message });

        res.json({ id: purchaseId, event: event.name, buyerName, qty, created_at: now });
      });
    });
  });
});

app.get('/purchases', (req, res) => {
  db.all('SELECT p.id, p.event_id, e.name as event_name, p.buyer_name, p.qty, p.created_at FROM purchases p LEFT JOIN events e ON p.event_id = e.id ORDER BY p.created_at DESC', (err, rows) => {
    if (err) return res.status(500).json({ error: err.message });
    res.json(rows);
  });
});

app.listen(PORT, () => {
  console.log(`Server running on http://localhost:${PORT}`);
});
