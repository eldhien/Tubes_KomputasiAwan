const API = 'http://localhost:3000';

async function fetchEvents() {
  const res = await fetch(`${API}/events`);
  return res.json();
}

async function fetchPurchases() {
  const res = await fetch(`${API}/purchases`);
  return res.json();
}

function formatCurrency(v) {
  return new Intl.NumberFormat('id-ID').format(v);
}

function renderEvents(events) {
  const container = document.getElementById('events');
  container.innerHTML = '';
  events.forEach((e) => {
    const div = document.createElement('div');
    div.className = 'event';
    div.innerHTML = `
      <h3>${e.name}</h3>
      <p>Tanggal: ${e.date}</p>
      <p>Harga: Rp ${formatCurrency(e.price)}</p>
      <p>Tersisa: ${e.available_tickets}</p>
      <button data-id="${e.id}">Beli Tiket</button>
    `;
    container.appendChild(div);
  });

  container.querySelectorAll('button').forEach((btn) => {
    btn.addEventListener('click', async () => {
      const eventId = btn.dataset.id;
      const buyerName = prompt('Masukkan nama pembeli:');
      if (!buyerName) return alert('Nama wajib diisi');
      const qtyStr = prompt('Jumlah tiket yang dibeli:', '1');
      const qty = parseInt(qtyStr, 10);
      if (!qty || qty <= 0) return alert('Jumlah tidak valid');

      try {
        const res = await fetch(`${API}/purchase`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ eventId: Number(eventId), buyerName, qty }),
        });
        const data = await res.json();
        if (!res.ok) throw new Error(data.error || 'Gagal membeli');
        alert(`Berhasil! ID tiket: ${data.id}`);
        await reload();
      } catch (err) {
        alert(err.message);
      }
    });
  });
}

function renderPurchases(purchases) {
  const container = document.getElementById('purchases');
  container.innerHTML = '';
  if (!purchases.length) return container.innerHTML = '<p>Belum ada pembelian.</p>';
  purchases.forEach(p => {
    const el = document.createElement('div');
    el.className = 'purchase';
    el.textContent = `${p.created_at.split('T')[0]} — ${p.buyer_name} membeli ${p.qty} tiket untuk ${p.event_name} (ID: ${p.id})`;
    container.appendChild(el);
  });
}

async function reload() {
  const [events, purchases] = await Promise.all([fetchEvents(), fetchPurchases()]);
  renderEvents(events);
  renderPurchases(purchases);
}

window.addEventListener('load', () => {
  reload();
});
