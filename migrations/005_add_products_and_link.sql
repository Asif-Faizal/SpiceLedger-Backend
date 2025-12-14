-- Create products table
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Seed default product if none exists
DO $$
DECLARE
    default_id UUID := gen_random_uuid();
BEGIN
    IF NOT EXISTS (SELECT 1 FROM products WHERE name = 'Cardamom') THEN
        INSERT INTO products (id, name, description) VALUES (default_id, 'Cardamom', 'Default product');
    ELSE
        SELECT id INTO default_id FROM products WHERE name = 'Cardamom';
    END IF;

    -- Add product_id to grades
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='grades' AND column_name='product_id') THEN
        ALTER TABLE grades ADD COLUMN product_id UUID;
        UPDATE grades SET product_id = default_id WHERE product_id IS NULL;
        ALTER TABLE grades ALTER COLUMN product_id SET NOT NULL;
        ALTER TABLE grades ADD CONSTRAINT fk_grades_product FOREIGN KEY (product_id) REFERENCES products(id);
    END IF;

    -- Add product_id to purchase_lots
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='purchase_lots' AND column_name='product_id') THEN
        ALTER TABLE purchase_lots ADD COLUMN product_id UUID;
        UPDATE purchase_lots pl SET product_id = g.product_id FROM grades g WHERE pl.grade_id = g.id;
        ALTER TABLE purchase_lots ALTER COLUMN product_id SET NOT NULL;
        ALTER TABLE purchase_lots ADD CONSTRAINT fk_purchase_lots_product FOREIGN KEY (product_id) REFERENCES products(id);
    END IF;

    -- Add product_id to sale_transactions
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='sale_transactions' AND column_name='product_id') THEN
        ALTER TABLE sale_transactions ADD COLUMN product_id UUID;
        UPDATE sale_transactions s SET product_id = g.product_id FROM grades g WHERE s.grade_id = g.id;
        ALTER TABLE sale_transactions ALTER COLUMN product_id SET NOT NULL;
        ALTER TABLE sale_transactions ADD CONSTRAINT fk_sale_transactions_product FOREIGN KEY (product_id) REFERENCES products(id);
    END IF;
END $$;
