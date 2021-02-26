package dbmigrations

import (
	"time"

	"github.com/fullstaq-labs/sqedule/dbutils/gormigrate"
	"gorm.io/gorm"
)

func init() {
	registerDbMigration(&migration20210226000010)
}

var migration20210226000010 = gormigrate.Migration{
	ID: "20210226000010 Release background job",
	Migrate: func(tx *gorm.DB) error {
		type Organization struct {
			ID string `gorm:"type:citext; primaryKey; not null"`
		}

		type BaseModel struct {
			OrganizationID string       `gorm:"type:citext; primaryKey; not null"`
			Organization   Organization `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
		}

		type DeploymentRequest struct {
			BaseModel
			ApplicationID string `gorm:"type:citext; primaryKey; not null"`
			ID            uint64 `gorm:"primaryKey; not null"`
		}

		type ApprovalRuleset struct {
			BaseModel
			ID string `gorm:"type:citext; primaryKey; not null"`
		}

		type ApprovalRulesetMajorVersion struct {
			OrganizationID string       `gorm:"type:citext; primaryKey; not null; index:approval_ruleset_major_version_idx,unique"`
			Organization   Organization `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
			ID             uint64       `gorm:"primaryKey; autoIncrement; not null"`
		}

		type ApprovalRulesetMinorVersion struct {
			BaseModel
			ApprovalRulesetMajorVersionID uint64 `gorm:"primaryKey; not null"`
			VersionNumber                 uint32 `gorm:"primaryKey; not null"`
		}

		type ReleaseBackgroundJob struct {
			BaseModel
			ApplicationID       string            `gorm:"type:citext; primaryKey; not null"`
			DeploymentRequestID uint64            `gorm:"primaryKey; not null"`
			DeploymentRequest   DeploymentRequest `gorm:"foreignKey:OrganizationID,ApplicationID,DeploymentRequestID; references:OrganizationID,ApplicationID,ID; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
			LockID              uint32            `gorm:"type:int; autoIncrement; unique; not null"`
			CreatedAt           time.Time         `gorm:"not null"`
		}

		type ReleaseBackgroundJobApprovalRulesetBinding struct {
			BaseModel

			ApplicationID                     string               `gorm:"type:citext; primaryKey; not null"`
			DeploymentRequestID               uint64               `gorm:"primaryKey; not null"`
			ReleaseBackgroundJob              ReleaseBackgroundJob `gorm:"foreignKey:OrganizationID,ApplicationID,DeploymentRequestID; references:OrganizationID,ApplicationID,DeploymentRequestID; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
			ApprovalRulesetID                 string               `gorm:"type:citext; primaryKey; not null"`
			ApprovalRulesetMajorVersionID     uint64               `gorm:"not null"`
			ApprovalRulesetMinorVersionNumber uint32               `gorm:"not null"`
			Mode                              string               `gorm:"type:approval_ruleset_binding_mode; not null"`
		}

		err := tx.AutoMigrate(&ReleaseBackgroundJob{}, &ReleaseBackgroundJobApprovalRulesetBinding{})
		if err != nil {
			return err
		}

		err = tx.Exec("ALTER TABLE ONLY release_background_job_approval_ruleset_bindings " +
			"ADD CONSTRAINT fk_approval_ruleset_bindings_organization " +
			"FOREIGN KEY (organization_id) " +
			"REFERENCES organizations(id) " +
			"ON DELETE CASCADE ON UPDATE CASCADE").Error
		if err != nil {
			return err
		}

		err = tx.Exec("ALTER TABLE ONLY release_background_job_approval_ruleset_bindings " +
			"ADD CONSTRAINT fk_approval_ruleset_bindings_release_background_job " +
			"FOREIGN KEY (organization_id,application_id,deployment_request_id) " +
			"REFERENCES release_background_jobs(organization_id,application_id,deployment_request_id) " +
			"ON DELETE CASCADE ON UPDATE CASCADE").Error
		if err != nil {
			return err
		}

		err = tx.Exec("ALTER TABLE ONLY release_background_job_approval_ruleset_bindings " +
			"ADD CONSTRAINT fk_approval_ruleset_bindings_approval_ruleset " +
			"FOREIGN KEY (organization_id,approval_ruleset_id) " +
			"REFERENCES approval_rulesets(organization_id,id) " +
			"ON DELETE CASCADE ON UPDATE CASCADE").Error
		if err != nil {
			return err
		}

		err = tx.Exec("ALTER TABLE ONLY release_background_job_approval_ruleset_bindings " +
			"ADD CONSTRAINT fk_approval_ruleset_bindings_approval_ruleset_major_version " +
			"FOREIGN KEY (organization_id,approval_ruleset_major_version_id) " +
			"REFERENCES approval_ruleset_major_versions(organization_id,id) " +
			"ON DELETE CASCADE ON UPDATE CASCADE").Error
		if err != nil {
			return err
		}

		err = tx.Exec("ALTER TABLE ONLY release_background_job_approval_ruleset_bindings " +
			"ADD CONSTRAINT fk_approval_ruleset_bindings_approval_ruleset_minor_version " +
			"FOREIGN KEY (organization_id,approval_ruleset_major_version_id,approval_ruleset_minor_version_number) " +
			"REFERENCES approval_ruleset_minor_versions(organization_id,approval_ruleset_major_version_id,version_number) " +
			"ON DELETE CASCADE ON UPDATE CASCADE").Error
		if err != nil {
			return err
		}

		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Migrator().DropTable("release_background_jobs", "release_background_job_approval_ruleset_bindings")
	},
}
