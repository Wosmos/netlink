# 📡 NetLink Chat App - API Reference

## 🔗 Base URL
```
Development: http://localhost:8080
Production: https://your-api-domain.com
```

## 🔐 Authentication
All API endpoints (except auth endpoints) require authentication via HTTP-only cookies set during login.

---

## 🔑 Authentication Endpoints

### Register User
**POST** `/api/auth/register`

Creates a new user account and sends verification email.

#### Request Body
```json
{
  "email": "user@example.com",
  "password": "password123",
  "name": "John Doe"
}
```

#### Response
```json
{
  "success": true,
  "message": "Verification email sent"
}
```

#### Error Responses
```json
{
  "success": false,
  "error": "Email already registered"
}
```

---

### Login User
**POST** `/api/auth/login`

Authenticates user and creates session.

#### Request Body
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

#### Response
```json
{
  "success": true,
  "data": {
    "id": 1,
    "email": "user@example.com",
    "name": "John Doe",
    "is_verified": true,
    "created_at": "2024-01-01T12:00:00Z"
  }
}
```

---

### Get Current User
**GET** `/api/auth/me`

Returns current authenticated user information.

#### Response
```json
{
  "success": true,
  "data": {
    "id": 1,
    "email": "user@example.com",
    "name": "John Doe",
    "avatar": "https://example.com/avatar.jpg",
    "is_verified": true,
    "last_seen_at": "2024-01-01T12:00:00Z",
    "created_at": "2024-01-01T12:00:00Z"
  }
}
```

---

### Logout User
**POST** `/api/auth/logout`

Destroys user session.

#### Response
```json
{
  "success": true,
  "message": "Logged out"
}
```

---

### Forgot Password
**POST** `/api/auth/forgot-password`

Sends password reset email to user.

#### Request Body
```json
{
  "email": "user@example.com"
}
```

#### Response
```json
{
  "success": true,
  "message": "If that email is registered, you will receive a reset link"
}
```

---

### Reset Password
**POST** `/api/auth/reset-password`

Resets user password using token from email.

#### Request Body
```json
{
  "token": "reset_token_from_email",
  "password": "newpassword123"
}
```

#### Response
```json
{
  "success": true,
  "message": "Password reset successful"
}
```

---

### Verify Email
**GET** `/api/auth/verify?token=verification_token`

Verifies user email address. Redirects to frontend login page.

---

## 💬 Chat Endpoints

### Get Conversations
**GET** `/api/conversations`

Returns list of user's conversations.

#### Response
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "type": "direct",
      "name": null,
      "members": [
        {
          "user_id": 1,
          "user": {
            "id": 1,
            "name": "John Doe",
            "email": "john@example.com",
            "avatar": "https://example.com/avatar1.jpg"
          },
          "role": "member",
          "joined_at": "2024-01-01T12:00:00Z"
        },
        {
          "user_id": 2,
          "user": {
            "id": 2,
            "name": "Jane Smith",
            "email": "jane@example.com",
            "avatar": "https://example.com/avatar2.jpg"
          },
          "role": "member",
          "joined_at": "2024-01-01T12:00:00Z"
        }
      ],
      "last_message": {
        "id": 123,
        "content": "Hello there!",
        "sender_id": 2,
        "created_at": "2024-01-01T12:30:00Z"
      },
      "unread_count": 2,
      "created_at": "2024-01-01T12:00:00Z"
    }
  ]
}
```

---

### Create Direct Chat
**POST** `/api/conversations/direct`

Creates a direct conversation with another user.

#### Request Body
```json
{
  "user_id": 2
}
```

#### Alternative (find by email)
```json
{
  "email": "friend@example.com"
}
```

#### Response
```json
{
  "success": true,
  "data": {
    "id": 1,
    "type": "direct",
    "members": [...],
    "created_at": "2024-01-01T12:00:00Z"
  }
}
```

---

### Create Group Chat
**POST** `/api/conversations/group`

Creates a group conversation.

#### Request Body
```json
{
  "name": "Project Team",
  "member_ids": [2, 3, 4]
}
```

#### Response
```json
{
  "success": true,
  "data": {
    "id": 2,
    "type": "group",
    "name": "Project Team",
    "members": [...],
    "created_at": "2024-01-01T12:00:00Z"
  }
}
```

---

### Get Messages
**GET** `/api/conversations/messages?id=1&limit=50&offset=0`

Returns messages for a conversation.

#### Query Parameters
- `id` (required): Conversation ID
- `limit` (optional): Number of messages to return (default: 50)
- `offset` (optional): Number of messages to skip (default: 0)

#### Response
```json
{
  "success": true,
  "data": [
    {
      "id": 123,
      "conversation_id": 1,
      "sender_id": 2,
      "sender": {
        "id": 2,
        "name": "Jane Smith",
        "email": "jane@example.com",
        "avatar": "https://example.com/avatar2.jpg"
      },
      "type": "text",
      "content": "Hello there!",
      "reply_to_id": null,
      "reply_to": null,
      "read_by": [1, 2],
      "created_at": "2024-01-01T12:30:00Z",
      "updated_at": "2024-01-01T12:30:00Z"
    }
  ]
}
```

---

### Send Message
**POST** `/api/conversations/messages?id=1`

Sends a message to a conversation.

#### Request Body
```json
{
  "content": "Hello everyone!",
  "type": "text",
  "reply_to_id": 122
}
```

#### Response
```json
{
  "success": true,
  "data": {
    "id": 124,
    "conversation_id": 1,
    "sender_id": 1,
    "content": "Hello everyone!",
    "type": "text",
    "reply_to_id": 122,
    "created_at": "2024-01-01T12:35:00Z"
  }
}
```

---

### Edit Message
**PUT** `/api/conversations/messages/edit?id=1&msg_id=124`

Edits a message (only sender can edit).

#### Request Body
```json
{
  "content": "Hello everyone! (edited)"
}
```

#### Response
```json
{
  "success": true,
  "data": {
    "id": 124,
    "content": "Hello everyone! (edited)",
    "updated_at": "2024-01-01T12:36:00Z"
  }
}
```

---

### Mark as Read
**POST** `/api/conversations/read?id=1`

Marks all messages in conversation as read.

#### Response
```json
{
  "success": true,
  "message": "Messages marked as read"
}
```

---

### Send Typing Indicator
**POST** `/api/conversations/typing?id=1`

Sends typing indicator to other users in conversation.

#### Response
```json
{
  "success": true,
  "message": "Typing indicator sent"
}
```

---

### Delete Conversation
**DELETE** `/api/conversations/delete?id=1`

Deletes a conversation (removes user from conversation).

#### Response
```json
{
  "success": true,
  "message": "Conversation deleted"
}
```

---

## 👥 User Endpoints

### Get Online Users
**GET** `/api/users/online`

Returns list of currently online user IDs.

#### Response
```json
{
  "success": true,
  "data": {
    "online_users": [1, 2, 5, 8]
  }
}
```

---

### Search Users
**GET** `/api/users/search?q=john`

Searches for users by name or email.

#### Query Parameters
- `q` (required): Search query

#### Response
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com",
      "avatar": "https://example.com/avatar1.jpg"
    },
    {
      "id": 3,
      "name": "Johnny Smith",
      "email": "johnny@example.com",
      "avatar": "https://example.com/avatar3.jpg"
    }
  ]
}
```

---

## 🔌 WebSocket Events

### Connection
Connect to WebSocket at: `ws://localhost:8080/ws`

Authentication is handled automatically via cookies.

### Client → Server Events

#### Typing Indicator
```json
{
  "type": "typing",
  "conversation_id": 1
}
```

### Server → Client Events

#### New Message
```json
{
  "type": "message",
  "conversation_id": 1,
  "user_id": 2,
  "payload": {
    "id": 125,
    "conversation_id": 1,
    "sender_id": 2,
    "content": "New message!",
    "type": "text",
    "created_at": "2024-01-01T12:40:00Z"
  },
  "timestamp": "2024-01-01T12:40:00Z"
}
```

#### Typing Indicator
```json
{
  "type": "typing",
  "conversation_id": 1,
  "user_id": 2,
  "timestamp": "2024-01-01T12:40:00Z"
}
```

#### User Online
```json
{
  "type": "online",
  "user_id": 3,
  "timestamp": "2024-01-01T12:40:00Z"
}
```

#### User Offline
```json
{
  "type": "offline",
  "user_id": 3,
  "timestamp": "2024-01-01T12:40:00Z"
}
```

#### Message Read
```json
{
  "type": "read",
  "conversation_id": 1,
  "user_id": 2,
  "message_id": 125,
  "timestamp": "2024-01-01T12:40:00Z"
}
```

#### New Conversation
```json
{
  "type": "conversation",
  "user_id": 1,
  "payload": {
    "id": 3,
    "type": "direct",
    "members": [...],
    "created_at": "2024-01-01T12:40:00Z"
  },
  "timestamp": "2024-01-01T12:40:00Z"
}
```

---

## 📝 Legacy Endpoints (To be removed)

### Tasks API
- `GET /api/tasks` - Get user tasks
- `POST /api/tasks` - Create task
- `POST /api/tasks/toggle?id=1` - Toggle task
- `POST /api/tasks/delete?id=1` - Delete task

### Notes API
- `GET /api/notes` - Get user notes
- `POST /api/notes` - Create note
- `PUT /api/notes/1` - Update note
- `DELETE /api/notes/1` - Delete note
- `POST /api/notes/pin?id=1` - Toggle pin

---

## ❌ Error Responses

### Common Error Format
```json
{
  "success": false,
  "error": "Error message here"
}
```

### HTTP Status Codes
- `200` - Success
- `400` - Bad Request (validation error)
- `401` - Unauthorized (not logged in)
- `403` - Forbidden (no permission)
- `404` - Not Found
- `409` - Conflict (duplicate resource)
- `500` - Internal Server Error

### Common Error Messages
- `"Not authenticated"` - User not logged in
- `"Invalid request"` - Malformed JSON or missing fields
- `"Conversation not found"` - Invalid conversation ID
- `"Permission denied"` - User doesn't have access
- `"Email already registered"` - Duplicate email during registration
- `"Invalid email or password"` - Login failed
- `"Token required"` - Missing token parameter
- `"Invalid or expired reset token"` - Password reset token invalid

---

## 🔧 Rate Limiting

### Current Limits
- Auth endpoints: 5 requests per minute per IP
- Message sending: 60 messages per minute per user
- API calls: 1000 requests per hour per user

### Rate Limit Headers
```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 59
X-RateLimit-Reset: 1640995200
```

---

## 📊 Response Times (Target)
- Auth endpoints: < 200ms
- Chat endpoints: < 100ms
- WebSocket latency: < 50ms
- File uploads: < 2s (depending on size)

---

*This API reference is automatically updated with each release.*