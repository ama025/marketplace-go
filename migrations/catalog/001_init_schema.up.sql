CREATE TABLE brands (
    id UUID PRIMARY KEY,
    title VARCHAR(200) NOT NULL
);

INSERT INTO 
    brands (id, title)
VALUES (
        'b0000001-0000-0000-0000-000000000001',
        'Polaris'
    );