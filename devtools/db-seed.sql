BEGIN;

INSERT INTO organizations (id, display_name) VALUES ('org', 'My Organization') ON CONFLICT DO NOTHING;

INSERT INTO service_accounts (organization_id, name, secret_hash, role, created_at, updated_at) VALUES (
    'org',
    'org_admin_sa',
    '$argon2id$v=19$m=16,t=2,p=1$WlBFUmxyMkJWakw4TUMxVw$NyRkqa3o0uaAHnp7XpjU5A', -- 123456
    'org_admin',
    NOW(),
    NOW()
) ON CONFLICT DO NOTHING;
INSERT INTO service_accounts (organization_id, name, secret_hash, role, created_at, updated_at) VALUES (
    'org',
    'admin_sa',
    '$argon2id$v=19$m=16,t=2,p=1$WlBFUmxyMkJWakw4TUMxVw$NyRkqa3o0uaAHnp7XpjU5A', -- 123456
    'admin',
    NOW(),
    NOW()
) ON CONFLICT DO NOTHING;

COMMIT;
