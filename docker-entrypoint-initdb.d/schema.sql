CREATE TABLE customers
(
    id          BIGSERIAL           PRIMARY KEY,
    name        TEXT                NOT NULL,
    phone       TEXT                NOT NULL UNIQUE,
    password    TEXT                NOT NULL,
    active      BOOLEAN             NOT NULL DEFAULT TRUE,
    created     TIMESTAMP           NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE managers
(
    id          BIGSERIAL           PRIMARY KEY,
    name        TEXT                NOT NULL,
    salary      INTEGER             NOT NULL CHECK ( salary > 0 ),
    plan        INTEGER             NOT NULL CHECK ( salary > 0 ),
    boss_id     BIGINT              REFERENCES managers,
    department  TEXT,
    login       TEXT                NOT NULL UNIQUE,
    phone       TEXT                NOT NULL UNIQUE,
    roles       TEXT                NOT NULL,
    password    TEXT                NOT NULL,
    active      BOOLEAN             NOT NULL DEFAULT TRUE,
    created     TIMESTAMP           NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE customers_tokens
(
    token       TEXT                NOT NULL UNIQUE,
    customer_id BIGINT              NOT NULL REFERENCES customers,
    expire      TIMESTAMP           NOT NULL DEFAULT CURRENT_TIMESTAMP + INTERVAL '1 hour',
    created     TIMESTAMP           NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE products
(
    id          SERIAL           PRIMARY KEY,
    name        TEXT                NOT NULL,
    price       INTEGER             NOT NULL CHECK ( price > 0 ),
    qty         INTEGER             NOT NULL,
    active      BOOLEAN             NOT NULL DEFAULT TRUE
);

CREATE TABLE purchases
(
    id          BIGSERIAL           PRIMARY KEY,
    product_id  BIGSERIAL            REFERENCES products,
    name        TEXT                NOT NULL,
    qty         INTEGER             NOT NULL,
    created     TIMESTAMP           NOT NULL DEFAULT CURRENT_TIMESTAMP,
    customer_id BIGINT              REFERENCES customers,
    price       INTEGER             NOT NULL CHECK (price > 0)
);

CREATE TABLE users
(
    id          BIGSERIAL           PRIMARY KEY,
    name        TEXT                NOT NULL,
    phone       TEXT                NOT NULL UNIQUE,
    password    TEXT,
    salary      INTEGER             NOT NULL DEFAULT 0,
    roles       TEXT[]              NOT NULL DEFAULT '{}',
    active      BOOLEAN             NOT NULL DEFAULT TRUE,
    created     TIMESTAMP           NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE managers_tokens
(
    token       TEXT        NOT NULL UNIQUE,
    manager_id  BIGINT      NOT NULL REFERENCES users,
    expire      TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP + INTERVAL '1 hour',
    created     TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE sales 
(
    id          BIGSERIAL       PRIMARY KEY,
    manager_id  BIGINT          NOT NULL REFERENCES users,
    customer_id BIGINT          NOT NULL,
    created     TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE sale_positions
(
    id          BIGSERIAL           PRIMARY KEY,
    sale_id     BIGINT              NOT NULL REFERENCES sales,
    product_id  BIGINT              NOT NULL REFERENCES products,
    price       INTEGER             NOT NULL CHECK ( price >= 0 ),
    qty         INTEGER             NOT NULL DEFAULT 0 CHECK ( qty >= 0),
    created     TIMESTAMP           NOT NULL DEFAULT CURRENT_TIMESTAMP
);