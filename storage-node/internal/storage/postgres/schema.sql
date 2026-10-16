CREATE TABLE IF NOT EXISTS key_values (
    "user" TEXT NOT NULL,
    "key" TEXT NOT NULL,
    "value" TEXT NOT NULL,
    PRIMARY KEY ("user", "key")
);

CREATE TABLE IF NOT EXISTS consensus_state (
    id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    promised_round NUMERIC(20, 0) NOT NULL DEFAULT 0
        CHECK (promised_round BETWEEN 0 AND 18446744073709551615),
    promised_node_id TEXT NOT NULL DEFAULT ''
);

INSERT INTO consensus_state (id, promised_round, promised_node_id)
VALUES (1, 0, '')
ON CONFLICT (id) DO NOTHING;
