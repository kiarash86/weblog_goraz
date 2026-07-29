CREATE TABLE
    comments (
        id SERIAL PRIMARY KEY,
        author_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        board_id INTEGER NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
        content Text NOT NULL DEFAULT 'comment'
    )