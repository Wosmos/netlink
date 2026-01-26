# Database Migration Guide

## Adding conversation_id to Tasks and Notes

### Automatic Migration (Recommended)

The backend now automatically handles the migration when it starts:

1. **Stop the backend** if it's running
2. **Restart the backend**:
   ```bash
   cd go-backend/backend
   go run main.go
   # or if you have a compiled binary
   ./server
   ```

The `InitSchema()` methods will:
- Create tables if they don't exist
- Add `conversation_id` column if it doesn't exist
- Create necessary indexes
- Preserve all existing data

### Manual Migration (If Needed)

If you prefer to run the migration manually:

```bash
# Connect to your PostgreSQL database
psql -U your_username -d your_database

# Run the migration script
\i backend/migrations/add_conversation_id.sql
```

Or using psql command:
```bash
psql -U your_username -d your_database -f backend/migrations/add_conversation_id.sql
```

### Verification

After migration, verify the columns exist:

```sql
-- Check tasks table
SELECT column_name, data_type, is_nullable 
FROM information_schema.columns 
WHERE table_name = 'tasks';

-- Check notes table
SELECT column_name, data_type, is_nullable 
FROM information_schema.columns 
WHERE table_name = 'notes';
```

You should see `conversation_id` column in both tables.

### Data Impact

**Existing Data**:
- All existing tasks and notes will have `conversation_id = NULL`
- This means they are **personal** tasks/notes (not tied to any conversation)
- No data loss occurs
- Users can still access all their existing tasks and notes

**New Data**:
- Tasks/notes created from a conversation will have `conversation_id` set
- Tasks/notes created from the main page will have `conversation_id = NULL`

### Rollback (If Needed)

If you need to rollback the migration:

```sql
-- Remove conversation_id from tasks
ALTER TABLE tasks DROP COLUMN IF EXISTS conversation_id;

-- Remove conversation_id from notes
ALTER TABLE notes DROP COLUMN IF EXISTS conversation_id;

-- Remove indexes
DROP INDEX IF EXISTS idx_tasks_conversation;
DROP INDEX IF EXISTS idx_notes_conversation;
```

**Warning**: This will delete the conversation associations but preserve the tasks/notes themselves.

### Troubleshooting

#### Error: "column conversation_id does not exist"

**Cause**: The migration hasn't run yet or failed.

**Solution**:
1. Stop the backend
2. Restart it - the InitSchema will run automatically
3. Check logs for any errors

#### Error: "relation conversations does not exist"

**Cause**: The conversations table doesn't exist yet.

**Solution**:
1. Make sure the chat system is initialized first
2. The conversations table should be created before tasks/notes
3. Check the initialization order in `main.go`

#### Error: "permission denied"

**Cause**: Database user doesn't have ALTER TABLE permissions.

**Solution**:
```sql
-- Grant necessary permissions
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO your_username;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO your_username;
```

### Migration Status Check

To check if migration is complete:

```sql
-- Check if conversation_id exists in tasks
SELECT EXISTS (
  SELECT 1 
  FROM information_schema.columns 
  WHERE table_name='tasks' AND column_name='conversation_id'
) AS tasks_migrated;

-- Check if conversation_id exists in notes
SELECT EXISTS (
  SELECT 1 
  FROM information_schema.columns 
  WHERE table_name='notes' AND column_name='conversation_id'
) AS notes_migrated;
```

Both should return `true`.

### Performance Impact

**During Migration**:
- Very fast (< 1 second for most databases)
- No downtime required
- Non-blocking operation

**After Migration**:
- Indexes created for optimal query performance
- No performance degradation
- Queries may be slightly faster due to better indexing

### Best Practices

1. **Backup First**: Always backup your database before migrations
   ```bash
   pg_dump -U your_username your_database > backup_$(date +%Y%m%d).sql
   ```

2. **Test in Development**: Run migration in dev environment first

3. **Monitor Logs**: Watch backend logs during startup for any errors

4. **Verify Data**: Check that existing tasks/notes are still accessible

### Migration Timeline

1. **Before Migration**:
   - Tasks and notes are user-specific only
   - No conversation association

2. **After Migration**:
   - Existing tasks/notes remain personal (conversation_id = NULL)
   - New tasks/notes can be conversation-specific
   - Full backward compatibility

---

**Status**: ✅ Migration script ready
**Risk Level**: Low (backward compatible)
**Downtime Required**: None
**Data Loss Risk**: None
