-- Migration 0027: create group_events and group_events_reaction tables

CREATE TABLE IF NOT EXISTS group_events (
    id INTEGER PRIMARY KEY,
    group_id INTEGER NOT NULL,
    creator_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    event_day DATE NOT NULL,
    event_time TIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    going INTEGER NOT NULL DEFAULT 0,
    not_going INTEGER NOT NULL DEFAULT 0,
    invited INTEGER NOT NULL DEFAULT 0,

    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_group_events_group_id ON group_events(group_id);
CREATE INDEX IF NOT EXISTS idx_group_events_creator_id ON group_events(creator_id);
CREATE INDEX IF NOT EXISTS idx_group_events_event_day ON group_events(event_day);

CREATE TABLE IF NOT EXISTS group_events_reaction (
    event_id INTEGER NOT NULL,
    group_id INTEGER NOT NULL,
    creator_id INTEGER NOT NULL,
    reactor_id INTEGER NOT NULL,
    reaction_type TEXT NOT NULL CHECK (reaction_type IN ('going', 'not_going', 'invited')),

    PRIMARY KEY (event_id, reactor_id),
    FOREIGN KEY (event_id) REFERENCES group_events(id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (reactor_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_group_events_reaction_event_id ON group_events_reaction(event_id);
CREATE INDEX IF NOT EXISTS idx_group_events_reaction_group_id ON group_events_reaction(group_id);
CREATE INDEX IF NOT EXISTS idx_group_events_reaction_creator_id ON group_events_reaction(creator_id);
CREATE INDEX IF NOT EXISTS idx_group_events_reaction_reactor_id ON group_events_reaction(reactor_id);
