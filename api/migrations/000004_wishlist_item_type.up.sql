CREATE TYPE snack_type AS ENUM (
    'Candy',
    'Baked Goods',
    'Beverages',
    'Pantry',
    'Chips/Crackers'
);

ALTER TABLE wishlist_items
    ADD COLUMN type snack_type NOT NULL DEFAULT 'Candy';

ALTER TABLE wishlist_items
    DROP COLUMN url,
    DROP COLUMN priority;
