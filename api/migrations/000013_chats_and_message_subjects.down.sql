DROP INDEX IF EXISTS chats_updated_idx;
DROP INDEX IF EXISTS chat_messages_chat_created_idx;
DROP TABLE IF EXISTS chat_messages;
DROP TABLE IF EXISTS chats;

ALTER TABLE messages
    DROP COLUMN IF EXISTS subject;
