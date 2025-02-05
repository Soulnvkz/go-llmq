-- Drop indexes first
DROP INDEX IF EXISTS idx_messages_chat_id_created_at;
DROP INDEX IF EXISTS idx_chats_updated_at;

-- Drop tables
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS chats;

-- Drop ENUM type
DROP TYPE IF EXISTS message_role;