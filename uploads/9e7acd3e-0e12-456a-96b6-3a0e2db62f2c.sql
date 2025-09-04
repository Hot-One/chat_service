-- Roles table
CREATE TABLE IF NOT EXISTS "roles"(
    "id"            UUID PRIMARY KEY,
    "name"          VARCHAR(255),
    "created_at"    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"    TIMESTAMP
);

-- Users table
CREATE TABLE IF NOT EXISTS "users"(
    "id"            UUID PRIMARY KEY,
    "name"          VARCHAR(255),
    "surname"       VARCHAR(255),
    "patranomic"    VARCHAR(255),
    "pinfl"         VARCHAR(255),
    "birth_date"    DATE,
    "issued_date"   DATE,
    "issued_by"     VARCHAR(255),
    "passport"      VARCHAR(255),
    "created_at"    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"    TIMESTAMP
);

-- User roles table
-- This table links users to their roles
-- It allows a user to have multiple roles
-- and a role to be assigned to multiple users
-- The "ON DELETE SET NULL" ensures that if a role or user is deleted,
-- the corresponding foreign key in this table will be set to NULL
-- instead of deleting the record, preserving the relationship history
-- between users and roles
CREATE TABLE IF NOT EXISTS "user_roles" (
    "id"            UUID PRIMARY KEY,
    "roles_id"      UUID REFERENCES roles(id) ON DELETE SET NULL,
    "users_id"      UUID REFERENCES users(id) ON DELETE SET NULL,
    "created_at"    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"    TIMESTAMP
);

-- Admins table
CREATE TABLE IF NOT EXISTS "admins"(
    "id"            UUID PRIMARY KEY,
    "users_id"      UUID REFERENCES users(id) ON DELETE SET NULL,
    "email"         VARCHAR(255) NOT NULL,
    "tg_link"       VARCHAR(255) NOT NULL,
    "created_at"    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"    TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "companies"(
    "id"            UUID PRIMARY KEY,
    "name"          VARCHAR(255),
    "created_at"    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"    TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "company_users" (
    "id"            UUID PRIMARY KEY,
    "companies_id"  UUID REFERENCES companies(id) ON DELETE SET NULL,
    "users_id"      UUID REFERENCES users(id) ON DELETE SET NULL,
    "created_at"    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"    TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "drivers" (
    "id"                                UUID PRIMARY KEY,
    "users_id"                          UUID REFERENCES users(id) ON DELETE SET NULL,
    "companies_id"                      UUID REFERENCES companies(id) ON DELETE SET NULL,
    "status"                            VARCHAR(50),
    "driver_license_file"               VARCHAR(255),
    "driver_license_expiry_date"        DATE,
    "medical_certificate_file"          VARCHAR(255),
    "medical_certificate_expiry_date"   DATE,
    "adr_certificate_file"              VARCHAR(255),
    "adr_certificate_expiry_date"       DATE,
    "passport_file"                     VARCHAR(255),
    "passport_expiry_date"              DATE,
    "created_at"                        TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"                        TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "company_drivers" (
    "id"            UUID PRIMARY KEY,
    "drivers_id"    UUID REFERENCES drivers(id) ON DELETE SET NULL,
    "created_at"    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"    TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "trucks" (
    "id"            UUID PRIMARY KEY,
    "companies_id"  UUID REFERENCES companies(id) ON DELETE SET NULL,
    "drivers_id"    UUID REFERENCES drivers(id) ON DELETE SET NULL,
    "model"         VARCHAR(255),
    "number"        VARCHAR(50),
    "created_at"    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"    TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "statuses" (
    "id"            UUID PRIMARY KEY,
    "name"          VARCHAR(255),
    "created_at"    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"    TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "truck_types" (
    "id"            UUID PRIMARY KEY,
    "name"          VARCHAR(255),
    "photo"         VARCHAR(255),
    "lifting_power" NUMERIC(12,2),
    "body_volume"   NUMERIC(12,2),
    "load_capacity" NUMERIC(12,2),
    "created_at"    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"    TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "truck_forms" (
    "id"                    UUID PRIMARY KEY,
    "name"                  VARCHAR(255),
    "temperature_control"   BOOLEAN DEFAULT FALSE,
    "is_pallet"             BOOLEAN DEFAULT FALSE,
    "created_at"            TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"            TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "loading_types" (
    "id"            UUID PRIMARY KEY,
    "name"          VARCHAR(255),
    "created_at"    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"    TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "truck_category_forms" (
    "id"                    UUID PRIMARY KEY,
    "truck_types_id"        UUID REFERENCES truck_types(id) ON DELETE SET NULL,
    "truck_forms_id"        UUID REFERENCES truck_forms(id) ON DELETE SET NULL,
    "created_at"            TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"            TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "truck_form_loading_types" (
    "id"                UUID PRIMARY KEY,
    "truck_forms_id"    UUID REFERENCES truck_forms(id) ON DELETE SET NULL,
    "loading_types_id"  UUID REFERENCES loading_types(id) ON DELETE SET NULL,
    "created_at"        TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"        TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "additional_services" (
    "id"            UUID PRIMARY KEY,
    "name"          VARCHAR(255),
    "created_at"    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"    TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "required_documents" (
    "id"            UUID PRIMARY KEY,
    "name"          VARCHAR(255),
    "is_free"       BOOLEAN DEFAULT FALSE,
    "created_at"    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"    TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "delivery_conditions" (
    "id"            UUID PRIMARY KEY,
    "code"          VARCHAR(50) UNIQUE NOT NULL,
    "name"          VARCHAR(255),
    "description"   TEXT,
    "created_at"    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"    TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "adr" (
    "id"            UUID PRIMARY KEY,
    "name"          VARCHAR(255),
    "created_at"    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"    TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "packages" (
    "id"            UUID PRIMARY KEY,
    "name"          VARCHAR(255),
    "created_at"    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"    TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "orders" (
    "id"            UUID PRIMARY KEY,
    "orders_number" VARCHAR(50) UNIQUE NOT NULL,
    "users_id"      UUID REFERENCES users(id) ON DELETE SET NULL,
    "statuses_id"   UUID REFERENCES statuses(id) ON DELETE SET NULL,
    
    "from_address"  VARCHAR(255),
    "to_address"    VARCHAR(255),

    "from_lat"      NUMERIC(9,6),
    "from_long"     NUMERIC(9,6),
    
    "to_lat"        NUMERIC(9,6),
    "to_long"       NUMERIC(9,6),

    "from_date"    TIMESTAMP,
    "to_date"      TIMESTAMP,

    "photos"         TEXT[],
    "files"          TEXT[],
    "description"    TEXT,
    "phone_number"   VARCHAR(50),
    "offer_accepted" BOOLEAN DEFAULT FALSE,

    "truck_types_id"            UUID REFERENCES truck_types(id) ON DELETE SET NULL,
    "truck_forms_id"            UUID REFERENCES truck_forms(id) ON DELETE SET NULL,
    "loading_types_id"          UUID REFERENCES loading_types(id) ON DELETE SET NULL,
    "delivery_conditions_id"    UUID REFERENCES delivery_conditions(id) ON DELETE SET NULL,

    "created_at"    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"    TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "order_additional_services" (
    "id"            UUID PRIMARY KEY,
    "orders_id"     UUID REFERENCES orders(id) ON DELETE SET NULL,
    "additional_services_id" UUID REFERENCES additional_services(id) ON DELETE SET NULL,
    "created_at"    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"    TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "order_required_documents" (
    "id"                    UUID PRIMARY KEY,
    "orders_id"             UUID REFERENCES orders(id) ON DELETE SET NULL,
    "required_documents_id" UUID REFERENCES required_documents(id) ON DELETE SET NULL,
    "created_at"            TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"            TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "order_adr" (
    "id"            UUID PRIMARY KEY,
    "orders_id"     UUID REFERENCES orders(id) ON DELETE SET NULL,
    "adr_id"        UUID REFERENCES adr(id) ON DELETE SET NULL,
    "created_at"    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"    TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "weight_types" (
    "id"            UUID PRIMARY KEY,
    "name"          VARCHAR(255),
    "created_at"    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"    TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "currencies" (
    "id"            UUID PRIMARY KEY,
    "name"          VARCHAR(255),
    "symbol"        VARCHAR(10),
    "rate"          NUMERIC(12,2),
    "created_at"    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"    TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "currency_rates" (
    "id"            UUID PRIMARY KEY,
    "currencies_id" UUID REFERENCES currencies(id) ON DELETE SET NULL,
    "rate"          NUMERIC(12,2),
    "date"          DATE,
    "created_at"    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"    TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "order_products" (
    "id"                UUID PRIMARY KEY,
    "description"       TEXT,
    "price"             NUMERIC(12,2),
    "weight"            NUMERIC(12,2),
    "orders_id"         UUID REFERENCES orders(id) ON DELETE SET NULL,
    "package_id"        UUID REFERENCES packages(id) ON DELETE SET NULL,
    "weight_types_id"   UUID REFERENCES weight_types(id) ON DELETE SET NULL,
    "currencies_id"     UUID REFERENCES currencies(id) ON DELETE SET NULL,
    "created_at"        TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"        TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "order_offers" (
    "id"                UUID PRIMARY KEY,
    "orders_id"         UUID REFERENCES orders(id) ON DELETE SET NULL,
    "companies_id"      UUID REFERENCES companies(id) ON DELETE SET NULL,
    "users_id"          UUID REFERENCES users(id) ON DELETE SET NULL,
    "drivers_id"        UUID REFERENCES drivers(id) ON DELETE SET NULL,
    "trucks_id"         UUID REFERENCES trucks(id) ON DELETE SET NULL,
    "price"             NUMERIC(12,2),
    "description"       TEXT,
    "created_at"        TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"        TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "order_history" (
    "id"            UUID PRIMARY KEY,
    "orders_id"     UUID REFERENCES orders(id) ON DELETE SET NULL,
    "statuses_id"   UUID REFERENCES statuses(id) ON DELETE SET NULL,
    "users_id"      UUID REFERENCES users(id) ON DELETE SET NULL,
    "roles_id"      UUID REFERENCES roles(id) ON DELETE SET NULL,
    "description"   TEXT,
    "photo"         TEXT,
    "created_at"    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"    TIMESTAMP
);
