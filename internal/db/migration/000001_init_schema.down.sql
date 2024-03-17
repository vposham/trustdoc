-- Drop the triggers
DROP TRIGGER IF EXISTS update_documents_change_timestamp ON documents;
DROP TRIGGER IF EXISTS update_users_change_timestamp ON users;


-- Drop the tables
DROP TABLE IF EXISTS documents CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- Drop the trigger function
DROP FUNCTION IF EXISTS update_change_timestamp_column();
