DROP TRIGGER IF EXISTS update_company_updated_at_trigger ON companies;
DROP FUNCTION IF EXISTS update_company_updated_at();
DROP INDEX IF EXISTS idx_companies_name_lower;
DROP INDEX IF EXISTS idx_companies_not_deleted;
DROP INDEX IF EXISTS idx_companies_type;
DROP INDEX IF EXISTS idx_companies_employees;
DROP TABLE IF EXISTS companies;
DROP TYPE IF EXISTS company_type;
