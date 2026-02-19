-- Remove registros de hardware_history sem component (anteriores à migration 008).
-- Esses registros contêm apenas snapshots JSONB sem dados estruturados.
DELETE FROM hardware_history WHERE component IS NULL;

-- Garantir que novos registros sempre tenham component e change_type preenchidos.
ALTER TABLE hardware_history ALTER COLUMN component SET NOT NULL;
ALTER TABLE hardware_history ALTER COLUMN change_type SET NOT NULL;
