# NetLink

A real-time chat application with web and mobile clients.

## Tech Stack

- **Backend:** Go (stdlib `net/http`), PostgreSQL (`pgx`), Gorilla WebSocket
- **Frontend:** Next.js 16, React 19, Tailwind CSS v4, TypeScript
- **Mobile:** Expo SDK 54, React Native, Zustand, TypeScript
- **Deployment:** Docker (Railway), Vercel (frontend), EAS Build (mobile)

## Features

- Real-time messaging via WebSocket (multi-device support)
- Direct and group conversations
- Message editing, deletion (soft delete), forwarding
- Emoji reactions (system + custom)
- Voice messages with waveform visualization
- Email verification and password reset (Gmail SMTP / Resend API)
- Session-based authentication (cookie + bearer token)
- Optimistic UI updates on web and mobile
- Rate limiting on auth endpoints

## Project Structure

```
backend/          Go HTTP + WebSocket server
  auth/           Authentication service
  handlers/       HTTP request handlers
  middleware/     Rate limiter, JSON response helpers
  models/         Data models
  repository/     PostgreSQL queries (raw SQL, no ORM)
  services/       Email service (Gmail SMTP / Resend)
  websocket/      Hub + client (real-time event broadcasting)
frontend/         Next.js web app
mobile/           Expo React Native app
```

## Getting Started

### Prerequisites

- Go 1.25+
- PostgreSQL
- Node.js / Bun (for frontend)
- Expo CLI (for mobile)

### Backend

```bash
cd backend
cp .env.example .env  # fill in DATABASE_URL, email config
go run main.go
```

### Frontend

```bash
cd frontend
bun install
bun dev
```

### Mobile

```bash
cd mobile
npm install
npx expo start
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `DATABASE_URL` | PostgreSQL connection string |
| `PORT` | Server port (default: 8080) |
| `ALLOWED_ORIGINS` | Comma-separated CORS origins |
| `ENV` | Set to `production` for secure cookies |
| `APP_URL` | Base URL for email links |
| `GMAIL_USERNAME` | Gmail address for SMTP |
| `GMAIL_APP_PASSWORD` | Gmail app password |
| `RESEND_API_KEY` | Resend API key (alternative to Gmail) |
| `FROM_EMAIL` | Sender email address |

## Running Tests

```bash
cd backend
go test ./... -v
```

## Known Limitations

- Voice files stored on local filesystem (should use S3/MinIO for production)
- No end-to-end or integration tests yet (unit tests only)
- HTML template routes (`/login`, `/register`) are legacy from early development
