#!/bin/bash

# Seed script for healthcare system database
# This script populates the database with test users and sample data
# SECURITY: Only use this for development/testing

set -e

# Get connection parameters from environment
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-}
DB_NAME=${DB_NAME:-healthcare}

# Export password for psql
export PGPASSWORD=$DB_PASSWORD

echo "Seeding database: $DB_NAME on $DB_HOST:$DB_PORT"

# Run seed SQL
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME << 'EOF'

-- SECURITY: These are test credentials ONLY
-- For production, never hardcode credentials or use this seed approach

-- Insert admin user
INSERT INTO users (email, first_name, last_name, role, password_hash, created_at, updated_at)
VALUES (
    'admin@healthcare.local',
    'Admin',
    'User',
    'admin',
    '$2a$12$xxx', -- Will be replaced by actual bcrypt hash
    NOW(),
    NOW()
) ON CONFLICT (email) DO NOTHING;

-- Insert registrar user
INSERT INTO users (email, first_name, last_name, role, password_hash, created_at, updated_at)
VALUES (
    'registrar@healthcare.local',
    'Registrar',
    'Staff',
    'registrar',
    '$2a$12$xxx', -- Will be replaced by actual bcrypt hash
    NOW(),
    NOW()
) ON CONFLICT (email) DO NOTHING;

-- Insert doctor user
INSERT INTO users (email, first_name, last_name, role, password_hash, created_at, updated_at)
VALUES (
    'doctor@healthcare.local',
    'Dr.',
    'Smith',
    'doctor',
    '$2a$12$xxx', -- Will be replaced by actual bcrypt hash
    NOW(),
    NOW()
) ON CONFLICT (email) DO NOTHING;

-- Insert sample patient
INSERT INTO patients (first_name, last_name, email, phone, date_of_birth, gender, address, medical_info, registered_by, created_at, updated_at)
SELECT
    'John',
    'Doe',
    'john@example.local',
    '+1-555-0100',
    '1990-05-15',
    'M',
    '123 Main St, Anytown, USA',
    'No known allergies',
    id,
    NOW(),
    NOW()
FROM users
WHERE email = 'registrar@healthcare.local'
ON CONFLICT DO NOTHING;

COMMIT;
EOF

echo "Database seed completed"
