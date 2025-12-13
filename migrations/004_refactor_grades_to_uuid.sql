-- Migration: Refactor Grades to use UUID
-- 1. Add ID to grades
-- 2. Migrate existing references in purchase_lots and sale_transactions
-- 3. Update schemas

-- A. Update Grades Table
-- We need to act carefully. 
-- 1. Add id column
ALTER TABLE grades ADD COLUMN id UUID DEFAULT gen_random_uuid();

-- 2. Make id not null (after population)
ALTER TABLE grades ALTER COLUMN id SET NOT NULL;

-- 3. Remove existing PK (name) and add constraint unique
ALTER TABLE grades DROP CONSTRAINT grades_pkey;
ALTER TABLE grades ADD CONSTRAINT grades_name_key UNIQUE (name);

-- 4. Add new PK
ALTER TABLE grades ADD CONSTRAINT grades_pkey PRIMARY KEY (id);


-- B. Update Purchase Lots Table
-- 1. Add grade_id column
ALTER TABLE purchase_lots ADD COLUMN grade_id UUID;

-- 2. Populate grade_id by looking up name
UPDATE purchase_lots 
SET grade_id = G.id
FROM grades G
WHERE purchase_lots.grade = G.name;

-- 3. Enforce Not Null (assuming all lots match a grade. If not, this fails, but valid app usage implies they do)
-- If data is inconsistent, we might need to handle nulls, but let's assume consisteny.
ALTER TABLE purchase_lots ALTER COLUMN grade_id SET NOT NULL;

-- 4. Drop old grade column
ALTER TABLE purchase_lots DROP COLUMN grade;

-- 5. Add FK constraint
ALTER TABLE purchase_lots ADD CONSTRAINT fk_purchase_lots_grade FOREIGN KEY (grade_id) REFERENCES grades(id);


-- C. Update Sale Transactions Table
-- 1. Add grade_id column
ALTER TABLE sale_transactions ADD COLUMN grade_id UUID;

-- 2. Populate grade_id
UPDATE sale_transactions 
SET grade_id = G.id
FROM grades G
WHERE sale_transactions.grade = G.name;

-- 3. Not Null
ALTER TABLE sale_transactions ALTER COLUMN grade_id SET NOT NULL;

-- 4. Drop old column
ALTER TABLE sale_transactions DROP COLUMN grade;

-- 5. Add FK
ALTER TABLE sale_transactions ADD CONSTRAINT fk_sale_transactions_grade FOREIGN KEY (grade_id) REFERENCES grades(id);
