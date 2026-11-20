-- +goose Up
INSERT INTO parts (name, display_order) VALUES
    ('Vocal',      1),
    ('Guitar',     2),
    ('Bass',       3),
    ('Keyboard',   4),
    ('Percussion', 5),
    ('Extra',      6),
    ('Technician', 7);

-- +goose Down
DELETE FROM parts
WHERE name IN ('Vocal', 'Guitar', 'Bass', 'Keyboard', 'Percussion', 'Extra', 'Technician');
