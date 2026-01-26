-- Migration: Add conversation_id to tasks and notes tables
-- Date: 2026-01-26

-- Add conversation_id column to tasks table
ALTER TABLE tasks 
ADD COLUMN IF NOT EXISTS conversation_id INTEGER REFERENCES conversations(id) ON DELETE CASCADE;

-- Add index for conversation_id in tasks
CREATE INDEX IF NOT EXISTS idx_tasks_conversation ON tasks(conversation_id);

-- Add conversation_id column to notes table
ALTER TABLE notes 
ADD COLUMN IF NOT EXISTS conversation_id INTEGER REFERENCES conversations(id) ON DELETE CASCADE;

-- Add index for conversation_id in notes
CREATE INDEX IF NOT EXISTS idx_notes_conversation ON notes(conversation_id);

-- Update existing queries to handle NULL conversation_id
-- (No data migration needed - existing records will have NULL conversation_id, which means personal tasks/notes)
