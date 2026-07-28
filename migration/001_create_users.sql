CREATE TABLE
    users (
        id SERIAL PRIMARY KEY,
        name varchar(50) UNIQUE NOT NULL DEFAULT name,
        password text NOT NULL
    )