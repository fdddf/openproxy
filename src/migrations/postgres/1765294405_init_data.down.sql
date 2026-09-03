DELETE FROM api_keys WHERE name = 'default';
DELETE FROM models WHERE provider_id IN (SELECT id FROM providers WHERE name IN ('openai', 'anthropic'));
DELETE FROM providers WHERE name IN ('openai', 'anthropic');
