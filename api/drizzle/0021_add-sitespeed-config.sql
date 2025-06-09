-- Custom SQL migration file, put your code below! --

-- Add sitespeed configuration columns to shop table
ALTER TABLE shop ADD COLUMN sitespeed_enabled INTEGER DEFAULT 0 NOT NULL;
ALTER TABLE shop ADD COLUMN sitespeed_urls TEXT DEFAULT '[]' NOT NULL;

-- Add url and label columns to shop_sitespeed table
ALTER TABLE shop_sitespeed ADD COLUMN url TEXT;
ALTER TABLE shop_sitespeed ADD COLUMN label TEXT;

-- Update existing rows to have the shop URL as default
UPDATE shop_sitespeed 
SET url = (SELECT url FROM shop WHERE shop.id = shop_sitespeed.shop_id),
    label = 'Homepage'
WHERE url IS NULL;

-- Make url and label columns required after update
-- SQLite doesn't support ALTER COLUMN, so we need to recreate the table
CREATE TABLE shop_sitespeed_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    shop_id INTEGER REFERENCES shop(id),
    url TEXT NOT NULL,
    label TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    ttfb INTEGER,
    fully_loaded INTEGER,
    largest_contentful_paint INTEGER,
    first_contentful_paint INTEGER,
    cumulative_layout_shift INTEGER,
    speed_index INTEGER,
    transfer_size INTEGER
);

-- Copy data to new table
INSERT INTO shop_sitespeed_new SELECT 
    id, 
    shop_id, 
    url, 
    label, 
    created_at, 
    ttfb, 
    fully_loaded, 
    largest_contentful_paint, 
    first_contentful_paint, 
    cumulative_layout_shift, 
    speed_index, 
    transfer_size 
FROM shop_sitespeed;

-- Drop old table and rename new one
DROP TABLE shop_sitespeed;
ALTER TABLE shop_sitespeed_new RENAME TO shop_sitespeed;