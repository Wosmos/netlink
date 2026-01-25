# 🔧 NetLink Chat App - Technical Documentation

## 📁 Project Structure

```
netlink-chat/
├── frontend/                 # Next.js React Frontend
│   ├── src/
│   │   ├── app/             # App Router pages
│   │   │   ├── chat/        # Chat pages
│   │   │   ├── login/       # Auth pages
│   │   │   ├── register/
│   │   │   ├── forgot-password/
│   │   │   └── reset-password/
│   │   ├── components/      # Reusable components
│   │   ├── context/         # React contexts
│   │   ├── lib/            # Utilities & API client
│   │   └── styles/         # Global styles
│   ├── public/             # Static assets
│   └── package.json
│
├── go-to-do/               # Go Backend Server
│   ├── auth/               # Authentication logic
│   ├── handlers/           # HTTP handlers
│   ├── models/             # Data models
│   ├── repository/         # Database layer
│   ├── services/           # Business logic
│   ├── websocket/          # WebSocket handling
│   ├── config/             # Configuration
│   ├── main.go             # Server entry point
│   └── .env                # Environment variables
│
└── docs/                   # Documentation
    ├── FEATURE_ROADMAP.md
    ├── TECHNICAL_DOCS.md
    └── API_REFERENCE.md
```

---

## 🏗️ Architecture Overview

### Frontend Architecture (Next.js 15)
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   React Pages   │    │   API Client    │    │   WebSocket     │
│                 │    │                 │    │   Client        │
│ • Chat UI       │◄──►│ • HTTP Requests │◄──►│                 │
│ • Auth Forms    │    │ • Error Handling│    │ • Real-time     │
│ • User Profile  │    │ • Response Cache│    │   Messages      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Auth Context  │    │   Local Storage │    │   State Mgmt    │
│                 │    │                 │    │                 │
│ • User Session  │    │ • Preferences   │    │ • Message Cache │
│ • Login State   │    │ • Theme Settings│    │ • UI State      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Backend Architecture (Go)
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP Server   │    │   WebSocket     │    │   Database      │
│                 │    │   Hub           │    │                 │
│ • REST API      │◄──►│                 │◄──►│ • PostgreSQL    │
│ • CORS Handling │    │ • Real-time     │    │ • Connection    │
│ • Middleware    │    │   Broadcasting  │    │   Pool          │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Auth Service  │    │   Email Service │    │   Repository    │
│                 │    │                 │    │   Layer         │
│ • JWT/Sessions  │    │ • Resend API    │    │                 │
│ • Password Hash │    │ • Templates     │    │ • CRUD Ops      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

---

## 🗄️ Database Schema

### Current Tables

```sql
-- Users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    phone VARCHAR(20),
    avatar VARCHAR(255),
    verification_token VARCHAR(255),
    is_verified BOOLEAN DEFAULT FALSE,
    reset_token VARCHAR(255),
    reset_token_expires TIMESTAMP,
    last_seen_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Sessions table
CREATE TABLE sessions (
    id VARCHAR(255) PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Conversations table
CREATE TABLE conversations (
    id SERIAL PRIMARY KEY,
    type VARCHAR(20) NOT NULL CHECK (type IN ('direct', 'group')),
    name VARCHAR(255),
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Conversation members table
CREATE TABLE conversation_members (
    id SERIAL PRIMARY KEY,
    conversation_id INTEGER REFERENCES conversations(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) DEFAULT 'member',
    joined_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(conversation_id, user_id)
);

-- Messages table
CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    conversation_id INTEGER REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(20) DEFAULT 'text',
    content TEXT NOT NULL,
    reply_to_id INTEGER REFERENCES messages(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Message read status
CREATE TABLE message_reads (
    id SERIAL PRIMARY KEY,
    message_id INTEGER REFERENCES messages(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    read_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(message_id, user_id)
);

-- Tasks table (legacy - can be removed)
CREATE TABLE tasks (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    completed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Notes table (legacy - can be removed)
CREATE TABLE notes (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    is_pinned BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

### Indexes for Performance
```sql
-- User lookups
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_verification_token ON users(verification_token);
CREATE INDEX idx_users_reset_token ON users(reset_token);

-- Session lookups
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- Conversation queries
CREATE INDEX idx_conversation_members_user_id ON conversation_members(user_id);
CREATE INDEX idx_conversation_members_conversation_id ON conversation_members(conversation_id);

-- Message queries
CREATE INDEX idx_messages_conversation_id ON messages(conversation_id);
CREATE INDEX idx_messages_sender_id ON messages(sender_id);
CREATE INDEX idx_messages_created_at ON messages(created_at);

-- Read status
CREATE INDEX idx_message_reads_user_id ON message_reads(user_id);
CREATE INDEX idx_message_reads_message_id ON message_reads(message_id);
```

---

## 🔌 API Reference

### Authentication Endpoints

#### POST /api/auth/register
```json
{
  "email": "user@example.com",
  "password": "password123",
  "name": "John Doe"
}
```

#### POST /api/auth/login
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

#### POST /api/auth/forgot-password
```json
{
  "email": "user@example.com"
}
```

#### POST /api/auth/reset-password
```json
{
  "token": "reset_token_here",
  "password": "newpassword123"
}
```

### Chat Endpoints

#### GET /api/conversations
Returns list of user's conversations

#### POST /api/conversations/direct
```json
{
  "user_id": 123,
  "email": "friend@example.com"
}
```

#### GET /api/conversations/messages?id=123
Returns messages for conversation

#### POST /api/conversations/messages?id=123
```json
{
  "content": "Hello world!",
  "type": "text",
  "reply_to_id": 456
}
```

### WebSocket Events

#### Client → Server
```json
{
  "type": "typing",
  "conversation_id": 123
}
```

#### Server → Client
```json
{
  "type": "message",
  "conversation_id": 123,
  "payload": {
    "id": 789,
    "content": "Hello!",
    "sender_id": 456,
    "created_at": "2024-01-01T12:00:00Z"
  }
}
```

---

## 🔧 Development Setup

### Prerequisites
- Go 1.21+
- Node.js 18+
- PostgreSQL 14+
- Bun (package manager)

### Environment Variables

#### Backend (.env)
```env
# Database
DATABASE_URL=postgres://user:pass@localhost:5432/chatdb?sslmode=disable

# App
APP_URL=http://localhost:8080

# Email (Resend)
RESEND_API_KEY=your_resend_api_key
FROM_EMAIL=noreply@yourdomain.com

# Security
JWT_SECRET=your_jwt_secret_here
SESSION_SECRET=your_session_secret
```

#### Frontend (.env.local)
```env
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_WS_URL=ws://localhost:8080/ws
```

### Running the Application

#### Backend
```bash
cd go-to-do
go mod tidy
go build -o server.exe
./server.exe
```

#### Frontend
```bash
cd frontend
bun install
bun run dev
```

---

## 🚀 Deployment Guide

### Docker Setup

#### Backend Dockerfile
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
CMD ["./server"]
```

#### Frontend Dockerfile
```dockerfile
FROM node:18-alpine AS builder
WORKDIR /app
COPY package.json bun.lockb ./
RUN npm install -g bun
RUN bun install
COPY . .
RUN bun run build

FROM node:18-alpine
WORKDIR /app
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/package.json ./
RUN npm install -g bun
RUN bun install --production
CMD ["bun", "start"]
```

#### Docker Compose
```yaml
version: '3.8'
services:
  postgres:
    image: postgres:14
    environment:
      POSTGRES_DB: chatdb
      POSTGRES_USER: chatuser
      POSTGRES_PASSWORD: chatpass
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  backend:
    build: ./go-to-do
    environment:
      DATABASE_URL: postgres://chatuser:chatpass@postgres:5432/chatdb?sslmode=disable
      APP_URL: http://localhost:8080
    ports:
      - "8080:8080"
    depends_on:
      - postgres

  frontend:
    build: ./frontend
    environment:
      NEXT_PUBLIC_API_URL: http://localhost:8080
      NEXT_PUBLIC_WS_URL: ws://localhost:8080/ws
    ports:
      - "3000:3000"
    depends_on:
      - backend

volumes:
  postgres_data:
```

### Production Deployment

#### Using Railway/Render/Vercel
1. **Backend**: Deploy Go server to Railway/Render
2. **Frontend**: Deploy to Vercel
3. **Database**: Use managed PostgreSQL (Neon, Supabase)

#### Environment Setup
```bash
# Production environment variables
DATABASE_URL=postgresql://prod_user:pass@prod_host:5432/prod_db
APP_URL=https://your-api-domain.com
RESEND_API_KEY=your_production_key
```

---

## 🔍 Testing Strategy

### Unit Tests
```go
// Example test for auth service
func TestAuthService_Login(t *testing.T) {
    // Setup test database
    // Create test user
    // Test login functionality
    // Assert results
}
```

### Integration Tests
```javascript
// Example API test
describe('Chat API', () => {
  test('should send message', async () => {
    const response = await api.sendMessage(1, 'Hello');
    expect(response.success).toBe(true);
  });
});
```

### E2E Tests (Playwright)
```javascript
test('user can send message', async ({ page }) => {
  await page.goto('/login');
  await page.fill('[name="email"]', 'test@example.com');
  await page.fill('[name="password"]', 'password');
  await page.click('button[type="submit"]');
  
  await page.goto('/chat/1');
  await page.fill('[placeholder="TRANSMIT DATA..."]', 'Hello!');
  await page.click('button[type="submit"]');
  
  await expect(page.locator('text=Hello!')).toBeVisible();
});
```

---

## 📊 Performance Optimization

### Database Optimization
- Connection pooling (200 max connections)
- Query optimization with indexes
- Pagination for message history
- Caching frequently accessed data

### Frontend Optimization
- Code splitting with Next.js
- Image optimization
- Lazy loading for chat history
- WebSocket connection management

### Caching Strategy
```go
// Redis caching for online users
func (s *ChatService) GetOnlineUsers() ([]int, error) {
    // Check Redis cache first
    // Fallback to database
    // Update cache
}
```

---

## 🔒 Security Considerations

### Authentication
- HTTP-only cookies for sessions
- CSRF protection
- Rate limiting on auth endpoints
- Password hashing with bcrypt

### Data Protection
- Input validation and sanitization
- SQL injection prevention
- XSS protection
- CORS configuration

### WebSocket Security
- Origin validation
- Authentication required
- Rate limiting for messages
- Connection limits per user

---

## 📈 Monitoring & Logging

### Logging Strategy
```go
// Structured logging
log.WithFields(log.Fields{
    "user_id": userID,
    "action": "send_message",
    "conversation_id": convID,
}).Info("Message sent")
```

### Metrics to Track
- API response times
- WebSocket connection count
- Message delivery rates
- Error rates by endpoint
- User activity patterns

### Health Checks
```go
func healthCheck(w http.ResponseWriter, r *http.Request) {
    // Check database connection
    // Check Redis connection
    // Return status
}
```

---

*This technical documentation should be updated as the application evolves.*