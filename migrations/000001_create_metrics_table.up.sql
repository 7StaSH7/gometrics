CREATE TABLE IF NOT EXISTS metrics (
  id TEXT NOT NULL,
  mType TEXT NOT NULL,
  delta BIGINT,
  value DOUBLE PRECISION,
  hash TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS metrics_id_uindex ON metrics (id);
