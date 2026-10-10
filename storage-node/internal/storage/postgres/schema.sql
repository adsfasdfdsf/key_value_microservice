CREATE TABLE IF NOT EXISTS key_values (
    "user" TEXT NOT NULL,
    "key" TEXT NOT NULL,
    "value" TEXT NOT NULL,
    PRIMARY KEY ("user", "key")
);
