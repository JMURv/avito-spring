package db

const getUserByEmail = `
SELECT id, email, password_hash, role 
FROM users 
WHERE email = $1
`

const createUser = `
INSERT INTO users (email, password_hash, role)
VALUES ($1, $2, $3)
RETURNING id
`

const getPVZ = `
SELECT 
	p.id,
	p.city,
	p.created_at AS registration_date,
	r.id AS reception_id,
	r.created_at AS reception_date,
	r.status,
	pr.id AS product_id,
	pr.created_at AS product_date,
	pr.type
FROM pickup_points p
LEFT JOIN receptions r ON p.id = r.pickup_point_id
LEFT JOIN products pr ON r.id = pr.reception_id
WHERE r.created_at BETWEEN $1 AND $2
ORDER BY p.created_at
LIMIT $3 OFFSET $4
`

const createPVZ = `
INSERT INTO pickup_points (city)
VALUES ($1)
RETURNING id, created_at
`

const findLastReception = `
SELECT 
	id,
	pickup_point_id,
	status,
	created_at
FROM receptions 
WHERE pickup_point_id = $1 AND status = 'in_progress'
ORDER BY created_at DESC 
LIMIT 1
`

const findLastReceptionForUpdate = `
SELECT 
	id,
	status,
	created_at,
	pickup_point_id
FROM receptions 
WHERE pickup_point_id = $1 AND status = 'in_progress'
FOR UPDATE
`

const createReception = `
INSERT INTO receptions (pickup_point_id, status)
VALUES ($1, 'in_progress')
RETURNING id, status, created_at, pickup_point_id
`

const closeReception = `
UPDATE receptions 
SET status = 'closed', closed_at = NOW() 
WHERE id = $1
`

const addItemToReception = `
INSERT INTO products (reception_id, type)
VALUES ($1, $2)
RETURNING id, reception_id, type, created_at
`

const deleteLastProduct = `
DELETE FROM products 
WHERE id IN (
	SELECT id 
	FROM products 
	WHERE reception_id = $1 
	ORDER BY created_at DESC 
	LIMIT 1
)
`

const listPVZs = `
SELECT id, city, created_at
FROM pickup_points
`
