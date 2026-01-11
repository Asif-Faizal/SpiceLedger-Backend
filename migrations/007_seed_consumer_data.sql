DO $$
DECLARE
    admin_id UUID;
    consumer_id UUID;
    cardamom_id UUID;
    grade_bold_id UUID;
    grade_green_id UUID;
    grade_super_id UUID;
BEGIN
    SELECT id INTO admin_id FROM users WHERE email = 'admin@example.com';

    SELECT id INTO consumer_id FROM users WHERE email = 'consumer@example.com';
    IF consumer_id IS NULL THEN
        consumer_id := gen_random_uuid();
        INSERT INTO users (id, email, name, role, password_hash, created_at, updated_at)
        VALUES (consumer_id, 'consumer@example.com', 'Consumer User', 'user', '$2a$10$pJSge4ksLkYEM4yAazFQjeuG/aZDfBcnA53j1jJriyIV8UuJEUBgC', NOW(), NOW())
        ON CONFLICT (email) DO NOTHING;
        SELECT id INTO consumer_id FROM users WHERE email = 'consumer@example.com';
    END IF;

    SELECT id INTO cardamom_id FROM products WHERE name = 'Cardamom';
    IF cardamom_id IS NULL THEN
        cardamom_id := gen_random_uuid();
        INSERT INTO products (id, name, description) VALUES (cardamom_id, 'Cardamom', 'Default product');
    END IF;

    SELECT id INTO grade_bold_id FROM grades WHERE name = 'Bold';
    SELECT id INTO grade_green_id FROM grades WHERE name = 'Green';
    SELECT id INTO grade_super_id FROM grades WHERE name = 'Super Bold';

    UPDATE purchase_lots SET user_id = consumer_id WHERE user_id = admin_id;
    UPDATE sale_transactions SET user_id = consumer_id WHERE user_id = admin_id;

    INSERT INTO purchase_lots (id, user_id, date, quantity, unit_cost, product_id, grade_id, created_at, updated_at)
    VALUES
        (gen_random_uuid(), consumer_id, CURRENT_DATE - INTERVAL '2 days', 200.00, 940.00, cardamom_id, grade_bold_id, NOW(), NOW()),
        (gen_random_uuid(), consumer_id, CURRENT_DATE - INTERVAL '2 days', 150.00, 905.00, cardamom_id, grade_green_id, NOW(), NOW()),
        (gen_random_uuid(), consumer_id, CURRENT_DATE - INTERVAL '1 day', 180.00, 980.00, cardamom_id, grade_super_id, NOW(), NOW()),
        (gen_random_uuid(), consumer_id, CURRENT_DATE - INTERVAL '1 day', 90.00, 960.00, cardamom_id, grade_bold_id, NOW(), NOW()),
        (gen_random_uuid(), consumer_id, CURRENT_DATE, 120.00, 995.00, cardamom_id, grade_bold_id, NOW(), NOW()),
        (gen_random_uuid(), consumer_id, CURRENT_DATE, 110.00, 915.00, cardamom_id, grade_green_id, NOW(), NOW());

    INSERT INTO sale_transactions (id, user_id, date, quantity, unit_price, product_id, grade_id, created_at, updated_at)
    VALUES
        (gen_random_uuid(), consumer_id, CURRENT_DATE - INTERVAL '1 day', 70.00, 1110.00, cardamom_id, grade_bold_id, NOW(), NOW()),
        (gen_random_uuid(), consumer_id, CURRENT_DATE - INTERVAL '1 day', 30.00, 1090.00, cardamom_id, grade_green_id, NOW(), NOW()),
        (gen_random_uuid(), consumer_id, CURRENT_DATE, 60.00, 1130.00, cardamom_id, grade_super_id, NOW(), NOW()),
        (gen_random_uuid(), consumer_id, CURRENT_DATE, 50.00, 1120.00, cardamom_id, grade_bold_id, NOW(), NOW());
END $$;

