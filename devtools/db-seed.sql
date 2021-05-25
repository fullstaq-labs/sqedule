BEGIN;


-- Organizations
INSERT INTO organizations (id, display_name) VALUES ('org1', 'My Organization') ON CONFLICT DO NOTHING;
INSERT INTO organizations (id, display_name) VALUES ('org2', 'My Organization 2') ON CONFLICT DO NOTHING;


-- Organization members for org1
INSERT INTO service_accounts (organization_id, name, password_hash, role, created_at, updated_at) VALUES (
    'org1',
    'org_admin_sa',
    '$argon2id$v=19$m=16,t=2,p=1$WlBFUmxyMkJWakw4TUMxVw$NyRkqa3o0uaAHnp7XpjU5A', -- 123456
    'org_admin',
    NOW(),
    NOW()
) ON CONFLICT DO NOTHING;
INSERT INTO service_accounts (organization_id, name, password_hash, role, created_at, updated_at) VALUES (
    'org1',
    'admin_sa',
    '$argon2id$v=19$m=16,t=2,p=1$WlBFUmxyMkJWakw4TUMxVw$NyRkqa3o0uaAHnp7XpjU5A', -- 123456
    'admin',
    NOW(),
    NOW()
) ON CONFLICT DO NOTHING;


-- Applications for org1
INSERT INTO applications (organization_id, id, created_at, updated_at) VALUES (
    'org1',
    'app1',
    NOW(),
    NOW()
) ON CONFLICT DO NOTHING;
INSERT INTO application_versions (organization_id, application_id, version_number, created_at, approved_at) VALUES (
    'org1',
    'app1',
    1,
    NOW(),
    NOW()
) ON CONFLICT DO NOTHING;
INSERT INTO application_adjustments (organization_id, application_version_id, adjustment_number, review_state, created_at, display_name) VALUES (
    'org1',
    (SELECT id FROM application_versions WHERE organization_id = 'org1' AND application_id = 'app1' AND version_number = 1 LIMIT 1),
    1,
    'approved',
    NOW(),
    'Application 1'
) ON CONFLICT DO NOTHING;

INSERT INTO applications (organization_id, id, created_at, updated_at) VALUES (
    'org1',
    'app2',
    NOW(),
    NOW()
) ON CONFLICT DO NOTHING;
INSERT INTO application_versions (organization_id, application_id, version_number, created_at, approved_at) VALUES (
    'org1',
    'app2',
    1,
    NOW(),
    NOW()
) ON CONFLICT DO NOTHING;
INSERT INTO application_adjustments (organization_id, application_version_id, adjustment_number, review_state, created_at, display_name) VALUES (
    'org1',
    (SELECT id FROM application_versions WHERE organization_id = 'org1' AND application_id = 'app2' AND version_number = 1 LIMIT 1),
    1,
    'approved',
    NOW(),
    'Application 2'
) ON CONFLICT DO NOTHING;


-- Approval rulesets (and bindings) for org1
INSERT INTO approval_rulesets (organization_id, id, created_at, updated_at) VALUES(
    'org1',
    'only afternoon',
    NOW(),
    NOW()
) ON CONFLICT DO NOTHING;
INSERT INTO approval_ruleset_versions (organization_id, approval_ruleset_id, version_number, created_at, approved_at) VALUES (
    'org1',
    'only afternoon',
    1,
    NOW(),
    NOW()
) ON CONFLICT DO NOTHING;
INSERT INTO approval_ruleset_adjustments (organization_id, approval_ruleset_version_id, adjustment_number, review_state, created_at, display_name, description, globally_applicable) VALUES (
    'org1',
    (SELECT id FROM approval_ruleset_versions
        WHERE organization_id = 'org1'
        AND approval_ruleset_id = 'only afternoon'
        AND version_number = 1
        LIMIT 1),
    1,
    'approved',
    NOW(),
    'Only afternoon',
    '',
    false
) ON CONFLICT DO NOTHING;
DO $$
DECLARE
    version_id BIGINT;
BEGIN
    SELECT id INTO version_id FROM approval_ruleset_versions
        WHERE organization_id = 'org1'
        AND approval_ruleset_id = 'only afternoon'
        AND version_number = 1
        LIMIT 1;
    IF (SELECT COUNT(*) FROM schedule_approval_rules WHERE organization_id = 'org1'
        AND approval_ruleset_version_id = version_id AND approval_ruleset_adjustment_number = 1 LIMIT 1) = 0
    THEN
        INSERT INTO schedule_approval_rules (organization_id, approval_ruleset_version_id, approval_ruleset_adjustment_number, created_at, begin_time, end_time) VALUES (
            'org1',
            version_id,
            1,
            NOW(),
            '12:00',
            '14:00'
        );
    END IF;
END $$;

INSERT INTO approval_rulesets (organization_id, id, created_at, updated_at) VALUES(
    'org1',
    'only evening',
    NOW(),
    NOW()
) ON CONFLICT DO NOTHING;
INSERT INTO approval_ruleset_versions (organization_id, approval_ruleset_id, version_number, created_at, approved_at) VALUES (
    'org1',
    'only evening',
    1,
    NOW(),
    NOW()
) ON CONFLICT DO NOTHING;
INSERT INTO approval_ruleset_adjustments (organization_id, approval_ruleset_version_id, adjustment_number, review_state, created_at, display_name, description, globally_applicable) VALUES (
    'org1',
    (SELECT id FROM approval_ruleset_versions
        WHERE organization_id = 'org1'
        AND approval_ruleset_id = 'only evening'
        AND version_number = 1
        LIMIT 1),
    1,
    'approved',
    NOW(),
    'Only evening',
    '',
    false
) ON CONFLICT DO NOTHING;
DO $$
DECLARE
    version_id BIGINT;
BEGIN
    SELECT id INTO version_id FROM approval_ruleset_versions
        WHERE organization_id = 'org1'
        AND approval_ruleset_id = 'only evening'
        AND version_number = 1
        LIMIT 1;
    IF (SELECT COUNT(*) FROM schedule_approval_rules WHERE organization_id = 'org1'
        AND approval_ruleset_version_id = version_id AND approval_ruleset_adjustment_number = 1 LIMIT 1) = 0
    THEN
        INSERT INTO schedule_approval_rules (organization_id, approval_ruleset_version_id, approval_ruleset_adjustment_number, created_at, begin_time, end_time) VALUES (
            'org1',
            version_id,
            1,
            NOW(),
            '18:00',
            '23:59'
        );
    END IF;
END $$;

INSERT INTO application_approval_ruleset_bindings (organization_id, application_id, approval_ruleset_id, created_at, updated_at) VALUES (
    'org1',
    'app1',
    'only afternoon',
    NOW(),
    NOW()
) ON CONFLICT DO NOTHING;
INSERT INTO application_approval_ruleset_binding_versions (organization_id, application_id, approval_ruleset_id, version_number, created_at, approved_at) VALUES (
    'org1',
    'app1',
    'only afternoon',
    1,
    NOW(),
    NOW()
) ON CONFLICT DO NOTHING;
INSERT INTO application_approval_ruleset_binding_adjustments (organization_id, application_approval_ruleset_binding_version_id, adjustment_number, review_state, created_at, mode) VALUES (
    'org1',
    (SELECT id FROM application_approval_ruleset_binding_versions
        WHERE organization_id = 'org1' AND application_id = 'app1'
        AND approval_ruleset_id = 'only afternoon' AND version_number = 1
        LIMIT 1),
    1,
    'approved',
    NOW(),
    'enforcing'
) ON CONFLICT DO NOTHING;

INSERT INTO application_approval_ruleset_bindings (organization_id, application_id, approval_ruleset_id, created_at, updated_at) VALUES (
    'org1',
    'app2',
    'only afternoon',
    NOW(),
    NOW()
) ON CONFLICT DO NOTHING;
INSERT INTO application_approval_ruleset_binding_versions (organization_id, application_id, approval_ruleset_id, version_number, created_at, approved_at) VALUES (
    'org1',
    'app2',
    'only afternoon',
    1,
    NOW(),
    NOW()
) ON CONFLICT DO NOTHING;
INSERT INTO application_approval_ruleset_binding_adjustments (organization_id, application_approval_ruleset_binding_version_id, adjustment_number, review_state, created_at, mode) VALUES (
    'org1',
    (SELECT id FROM application_approval_ruleset_binding_versions
        WHERE organization_id = 'org1' AND application_id = 'app2'
        AND approval_ruleset_id = 'only afternoon' AND version_number = 1
        LIMIT 1),
    1,
    'approved',
    NOW(),
    'enforcing'
) ON CONFLICT DO NOTHING;

INSERT INTO application_approval_ruleset_bindings (organization_id, application_id, approval_ruleset_id, created_at, updated_at) VALUES (
    'org1',
    'app2',
    'only evening',
    NOW(),
    NOW()
) ON CONFLICT DO NOTHING;
INSERT INTO application_approval_ruleset_binding_versions (organization_id, application_id, approval_ruleset_id, version_number, created_at, approved_at) VALUES (
    'org1',
    'app2',
    'only evening',
    1,
    NOW(),
    NOW()
) ON CONFLICT DO NOTHING;
INSERT INTO application_approval_ruleset_binding_adjustments (organization_id, application_approval_ruleset_binding_version_id, adjustment_number, review_state, created_at, mode) VALUES (
    'org1',
    (SELECT id FROM application_approval_ruleset_binding_versions
        WHERE organization_id = 'org1' AND application_id = 'app2'
        AND approval_ruleset_id = 'only evening' AND version_number = 1
        LIMIT 1),
    1,
    'approved',
    NOW(),
    'enforcing'
) ON CONFLICT DO NOTHING;


-- Releases for org1
DO $$
DECLARE
    n_releases INT;
    n_releases_finished INT;
BEGIN
    n_releases := 120;
    n_releases_finished := 118;

    IF (SELECT COUNT(*) FROM releases WHERE organization_id = 'org1' AND application_id = 'app1' LIMIT 1) = 0 THEN
        -- Create n_releases_finished releases that are finished
        INSERT INTO releases (organization_id, application_id, state, created_at, updated_at, finalized_at)
        SELECT
            'org1' AS organization_id,
            'app1' AS application_id,
            'approved' AS state,
            NOW() - (INTERVAL '1 day' * series) AS created_at,
            NOW() - (INTERVAL '1 day' * series) AS updated_at,
            NOW() - (INTERVAL '1 day' * series) AS finalized_at
        FROM generate_series(1, n_releases_finished) series;

        INSERT INTO release_created_events (organization_id, release_id, application_id, created_at)
        SELECT
            'org1' AS organization_id,
            (SELECT id FROM releases OFFSET series - 1 LIMIT 1) AS release_id,
            'app1' AS application_id,
            NOW() - (INTERVAL '1 day' * series) AS created_at
        FROM generate_series(1, n_releases_finished) series;

        INSERT INTO release_rule_processed_events (organization_id, release_id, application_id, created_at, result_state, ignored_error)
        SELECT
            'org1' AS organization_id,
            (SELECT id FROM releases OFFSET series - 1 LIMIT 1) AS release_id,
            'app1' AS application_id,
            NOW() - (INTERVAL '1 day' * series) AS created_at,
            'approved' AS result_state,
            true AS ignored_error
        FROM generate_series(1, n_releases_finished) series;


        -- Create 2 releases that are in progress
        INSERT INTO releases (organization_id, application_id, state, created_at, updated_at) VALUES (
            'org1',
            'app1',
            'in_progress',
            (current_date || ' 13:00')::timestamp with time zone,
            NOW()
        );
        INSERT INTO release_created_events (organization_id, release_id, application_id, created_at) VALUES (
            'org1',
            (SELECT currval('releases_id_seq')),
            'app1',
            (current_date || ' 13:00')::timestamp with time zone
        );
        INSERT INTO release_approval_ruleset_bindings (organization_id, application_id, release_id, approval_ruleset_id, approval_ruleset_version_id, approval_ruleset_adjustment_number, mode) VALUES (
            'org1',
            'app1',
            (SELECT currval('releases_id_seq')),
            'only afternoon',
            (SELECT id FROM approval_ruleset_versions
                WHERE organization_id = 'org1'
                AND approval_ruleset_id = 'only afternoon'
                AND version_number = 1
                LIMIT 1),
            1,
            'enforcing'
        );

        INSERT INTO releases (organization_id, application_id, state, created_at, updated_at) VALUES (
            'org1',
            'app1',
            'in_progress',
            (current_date || ' 18:00')::timestamp with time zone,
            NOW()
        );
        INSERT INTO release_created_events (organization_id, release_id, application_id, created_at) VALUES (
            'org1',
            (SELECT currval('releases_id_seq')),
            'app1',
            (current_date || ' 18:00')::timestamp with time zone
        );
        INSERT INTO release_approval_ruleset_bindings (organization_id, application_id, release_id, approval_ruleset_id, approval_ruleset_version_id, approval_ruleset_adjustment_number, mode) VALUES (
            'org1',
            'app1',
            (SELECT currval('releases_id_seq')),
            'only afternoon',
            (SELECT id FROM approval_ruleset_versions
                WHERE organization_id = 'org1'
                AND approval_ruleset_id = 'only afternoon'
                AND version_number = 1
                LIMIT 1),
            1,
            'enforcing'
        );
    END IF;
END $$;


DO $$
DECLARE
    n_apps INT;
    n_versions INT;
    n_adjustments INT;
BEGIN
    n_apps := 12;
    n_versions := 60;
    n_adjustments := 15;

    -- Create applications for org2
    INSERT INTO applications (organization_id, id, created_at, updated_at)
    SELECT
        'org2' AS organization_id,
        'app' || series AS application_id,
        NOW(),
        NOW()
    FROM generate_series(1, n_apps) series
    ON CONFLICT DO NOTHING;

    IF (SELECT COUNT(*) FROM application_versions WHERE organization_id = 'org2' AND application_id = 'app1' LIMIT 1) = 0 THEN
        -- For each application, create (n_versions - 1) versions that are finalized
        INSERT INTO application_versions
            (organization_id, application_id, version_number, created_at, approved_at)
        SELECT
            'org2' AS organization_id,
            'app' || app_nums AS application_id,
            version_nums AS version_number,
            NOW() AS created_at,
            NOW() AS approved_at
        FROM generate_series(1, n_apps) app_nums,
            generate_series(1, n_versions) version_nums;

        -- For each application, create 1 proposal version
        INSERT INTO application_versions
            (organization_id, application_id, created_at)
        SELECT
            'org2' AS organization_id,
            'app' || app_nums AS application_id,
            NOW() AS created_at
        FROM generate_series(1, n_apps) app_nums;

        -- For each version, create (n_adjustments - 1) adjustments that are not yet approved
        INSERT INTO application_adjustments
            (organization_id, application_version_id, adjustment_number, review_state, created_at, display_name)
        SELECT
            'org2' AS organization_id,
            versions.id AS application_version_id,
            adjustment_nums AS adjustment_number,
            'draft' AS review_state,
            NOW() AS created_at,
            'Draft ' || adjustment_nums AS display_name
        FROM generate_series(1, n_apps) app_nums,
            generate_series(1, n_versions) version_nums,
            generate_series(1, n_adjustments - 1) adjustment_nums,
            application_versions versions
        WHERE versions.organization_id = 'org2'
        AND versions.application_id = 'app' || app_nums
        AND versions.version_number = version_nums
        ORDER BY version_nums, adjustment_nums;

        -- For each version, create 1 adjustment that is approved
        INSERT INTO application_adjustments
            (organization_id, application_version_id, adjustment_number, review_state, created_at, display_name)
        SELECT
            'org2' AS organization_id,
            versions.id AS application_version_id,
            n_adjustments AS adjustment_number,
            'approved' AS review_state,
            NOW() AS created_at,
            'Final'
        FROM generate_series(1, n_apps) app_nums,
            application_versions versions
        WHERE versions.organization_id = 'org2'
        AND versions.application_id = 'app' || app_nums;
    END IF;
END $$;


COMMIT;
-- For some reason, we need to perform a regular vacuum too, not just a full vacuum,
-- in order to get rid of all invisible tuples.
VACUUM FULL ANALYZE;
VACUUM ANALYZE;
