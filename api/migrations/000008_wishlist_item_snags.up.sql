CREATE TABLE wishlist_item_snags (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wishlist_item_id    UUID NOT NULL REFERENCES wishlist_items(id) ON DELETE CASCADE,
    snagged_by_user_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    snagged_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (wishlist_item_id)
);

CREATE INDEX idx_wishlist_item_snags_snagged_by_user_id ON wishlist_item_snags(snagged_by_user_id);
