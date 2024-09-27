CREATE TYPE company_type AS ENUM ('Corporations', 'NonProfit', 'Cooperative', 'Sole Proprietorship');

CREATE TABLE IF NOT EXISTS companies (
    id UUID PRIMARY KEY,
    name VARCHAR(15) UNIQUE NOT NULL,
    description TEXT,
    amount_of_employees INTEGER NOT NULL,
    registered BOOLEAN NOT NULL,
    type company_type NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- Case-insensitive index on name
CREATE UNIQUE INDEX idx_companies_name_lower ON companies (LOWER(name));

-- Partial index for non-deleted companies
CREATE INDEX idx_companies_not_deleted ON companies (id) WHERE deleted_at IS NULL;

-- Index on type
CREATE INDEX idx_companies_type ON companies (type);

-- Index on amount_of_employees
CREATE INDEX idx_companies_employees ON companies (amount_of_employees);

CREATE OR REPLACE FUNCTION update_company_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_company_updated_at_trigger
BEFORE UPDATE ON companies
FOR EACH ROW
EXECUTE FUNCTION update_company_updated_at();
