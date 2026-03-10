-- Sample provider data with placeholder API keys
-- Replace the API keys with your actual keys before use
INSERT INTO providers (name, platform, api_key, base_url)
VALUES
    ('openai', 'openai', 'your-openai-api-key-here', 'https://api.openai.com/v1'),
    ('anthropic', 'anthropic', 'your-anthropic-api-key-here', 'https://api.anthropic.com/v1');

-- Sample model mappings
INSERT INTO models (name, provider_id, real_model)
VALUES
    ('gpt-3.5-turbo', (SELECT id FROM providers WHERE name='openai'), 'gpt-3.5-turbo'),
    ('gpt-4', (SELECT id FROM providers WHERE name='openai'), 'gpt-4'),
    ('claude-3-opus', (SELECT id FROM providers WHERE name='anthropic'), 'claude-3-opus-20240229'),
    ('claude-3-sonnet', (SELECT id FROM providers WHERE name='anthropic'), 'claude-3-sonnet-20240229');

-- Default API key placeholder - generate your own secure key
INSERT INTO api_keys (name, key, user_id)
VALUES ('default', 'generate-a-secure-uuid-here', 1);