CREATE TABLE
    Boards (
        id SERIAL,
        authorID SERIAL
         FOREIGN KEY (authorID) REFERENCES Authors (id) ON DELETE CASCADE,
        name varchar(50),
        context varchar(1000),
        isPrivate boolean,
        imgPath varchar(100) OPTIMAL NULL
    )