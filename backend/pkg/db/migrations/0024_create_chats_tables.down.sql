-- Rollback 0024: drop chats, chat_participants and chat_messages tables

DROP TABLE IF EXISTS chat_participants;
DROP TABLE IF EXISTS chat_messages;
DROP TABLE IF EXISTS chats;
