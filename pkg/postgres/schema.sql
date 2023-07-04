-- PostgresSQL
CREATE DATABASE playground;

CREATE TABLE users
(
    user_id SERIAL,
    name VARCHAR(30),
    email VARCHAR(30),
    password VARCHAR(30),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    modified_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_admin BOOLEAN,
    PRIMARY KEY(user_id)
);

CREATE TABLE products
(
    product_id    SERIAL,
    name  VARCHAR(30),
    price INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    modified_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(product_id)
);

CREATE TABLE user_products
(
    user_product_id SERIAL,
    CONSTRAINT fk_user_id
    FOREIGN KEY(user_id)
    REFERENCES users(user_id)
    ON DELETE CASCADE,
    CONSTRAINT fk_product_id
    FOREIGN KEY(product_id)
    REFERENCES products(product_id)
    ON DELETE CASCADE
)