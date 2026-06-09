ALTER TABLE wishlist_item_snags
    DROP COLUMN IF EXISTS tracking_number,
    DROP COLUMN IF EXISTS delivery_method;
