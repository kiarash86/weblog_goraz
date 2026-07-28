CREATE TABLE
    Boards (
        id SERIAL PRIMARY KEY,
        author_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE
        name varchar(50)  NOT NULL DEFAULT TITLE,
        content text  NOT NULL DEFAULT TEXT,
        is_private boolean  NOT NULL DEFAULT FALSE,
        img_path varchar(100) 
    )