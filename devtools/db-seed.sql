BEGIN;


-- Organizations
INSERT INTO organizations (id, display_name) VALUES ('org1', 'My Organization') ON CONFLICT DO NOTHING;
INSERT INTO organizations (id, display_name) VALUES ('org2', 'My Organization 2') ON CONFLICT DO NOTHING;


-- Organization members for org1
INSERT INTO service_accounts (organization_id, name, secret_hash, role, created_at, updated_at) VALUES (
    'org1',
    'org_admin_sa',
    '$argon2id$v=19$m=16,t=2,p=1$WlBFUmxyMkJWakw4TUMxVw$NyRkqa3o0uaAHnp7XpjU5A', -- 123456
    'org_admin',
    NOW(),
    NOW()
) ON CONFLICT DO NOTHING;
INSERT INTO service_accounts (organization_id, name, secret_hash, role, created_at, updated_at) VALUES (
    'org1',
    'admin_sa',
    '$argon2id$v=19$m=16,t=2,p=1$WlBFUmxyMkJWakw4TUMxVw$NyRkqa3o0uaAHnp7XpjU5A', -- 123456
    'admin',
    NOW(),
    NOW()
) ON CONFLICT DO NOTHING;


-- Applications for org1
INSERT INTO applications (organization_id, id, created_at) VALUES (
    'org1',
    'app1',
    NOW()
) ON CONFLICT DO NOTHING;
INSERT INTO application_major_versions (organization_id, application_id, version_number, created_at, updated_at) VALUES (
    'org1',
    'app1',
    1,
    NOW(),
    NOW()
) ON CONFLICT DO NOTHING;
INSERT INTO application_minor_versions (organization_id, application_major_version_id, version_number, review_state, created_at, display_name) VALUES (
    'org1',
    (SELECT id FROM application_major_versions WHERE organization_id = 'org1' AND application_id = 'app1' AND version_number = 1 LIMIT 1),
    1,
    'approved',
    NOW(),
    'Application 1'
) ON CONFLICT DO NOTHING;


-- Deployment requests for org1
DO $$
BEGIN
    IF (SELECT COUNT(*) FROM deployment_requests WHERE organization_id = 'org1' AND application_id = 'app1' LIMIT 1) = 0 THEN
        INSERT INTO deployment_requests (organization_id, application_id, state, created_at, updated_at, finalized_at) VALUES (
            'org1',
            'app1',
            'approved',
            NOW(),
            NOW(),
            NOW()
        );
    END IF;
END $$;


DO $$
DECLARE
    n_apps INT;
    n_major_versions INT;
    n_minor_versions INT;
BEGIN
    n_apps := 12;
    n_major_versions := 60;
    n_minor_versions := 15;

    -- Create applications for org2
    INSERT INTO applications (organization_id, id, created_at)
    SELECT
        'org2' AS organization_id,
        'app' || series AS application_id,
        NOW()
    FROM generate_series(1, n_apps) series
    ON CONFLICT DO NOTHING;

    IF (SELECT COUNT(*) FROM application_major_versions WHERE organization_id = 'org2' AND application_id = 'app1' LIMIT 1) = 0 THEN
        -- For each application, create (n_major_versions - 1) major versions that are finalized
        INSERT INTO application_major_versions
            (organization_id, application_id, version_number, created_at, updated_at)
        SELECT
            'org2' AS organization_id,
            'app' || app_nums AS application_id,
            major_nums AS version_number,
            NOW() AS created_at,
            NOW() AS updated_at
        FROM generate_series(1, n_apps) app_nums,
            generate_series(1, n_major_versions) major_nums;

        -- For each application, create 1 major version that's still draft
        INSERT INTO application_major_versions
            (organization_id, application_id, created_at, updated_at)
        SELECT
            'org2' AS organization_id,
            'app' || app_nums AS application_id,
            NOW() AS created_at,
            NOW() AS updated_at
        FROM generate_series(1, n_apps) app_nums;

        -- For each major version, create (n_minor_versions - 1) minor versions that are not yet approved
        INSERT INTO application_minor_versions
            (organization_id, application_major_version_id, version_number, review_state, created_at, display_name)
        SELECT
            'org2' AS organization_id,
            major_versions.id AS application_major_version_id,
            minor_nums AS version_number,
            'draft' AS review_state,
            NOW() AS created_at,
            'Draft ' || minor_nums AS display_name
        FROM generate_series(1, n_apps) app_nums,
            generate_series(1, n_major_versions) major_nums,
            generate_series(1, n_minor_versions - 1) minor_nums,
            application_major_versions major_versions
        WHERE major_versions.organization_id = 'org2'
        AND major_versions.application_id = 'app' || app_nums
        AND major_versions.version_number = major_nums
        ORDER BY major_nums, minor_nums;

        -- For each major version, create 1 minor version that is approved
        INSERT INTO application_minor_versions
            (organization_id, application_major_version_id, version_number, review_state, created_at, display_name)
        SELECT
            'org2' AS organization_id,
            major_versions.id AS application_major_version_id,
            n_minor_versions AS version_number,
            'approved' AS review_state,
            NOW() AS created_at,
            'Final'
        FROM generate_series(1, n_apps) app_nums,
            application_major_versions major_versions
        WHERE major_versions.organization_id = 'org2'
        AND major_versions.application_id = 'app' || app_nums;
    END IF;
END $$;


COMMIT;
-- For some reason, we need to perform a regular vacuum too, not just a full vacuum,
-- in order to get rid of all invisible tuples.
VACUUM FULL ANALYZE;
VACUUM ANALYZE;
