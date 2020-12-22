CREATE TABLE positions 
(
    id          BIGSERIAL       PRIMARY KEY,
    product_id  BIGSERIAL            REFERENCES products,
    qty         INTEGER         NOT NULL,
    price       INTEGER         NOT NULL
)