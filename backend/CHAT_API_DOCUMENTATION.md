# Net Link API - Private Integration Guide

A real-time chat API built with Go, PostgreSQL, and WebSocket support. This guide covers how to integrate this API as a private chat service in your projects.

## Quick Start

### Prerequisites
- Go 1.21+
- PostgreSQL 14+

### Environment Setup

Create a `.env` file:
```env
DATABASE_URL=postgres://username:password@localhost:5432/your_db?sslmode=disable
APP_URL=http://localhost:8080
```

### Run the Server
```bash
cd go-to-do
go run main.go
```

Server starts at `http://localhost:8080` with WebSocket at `ws://localhost:8080/ws`

---

## Database Schema

The API auto-creates these tables on startup:

```sql
-- Users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20),
    name VARCHAR(255),
    avatar VARCHAR(500),
    password_hash VARCHAR(255) NOT NULL,
    is_verified BOOLEAN DEFAULT FALSE,
    last_seen_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Conversations (direct or group)
CREATE TABLE conversations (
    id SERIAL PRIMARY KEY,
    type VARCHAR(20) NOT NULL DEFAULT 'direct',  -- 'direct' or 'group'
    name VARCHAR(255),                            -- For groups
    avatar TEXT,                                  -- For groups
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Conversation members
CREATE TABLE conversation_members (
    conversation_id INTEGER REFERENCES conversations(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) DEFAULT 'member',  -- 'admin' or 'member'
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    last_read_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (conversation_id, user_id)
);

-- Messages
CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    conversation_id INTEGER REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id INTEGER REFERENCES users(id),
    type VARCHAR(20) DEFAULT 'text',  -- 'text', 'image', 'file', 'system'
    content TEXT NOT NULL,
    reply_to_id INTEGER REFERENCES messages(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
```

---

## Authentication

All API endpoints (except login/register) require authentication via session cookie.

### Register User
```http
POST /api/auth/register
Content-Type: application/json

{
    "email": "user@example.com",
    "password": "securepassword123",
    "name": "John Doe"
}
```

**Response:**
```json
{
    "success": true,
    "message": "Verification email sent"
}
```

### Login
```http
POST /api/auth/login
Content-Type: application/json

{
    "email": "user@example.com",
    "password": "securepassword123"
}
```

**Response:**
```json
{
    "success": true,
    "data": {
        "id": 1,
        "email": "user@example.com",
        "name": "John Doe",
        "is_verified": true
    }
}
```

> **Note:** The response sets a `session_id` cookie (HttpOnly, 7-day expiry). Include this cookie in all subsequent requests.

### Get Current User
```http
GET /api/auth/me
Cookie: session_id=<session_token>
```

**Response:**
```json
{
    "success": true,
    "data": {
        "id": 1,
        "email": "user@example.com",
        "name": "John Doe",
        "is_verified": true,
        "last_seen_at": "2024-01-15T10:30:00Z"
    }
}
```

### Logout
```http
POST /api/auth/logout
Cookie: session_id=<session_token>
```

---

## Chat API Endpoints

### List Conversations
```http
GET /api/conversations
Cookie: session_id=<session_token></session_token>*Response:**
```json
{
    "success": true,
    "data": [
        {
            "id": 1,
            "type": "direct",
            "name": "",
            "created_at": "2024-01-15T10:00:00Z",
            "updated_at": "2024-01-15T12:30:00Z",
            "members": [
                {
                    "user_id": 1,
                    "role": "member",
                    "user": {
                        "id": 1,
                        "email": "user1@example.com",
                        "name": "User One"
                    }
                },
                {
                    "user_id": 2,
                    "role": "member",
                    "user": {
                        "id": 2,
                        "email": "user2@example.com",
                        "name": "User Two"
                    }
                }
            ],
            "last_message": {
                "id": 42,
                "content": "Hello!",
                "type": "text",
                "created_at": "2024-01-15T12:30:00Z"
            },
            "unread_count": 3
        }
    ]
}
```

### Create Direct Chat
```http
POST /api/conversations/direct
Cookie: session_id=<session_token>
Content-Type: application/json

{
    "user_id": 2
}
```

Or find user by email/phone:
```json
{
    "email": "user2@example.com"
}
```
```json
{
    "phone": "+1234567890"
}
```

**Response:**
```json
{
    "success": true,
    "data": {
        "id": 1,
        "type": "direct",
        "members": [...],
        "created_at": "2024-01-15T10:00:00Z"
    }
}
```

### Create Group Chat
```http
POST /api/conversations/group
Cookie: session_id=<session_token>
Content-Type: application/json

{
    "name": "Project Team",
    "member_ids": [2, 3, 4]
}
```

**Response:**
```json
{
    "id": 2,
    "type": "group",
    "name": "Project Team",
    "created_by": 1,
    "members": [...]
}
```

### Get Messages
```http
GET /api/conversations/messages?id=1&limit=50&offset=0
Cookie: session_id=<session_token>
```

**Parameters:**
- `id` (required): Conversation ID
- `limit` (optional): Number of messages (default: 50)
- `offset` (optional): Pagination offset (default: 0)

**Response:**
```json
{
    "success": true,
    "data": [
        {
            "id": 1,
            "conversation_id": 1,
            "sender_id": 2,
            "type": "text",
            "content": "Hey there!",
            "created_at": "2024-01-15T10:00:00Z",
            "updated_at": "2024-01-15T10:00:00Z",
            "sender": {
                "id": 2,
                "email": "user2@example.com",
                "name": "User Two"
            }
        }
    ]
}
```

### Send Message
```http
POST /api/conversations/messages?id=1
Cookie: session_id=<session_token>
Content-Type: application/json

{
    "content": "Hello, world!",
    "type": "text",
    "reply_to_id": null
}
```

**Message Types:** `text`, `image`, `file`, `system`

**Response:**
```json
{
    "success": true,
    "data": {
        "id": 43,
        "conversation_id": 1,
        "sender_id": 1,
        "type": "text",
        "content": "Hello, world!",
        "created_at": "2024-01-15T12:35:00Z",
        "sender": {
            "id": 1,
            "name": "User One"
        }
    }
}
```

### Edit Message
```http
PUT /api/conversations/messages/edit?id=1&msg_id=43
Cookie: session_id=<session_token>
Content-Type: application/json

{
    "content": "Hello, world! (edited)"
}
```

### Mark Conversation as Read
```http
POST /api/conversations/read?id=1
Cookie: session_id=<session_token>
```

### Send Typing Indicator
```http
POST /api/conversations/typing?id=1
Cookie: session_id=<session_token>
```

### Delete Conversation
```http
DELETE /api/conversations/delete?id=1
Cookie: session_id=<session_token>
```

> For groups: removes you from the group. For direct chats: deletes the conversation.

---

## User Endpoints

### Search Users
```http
GET /api/users/search?q=john
Cookie: session_id=<session_token>
```

**Response:**
```json
{
    "success": true,
    "data": [
        {
            "id": 2,
            "email": "john@example.com",
            "name": "John Doe",
            "avatar": ""
        }
    ]
}
```

### Get Online Users
```http
GET /api/users/online
Cookie: session_id=<session_token>
```

**Response:**
```json
{
    "online_users": [1, 3, 5]
}
```

---

## WebSocket Real-Time Events

### Connect
```javascript
const ws = new WebSocket('ws://localhost:8080/ws');
// Include session cookie for authentication
```

### Event Types

| Event | Description |
|-------|-------------|
| `message` | New message received |
| `typing` | User is typing |
| `stop_typing` | User stopped typing |
| `online` | User came online |
| `offline` | User went offline |
| `read` | Messages marked as read |
| `conversation` | New conversation created |

### Event Format
```json
{
    "type": "message",
    "conversation_id": 1,
    "user_id": 2,
    "payload": { /* message object */ },
    "timestamp": "2024-01-15T12:35:00Z"
}
```

### JavaScript Integration Example
```javascript
class ChatClient {
    constructor(baseUrl) {
        this.baseUrl = baseUrl;
        this.ws = null;
        this.handlers = {};
    }

    connect() {
        this.ws = new WebSocket(`${this.baseUrl.replace('http', 'ws')}/ws`);
        
        this.ws.onmessage = (event) => {
            const data = JSON.parse(event.data);
            if (this.handlers[data.type]) {
                this.handlers[data.type](data);
            }
        };

        this.ws.onopen = () => console.log('Connected');
        this.ws.onclose = () => setTimeout(() => this.connect(), 3000);
    }

    on(eventType, handler) {
        this.handlers[eventType] = handler;
    }

    sendTyping(conversationId) {
        this.ws.send(JSON.stringify({
            type: 'typing',
            conversation_id: conversationId
        }));
    }
}

// Usage
const chat = new ChatClient('http://localhost:8080');
chat.connect();

chat.on('message', (event) => {
    console.log('New message:', event.payload);
});

chat.on('typing', (event) => {
    console.log(`User ${event.user_id} is typing...`);
});

chat.on('online', (event) => {
    console.log(`User ${event.user_id} is online`);
});
```

---

## Integration with Your Project

### 1. As a Microservice

Run the chat API as a separate service and connect from your main app:

```javascript
// Your frontend/backend
const CHAT_API = 'http://localhost:8080';

async function sendMessage(conversationId, content) {
    const response = await fetch(`${CHAT_API}/api/conversations/messages?id=${conversationId}`, {
        method: 'POST',
        credentials: 'include',  // Important for cookies
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ content, type: 'text' })
    });
    return response.json();
}
```

### 2. Using a Separate Database

Update your `.env` to point to a different database:
```env
DATABASE_URL=postgres://user:pass@your-host:5432/chat_db?sslmode=require
```

The schema auto-initializes on first run.

### 3. CORS Configuration

The API allows CORS from `localhost:3000` by default. To customize, modify `main.go`:

```go
corsMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Add your allowed origins
        allowedOrigins := []string{
            "http://localhost:3000",
            "https://your-app.com",
        }
        origin := r.Header.Get("Origin")
        for _, allowed := range allowedOrigins {
            if origin == allowed {
                w.Header().Set("Access-Control-Allow-Origin", origin)
                break
            }
        }
        // ... rest of middleware
    }
}
```

### 4. Custom Port

Change the server port in `main.go`:
```go
server := &http.Server{
    Addr: ":3001",  // Change from 8080
    // ...
}
```

---

## Data Models

### User
```go
type User struct {
    ID         int       `json:"id"`
    Email      string    `json:"email"`
    Phone      string    `json:"phone,omitempty"`
    Name       string    `json:"name"`
    Avatar     string    `json:"avatar,omitempty"`
    IsVerified bool      `json:"is_verified"`
    LastSeenAt time.Time `json:"last_seen_at,omitempty"`
    CreatedAt  time.Time `json:"created_at"`
}
```

### Conversation
```go
type Conversation struct {
    ID          int                  `json:"id"`
    Type        string               `json:"type"`      // "direct" or "group"
    Name        string               `json:"name"`      // For groups
    Avatar      string               `json:"avatar"`    // For groups
    CreatedBy   int                  `json:"created_by"`
    Members     []ConversationMember `json:"members,omitempty"`
    LastMessage *Message             `json:"last_message,omitempty"`
    UnreadCount int                  `json:"unread_count,omitempty"`
    CreatedAt   time.Time            `json:"created_at"`
    UpdatedAt   time.Time            `json:"updated_at"`
}
```

### Message
```go
type Message struct {
    ID             int       `json:"id"`
    ConversationID int       `json:"conversation_id"`
    SenderID       int       `json:"sender_id"`
    Type           string    `json:"type"`     // "text", "image", "file", "system"
    Content        string    `json:"content"`
    ReplyToID      *int      `json:"reply_to_id,omitempty"`
    Sender         *User     `json:"sender,omitempty"`
    ReplyTo        *Message  `json:"reply_to,omitempty"`
    CreatedAt      time.Time `json:"created_at"`
    UpdatedAt      time.Time `json:"updated_at"`
}
```

---

## Error Responses

All errors follow this format:
```json
{
    "success": false,
    "error": "Error message here"
}
```

Common HTTP status codes:
- `400` - Bad Request (invalid input)
- `401` - Unauthorized (not logged in)
- `404` - Not Found (conversation/user not found)
- `405` - Method Not Allowed
- `500` - Internal Server Error

---

## Performance Configuration

The server is optimized for high concurrency:

```go
// Connection pool settings (main.go)
poolConfig.MaxConns = 200
poolConfig.MinConns = 20
poolConfig.MaxConnLifetime = 30 * time.Minute
poolConfig.MaxConnIdleTime = 5 * time.Minute

// Server timeouts
server := &http.Server{
    ReadTimeout:    30 * time.Second,
    WriteTimeout:   30 * time.Second,
    IdleTimeout:    120 * time.Second,
    MaxHeaderBytes: 1 << 20,  // 1MB
}
```

---

## Security Notes

1. **Session-based auth**: Uses HttpOnly cookies (7-day expiry)
2. **Password hashing**: bcrypt with cost factor 8
3. **Email verification**: Required before login
4. **CORS**: Configured for specific origins
5. **WebSocket auth**: Validates session cookie on connection

For production:
- Enable HTTPS
- Set `Secure: true` on cookies
- Use environment variables for secrets
- Configure proper CORS origins
