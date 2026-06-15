-- Default admin user: admin@dosulogi.com / Admin@123
-- bcrypt hash of Admin@123
INSERT INTO users (email, password, full_name, role)
SELECT 'admin@dosulogi.com',
       '$2a$10$qHRryOHURzxNn8k7biC6VOvcjiiCXoMiR/QWMR/b67uZSvqerFYrC',
       'System Admin',
       'admin'
WHERE NOT EXISTS (SELECT 1 FROM users WHERE email = 'admin@dosulogi.com');
