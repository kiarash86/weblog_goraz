CREATE TABLE
    board_shares (
        user_id INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
        board_id INTEGER NOT NULL REFERENCES boards (id) ON DELETE CASCADE,
        PRIMARY KEY (user_id, board_id)
    )