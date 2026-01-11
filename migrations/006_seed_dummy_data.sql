DO $$
DECLARE
    admin_id UUID;
    cardamom_id UUID;
    grade_bold_id UUID;
    grade_green_id UUID;
    grade_super_id UUID;
BEGIN
    SELECT id INTO admin_id FROM users WHERE email = 'admin@example.com';
    IF admin_id IS NULL THEN
        admin_id := gen_random_uuid();
        INSERT INTO users (id, email, name, role, password_hash, created_at, updated_at)
        VALUES (admin_id, 'admin@example.com', 'Admin User', 'admin', '$2a$10$pJSge4ksLkYEM4yAazFQjeuG/aZDfBcnA53j1jJriyIV8UuJEUBgC', NOW(), NOW())
        ON CONFLICT (email) DO UPDATE SET role = 'admin', updated_at = NOW();
        SELECT id INTO admin_id FROM users WHERE email = 'admin@example.com';
    END IF;

    SELECT id INTO cardamom_id FROM products WHERE name = 'Cardamom';
    IF cardamom_id IS NULL THEN
        cardamom_id := gen_random_uuid();
        INSERT INTO products (id, name, description) VALUES (cardamom_id, 'Cardamom', 'Seeded product');
    END IF;

    IF NOT EXISTS (SELECT 1 FROM grades WHERE name = 'Bold') THEN
        grade_bold_id := gen_random_uuid();
        INSERT INTO grades (id, name, description, product_id) VALUES (grade_bold_id, 'Bold', 'High quality grade', cardamom_id);
    ELSE
        SELECT id INTO grade_bold_id FROM grades WHERE name = 'Bold';
        UPDATE grades SET product_id = cardamom_id WHERE id = grade_bold_id;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM grades WHERE name = 'Green') THEN
        grade_green_id := gen_random_uuid();
        INSERT INTO grades (id, name, description, product_id) VALUES (grade_green_id, 'Green', 'Fresh green pods', cardamom_id);
    ELSE
        SELECT id INTO grade_green_id FROM grades WHERE name = 'Green';
        UPDATE grades SET product_id = cardamom_id WHERE id = grade_green_id;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM grades WHERE name = 'Super Bold') THEN
        grade_super_id := gen_random_uuid();
        INSERT INTO grades (id, name, description, product_id) VALUES (grade_super_id, 'Super Bold', 'Premium grade', cardamom_id);
    ELSE
        SELECT id INTO grade_super_id FROM grades WHERE name = 'Super Bold';
        UPDATE grades SET product_id = cardamom_id WHERE id = grade_super_id;
    END IF;
END $$;
