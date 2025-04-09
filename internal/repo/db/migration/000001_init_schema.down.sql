DROP ROLE user_role;
DROP ROLE reception_status;
DROP ROLE allowed_city;
DROP ROLE product_type;

DROP TABLE users;
DROP TABLE pickup_points;
DROP TABLE receptions;
DROP TABLE products;

DROP INDEX idx_pickup_points_city;
DROP INDEX idx_receptions_pickup_point;
DROP INDEX idx_receptions_status;
DROP INDEX idx_products_reception;