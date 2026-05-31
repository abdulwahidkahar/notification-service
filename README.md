# notification-service

Eksperimen notification service berbasis Go dan Redis Streams untuk mempelajari pola message queue, consumer group, worker paralel, retry, dan dead letter stream.

Project ini dibuat sebagai contoh sederhana bahwa menambah worker tidak selalu otomatis membuat proses lebih cepat. Penjelasan lengkapnya ada di blog: [menambah worker tidak selalu berarti lebih cepat: eksperimen message queue dengan redis streams dan go](https://www.abdulwahidkahar.com/blog/menambah-worker-tidak-selalu-berarti-lebih-cepat-eksperimen-message-queue-dengan-redis-streams-dan-go).

## fitur

- publish event notifikasi dummy ke Redis Stream
- consume message menggunakan Redis consumer group
- worker name otomatis berdasarkan process id
- acknowledge message setelah diproses
- retry message ketika proses gagal
- dead letter stream setelah retry mencapai batas maksimum
- log statistik throughput setiap 100 message

## kebutuhan

- Go sesuai versi di `go.mod`
- Docker dan Docker Compose
- Redis, bisa dijalankan lewat `docker-compose.yaml`

## setup

Jalankan dari root project:

```bash
go mod download
docker compose up -d
```

Cek Redis sudah hidup:

```bash
docker compose ps
```

## menjalankan producer

Producer akan mengirim 1000 message dummy ke stream `notification:events`.

```bash
go run ./cmd/producer
```

## menjalankan worker

Worker akan membaca message dari stream menggunakan consumer group `notification-workers`.

```bash
go run ./cmd/worker
```

Untuk eksperimen beberapa worker, buka beberapa terminal dan jalankan command yang sama:

```bash
go run ./cmd/worker
```

Setelah worker aktif, jalankan producer lagi:

```bash
go run ./cmd/producer
```

## stream yang digunakan

- `notification:events`: stream utama untuk event notifikasi
- `notification:dead-letter`: stream untuk message yang gagal setelah retry maksimum
- `notification-workers`: consumer group untuk worker

## struktur folder

```text
.
├── cmd
│   ├── producer
│   │   └── main.go
│   └── worker
│       └── main.go
├── internal
│   ├── model
│   │   └── event.go
│   └── queue
│       ├── consumer.go
│       ├── producer.go
│       └── redis.go
├── docker-compose.yaml
├── go.mod
└── go.sum
```

## perintah berguna

Menjalankan test:

```bash
go test ./...
```

Melihat isi stream utama:

```bash
docker compose exec redis redis-cli xinfo stream notification:events
```

Melihat pending message di consumer group:

```bash
docker compose exec redis redis-cli xpending notification:events notification-workers
```

Melihat dead letter stream:

```bash
docker compose exec redis redis-cli xrange notification:dead-letter - +
```

Menghapus data Redis lokal:

```bash
docker compose down
```

## catatan

Redis client saat ini menggunakan alamat default `localhost:6379` di `internal/queue/redis.go`. Jika Redis berjalan di host atau port berbeda, ubah konfigurasi client di file tersebut.
