Cara Menjalankan (Quick Start)
Anda tidak perlu menginstal Go atau compiler C di komputer Anda. Cukup clone repository ini dan jalankan menggunakan Docker.

1. Clone repository
Buka terminal Anda (PowerShell atau Command Prompt) dan jalankan perintah berikut:


git clone https://github.com/eldhien/Tugas_Besar_KomputasiAwan.git
cd <NAMA_FOLDER_PROJECT>
2. Jalankan dengan Docker Compose
Build image dan jalankan container:



docker compose up -d --build
Perintah ini akan membangun binary Go dan menyiapkan lingkungan (environment) secara otomatis.

Server akan berjalan dan mendengarkan (listen) pada port 3000.

3. Menghentikan aplikasi
Untuk menghentikan dan menghapus container:


docker compose down
Endpoint API
Base URL: http://localhost:3000

GET /events - menampilkan daftar acara

GET /events/:id - mengambil data satu acara

POST /purchase - membuat pembelian tiket dengan format JSON { eventId, buyerName, qty }

GET /purchases - menampilkan daftar pembelian

Catatan
Penyimpanan Data: Database SQLite disimpan secara permanen dalam volume Docker bernama tickets-data.

Pemecahan Masalah (Troubleshooting):

Jika Docker gagal mengunduh base image, pastikan koneksi internet Anda stabil.

Pastikan port 3000 diizinkan aksesnya oleh firewall komputer Anda.