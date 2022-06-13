INSERT INTO system_configs (id, created_at, updated_at, config_key, config_value)
VALUES ('default', datetime('now'), datetime('now'), 'default_tiup_home', '/home/tidb/.tiup');
INSERT INTO system_configs (id, created_at, updated_at, config_key, config_value)
VALUES ('em', datetime('now'), datetime('now'), 'em_tiup_home', '/home/tidb/.em');