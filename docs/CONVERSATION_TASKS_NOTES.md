# Conversation-Specific Tasks & Notes

## Overview

Tasks and Notes are now conversation-specific! Each chat or group can have its own tasks and notes, making collaboration much more organized.

## Features

### 1. Personal vs Conversation Tasks/Notes

**Personal** (conversation_id = null):
- Your own tasks and notes
- Not tied to any conversation
- Visible when no conversation is selected

**Conversation-Specific** (conversation_id = X):
- Shared with conversation members
- Tied to a specific chat or group
- Visible when that conversation is active

### 2. Automatic Context Switching

When you:
- Open a conversation → See that conversation's tasks/notes
- Go to main chat page → See your personal tasks/notes
- Switch conversations → Tasks/notes update automatically

## Database Schema

### Tasks Table
```sql
CREATE TABLE tasks (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id),
  conversation_id INTEGER REFERENCES conversations(id), -- NULL for personal
  text TEXT NOT NULL,
  completed BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Notes Table
```sql
CREATE TABLE notes (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id),
  conversation_id INTEGER REFERENCES conversations(id), -- NULL for personal
  title VARCHAR(255) NOT NULL DEFAULT '',
  content TEXT NOT NULL DEFAULT '',
  color VARCHAR(20) DEFAULT '#ffffff',
  pinned BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## API Endpoints

### Tasks

**Get Tasks**:
```
GET /api/tasks                          # Personal tasks
GET /api/tasks?conversation_id=123     # Conversation tasks
```

**Create Task**:
```json
POST /api/tasks
{
  "text": "Task description",
  "conversation_id": 123  // Optional
}
```

**Toggle/Delete** (same as before):
```
POST /api/tasks/toggle?id=1
POST /api/tasks/delete?id=1
```

### Notes

**Get Notes**:
```
GET /api/notes                          # Personal notes
GET /api/notes?conversation_id=123     # Conversation notes
```

**Create Note**:
```json
POST /api/notes
{
  "title": "Note title",
  "content": "Note content",
  "conversation_id": 123  // Optional
}
```

**Update/Delete/Pin** (same as before):
```
PUT /api/notes/1
DELETE /api/notes/1
POST /api/notes/pin?id=1
```

## Frontend Implementation

### API Client (`lib/api.ts`)

```typescript
// Tasks
getTasks: (conversationId?: number) => ...
createTask: (text: string, conversationId?: number) => ...

// Notes
getNotes: (conversationId?: number) => ...
createNote: (title: string, content: string, conversationId?: number) => ...
```

### Panels

**NotesPanel.tsx**:
```typescript
interface NotesPanelProps {
  onClose: () => void;
  conversationId?: number; // Optional conversation ID
}
```

**TasksPanel.tsx**:
```typescript
interface TasksPanelProps {
  onClose: () => void;
  conversationId?: number; // Optional conversation ID
}
```

### Sidebar Integration

The Sidebar automatically:
1. Detects current conversation from URL
2. Passes conversation ID to panels
3. Updates when conversation changes

```typescript
// Extract conversation ID from pathname
useEffect(() => {
  const match = pathname?.match(/^\/chat\/(\d+)$/);
  if (match) {
    setCurrentConversationId(parseInt(match[1]));
  } else {
    setCurrentConversationId(undefined);
  }
}, [pathname]);

// Pass to panels
<NotesPanel conversationId={currentConversationId} />
<TasksPanel conversationId={currentConversationId} />
```

## User Experience

### Scenario 1: Personal Tasks
```
User on main chat page
├── Opens Tasks panel
├── Sees personal tasks only
└── Creates task → Saved as personal
```

### Scenario 2: Conversation Tasks
```
User in conversation #123
├── Opens Tasks panel
├── Sees tasks for conversation #123
└── Creates task → Saved with conversation_id=123
```

### Scenario 3: Switching Conversations
```
User in conversation #123
├── Opens Tasks panel → Sees tasks for #123
├── Switches to conversation #456
└── Tasks panel updates → Shows tasks for #456
```

## Migration

### Existing Data
All existing tasks and notes will have `conversation_id = NULL`, making them personal tasks/notes.

### Database Migration
The schema changes are backward compatible:
- New column `conversation_id` is nullable
- Existing data remains valid
- No data loss

### Steps to Deploy
1. Backend: Rebuild Go server (schema auto-updates)
2. Frontend: No special steps needed
3. Existing tasks/notes become personal automatically

## Benefits

### For Users
- **Organization**: Tasks/notes grouped by conversation
- **Collaboration**: Share tasks with group members
- **Context**: See relevant tasks when chatting
- **Flexibility**: Still have personal tasks/notes

### For Teams
- **Project Management**: Each project chat has its own tasks
- **Meeting Notes**: Notes tied to specific discussions
- **Shared To-Dos**: Everyone sees the same tasks
- **Accountability**: Track who created what

## Use Cases

### 1. Team Project
```
Project Chat #123
├── Tasks: "Design mockups", "Review code", "Deploy"
└── Notes: "Meeting notes", "Requirements", "API docs"
```

### 2. Family Group
```
Family Chat #456
├── Tasks: "Buy groceries", "Pick up kids", "Plan vacation"
└── Notes: "Shopping list", "Vacation ideas", "Recipes"
```

### 3. Personal Organization
```
No conversation selected
├── Tasks: "Personal goals", "Workout", "Read book"
└── Notes: "Journal", "Ideas", "Reminders"
```

## Future Enhancements

### Short Term
1. **Task Assignment**: Assign tasks to specific members
2. **Due Dates**: Add deadlines to tasks
3. **Task Comments**: Discuss tasks within the conversation
4. **Note Sharing**: Share notes between conversations

### Long Term
1. **Task Templates**: Reusable task lists
2. **Note Collaboration**: Real-time collaborative editing
3. **Task Dependencies**: Link related tasks
4. **Analytics**: Task completion rates, productivity metrics
5. **Integrations**: Sync with external tools (Trello, Notion, etc.)

## Technical Details

### Backend Changes

**Models**:
- Added `ConversationID *int` to Task and Note models
- Changed `Pinned` to `is_pinned` in Note model (matches frontend)

**Repositories**:
- Added `GetByConversation()` methods
- Added `CreateForConversation()` methods
- Updated queries to filter by conversation_id

**Handlers**:
- Added conversation_id parameter support
- Updated request/response structures
- Maintained backward compatibility

### Frontend Changes

**API Client**:
- Added optional conversationId parameters
- Updated all task/note methods

**Components**:
- Panels accept optional conversationId prop
- Auto-reload when conversationId changes
- Sidebar extracts conversationId from URL

## Testing Checklist

### Personal Tasks/Notes
- [ ] Can create personal tasks
- [ ] Can create personal notes
- [ ] Personal items show when no conversation selected
- [ ] Personal items don't show in conversations

### Conversation Tasks/Notes
- [ ] Can create tasks in conversation
- [ ] Can create notes in conversation
- [ ] Items show only in their conversation
- [ ] Items update when switching conversations

### CRUD Operations
- [ ] Create works for both personal and conversation
- [ ] Read filters correctly by conversation
- [ ] Update works regardless of conversation
- [ ] Delete works regardless of conversation
- [ ] Toggle/Pin works correctly

### UI/UX
- [ ] Panels open/close smoothly
- [ ] Context switches automatically
- [ ] No data loss when switching
- [ ] Loading states work correctly
- [ ] Error handling works

## Troubleshooting

### Tasks/Notes Not Showing
1. Check browser console for errors
2. Verify API response format
3. Check conversation_id in database
4. Verify user permissions

### Wrong Tasks/Notes Showing
1. Check URL for correct conversation ID
2. Verify Sidebar extracts ID correctly
3. Check API query parameters
4. Verify database filtering

### Can't Create Tasks/Notes
1. Check authentication
2. Verify API endpoint
3. Check request payload
4. Verify database constraints

---

**Status**: ✅ Implemented and Ready
**Version**: 2.0.0
**Breaking Changes**: None (backward compatible)
