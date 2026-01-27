-- Migration: Add custom reactions support
-- Date: 2026-01-26

-- Add custom reaction columns to message_reactions
ALTER TABLE message_reactions 
ADD COLUMN IF NOT EXISTS is_custom BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS custom_url TEXT;

-- Create index for faster reaction queries
CREATE INDEX IF NOT EXISTS idx_reactions_message_emoji ON message_reactions(message_id, emoji);

-- Create materialized view for reaction counts (optional, for extreme performance)
-- This can be refreshed periodically or on-demand
CREATE MATERIALIZED VIEW IF NOT EXISTS message_reaction_counts AS
SELECT 
    message_id,
    emoji,
    is_custom,
    custom_url,
    COUNT(*) as count,
    ARRAY_AGG(user_id) as user_ids
FROM message_reactions
GROUP BY message_id, emoji, is_custom, custom_url;

-- Create index on materialized view
CREATE UNIQUE INDEX IF NOT EXISTS idx_reaction_counts_msg_emoji 
ON message_reaction_counts(message_id, emoji);

-- Function to refresh reaction counts for a specific message (for real-time updates)
CREATE OR REPLACE FUNCTION refresh_message_reaction_counts(msg_id INTEGER)
RETURNS void AS $$
BEGIN
    DELETE FROM message_reaction_counts WHERE message_id = msg_id;
    INSERT INTO message_reaction_counts
    SELECT 
        message_id,
        emoji,
        is_custom,
        custom_url,
        COUNT(*) as count,
        ARRAY_AGG(user_id) as user_ids
    FROM message_reactions
    WHERE message_id = msg_id
    GROUP BY message_id, emoji, is_custom, custom_url;
END;
$$ LANGUAGE plpgsql;
