## Workmate Backend

### Menjalankan dengan Docker

1. Build dan jalankan layanan:
   ```bash
   docker compose up -d --build
   ```

2. API akan tersedia di `http://localhost:8080`.

3. Endpoint:
   - GET `/healthz`
   - POST `/api/register`
   - POST `/api/login`

### Environment

- `PORT` default 8080
- `JWT_SECRET` default `dev-secret-change-me`
- `DATABASE_URL` untuk container sudah diset ke `postgres://postgres:postgres@db:5432/postgres?sslmode=disable`


