ALTER TABLE wishlist_item_snags
    ADD COLUMN delivery_method TEXT NOT NULL DEFAULT 'in_person'
        CHECK (delivery_method IN ('in_person', 'mail')),
    ADD COLUMN tracking_number TEXT;
