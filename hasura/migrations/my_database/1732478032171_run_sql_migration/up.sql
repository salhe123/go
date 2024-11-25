-- Rename ticket_price to amount
ALTER TABLE tickets RENAME COLUMN ticket_price TO amount;

-- Add new columns
ALTER TABLE tickets
ADD COLUMN phoneNumber VARCHAR(15), -- Adjust the size based on your requirement
ADD COLUMN quantity INTEGER DEFAULT 1, -- Default quantity to 1, adjust as necessary
ADD COLUMN catchedTicket BOOLEAN DEFAULT FALSE; -- Default value as FALSE

-- Optionally, if you want to add a `created_at` column, if it doesn't exist
ALTER TABLE tickets
ADD COLUMN created_at TIMESTAMP DEFAULT now();
