CREATE TABLE scheduler (
                       id INTEGER PRIMARY KEY AUTOINCREMENT,
                       date VARCHAR(32) DEFAULT "",
                       title TEXT DEFAULT "",
                       comment TEXT DEFAULT "",
                       repeat VARCHAR(128) DEFAULT ""
);
CREATE INDEX scheduler_date ON scheduler (date);