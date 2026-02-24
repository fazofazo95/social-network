-- Migration 0024: create chats, chat_participants and chat_messages tables

CREATE TABLE IF NOT EXISTS chats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL CHECK (type IN ('direct', 'group')),
    group_id INTEGER,
    user_low_id INTEGER,
    user_high_id INTEGER,
    created_by INTEGER NOT NULL,
    last_message_id INTEGER,
    last_message_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CHECK (
        (type = 'group' AND group_id IS NOT NULL AND user_low_id IS NULL AND user_high_id IS NULL)
        OR
        (type = 'direct' AND group_id IS NULL AND user_low_id IS NOT NULL AND user_high_id IS NOT NULL AND user_low_id < user_high_id)
    ),

    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (user_low_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (user_high_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_chats_group_unique
ON chats(group_id)
WHERE type = 'group';

CREATE UNIQUE INDEX IF NOT EXISTS idx_chats_direct_pair_unique
ON chats(user_low_id, user_high_id)
WHERE type = 'direct';

CREATE INDEX IF NOT EXISTS idx_chats_last_message_at ON chats(last_message_at);
CREATE INDEX IF NOT EXISTS idx_chats_created_at ON chats(created_at);

CREATE TABLE IF NOT EXISTS chat_messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    chat_id INTEGER NOT NULL,
    sender_id INTEGER NOT NULL,
    message_type TEXT NOT NULL DEFAULT 'text' CHECK (message_type IN ('text', 'image', 'text_image')),
    body TEXT,
    media_url TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    edited_at DATETIME,
    deleted_at DATETIME,

    CHECK (
        (message_type = 'text' AND body IS NOT NULL AND length(trim(body)) > 0)
        OR
        (message_type = 'image' AND media_url IS NOT NULL AND length(trim(media_url)) > 0)
        OR
        (message_type = 'text_image' AND body IS NOT NULL AND length(trim(body)) > 0 AND media_url IS NOT NULL AND length(trim(media_url)) > 0)
    ),

    FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE,
    FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_chat_messages_chat_id_id ON chat_messages(chat_id, id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_chat_id_created_at ON chat_messages(chat_id, created_at);
CREATE INDEX IF NOT EXISTS idx_chat_messages_sender_id_created_at ON chat_messages(sender_id, created_at);

CREATE TABLE IF NOT EXISTS chat_participants (
    chat_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    joined_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    left_at DATETIME,
    muted_until DATETIME,
    is_archived BOOLEAN NOT NULL DEFAULT 0,
    last_read_message_id INTEGER,

    PRIMARY KEY (chat_id, user_id),
    FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_chat_participants_user_active
ON chat_participants(user_id, left_at);

CREATE INDEX IF NOT EXISTS idx_chat_participants_chat_id
ON chat_participants(chat_id);
