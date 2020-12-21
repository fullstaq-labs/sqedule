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

INSERT INTO applications (organization_id, id, created_at) VALUES (
    'org',
    'app1',
    NOW()
) ON CONFLICT DO NOTHING;
INSERT INTO application_major_versions (organization_id, application_id, version_number, created_at, updated_at) VALUES (
    'org',
    'app1',
    1,
    NOW(),
    NOW()
) ON CONFLICT DO NOTHING;
INSERT INTO application_minor_versions (organization_id, application_major_version_id, version_number, review_state, created_at, display_name) VALUES (
    'org',
    (SELECT id FROM application_major_versions WHERE organization_id = 'org' AND application_id = 'app1' AND version_number = 1 LIMIT 1),
    1,
    'approved',
    NOW(),
    'Application 1'
) ON CONFLICT DO NOTHING;

COMMIT;
