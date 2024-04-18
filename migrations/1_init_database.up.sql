CREATE SCHEMA cars;

CREATE TABLE car (
    id SERIAL PRIMARY KEY,
    mark VARCHAR(100) NOT NULL,
    model VARCHAR(100) NOT NULL,
    year INTEGER NOT NULL,
    reg_num VARCHAR(20) UNIQUE NOT NULL,
    owner_name VARCHAR(100),
    owner_surname VARCHAR(100),
    owner_patronymic VARCHAR(100)
);