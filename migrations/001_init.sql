CREATE TABLE IF NOT EXISTS klines (
    symbol VARCHAR(20) NOT NULL,
    interval VARCHAR(5) NOT NULL,
    open_time TIMESTAMPTZ NOT NULL,
    open VARCHAR(40) NOT NULL,
    high VARCHAR(40) NOT NULL,
    low VARCHAR(40) NOT NULL,
    close VARCHAR(40) NOT NULL,
    volume VARCHAR(40) NOT NULL DEFAULT '0',
    close_time TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (symbol, interval, open_time)
);

CREATE INDEX idx_klines_symbol_interval ON klines(symbol, interval, open_time DESC);

CREATE TABLE IF NOT EXISTS trades (
    id VARCHAR(64) PRIMARY KEY,
    symbol VARCHAR(20) NOT NULL,
    price VARCHAR(40) NOT NULL,
    quantity VARCHAR(40) NOT NULL,
    maker_side VARCHAR(4) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_trades_symbol ON trades(symbol, created_at DESC);
