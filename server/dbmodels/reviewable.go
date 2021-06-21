package dbmodels

import (
	"database/sql"
	"reflect"
	"strings"
	"time"

	"github.com/fullstaq-labs/sqedule/lib"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/reviewstate"
	"github.com/fullstaq-labs/sqedule/server/httpapi/json/proposalstate"
	"gorm.io/gorm"
)

//
// ******** Types, constants & variables ********
//

type IReviewable interface {
	GetPrimaryKey() interface{}
	GetPrimaryKeyGormValue() []interface{}
	AssociateWithVersion(version IReviewableVersion)
}

type IReviewableVersion interface {
	GetID() interface{}
	GetReviewablePrimaryKey() interface{}
	GetVersionNumber() *uint32
	AssociateWithReviewable(reviewable IReviewable)
	AssociateWithAdjustment(adjustment IReviewableAdjustment)
}

type IReviewableAdjustment interface {
	GetReviewState() reviewstate.State
	GetVersionID() interface{}
	AssociateWithVersion(version IReviewableVersion)
}

type ReviewableBase struct {
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

type ReviewableVersionBase struct {
	ID            uint64       `gorm:"primaryKey; autoIncrement; not null"`
	VersionNumber *uint32      `gorm:"type:int; check:(version_number > 0)"`
	CreatedAt     time.Time    `gorm:"not null"`
	ApprovedAt    sql.NullTime `gorm:"check:((approved_at IS NULL) = (version_number IS NULL))"`
}

type ReviewableAdjustmentBase struct {
	AdjustmentNumber uint32            `gorm:"type:int; primaryKey; not null; check:(adjustment_number > 0)"`
	ReviewState      reviewstate.State `gorm:"type:review_state; not null"`
	ReviewComments   sql.NullString
	CreatedAt        time.Time `gorm:"not null"`
}

//
// ******** ReviewableVersionBase methods ********
//

func (version ReviewableVersionBase) GetVersionNumber() *uint32 {
	return version.VersionNumber
}

//
// ******** ReviewableAdjustmentBase methods ********
//

func (adjustment ReviewableAdjustmentBase) GetReviewState() reviewstate.State {
	return adjustment.ReviewState
}

//
// ******** Find/load functions ********
//

// LoadReviewablesLatestVersions loads all Versions associated with the given Reviewables.
// For each found Version, it calls `reviewable.AssociateWithVersion(version)` on an
// appropriate Reviewable object.
//
// Parameters:
//
//  - `versionType` is the concrete IReviewableVersion type.
//  - `primaryKeyColumnNamesInVersionTable` are the (possibly composite) foreign key columns, in the Version's table, that refer to the IReviewable's primary key (excluding OrganizationID).
func LoadReviewablesLatestVersions(db *gorm.DB,
	organizationID string,
	reviewables []IReviewable,
	versionType reflect.Type,
	primaryKeyColumnNamesInVersionTable []string) error {

	if len(reviewables) == 0 {
		return nil
	}

	// All the Reviewable objects' primary keys as Gorm values
	var reviewablePrimaryKeysGormValues [][]interface{} = CollectReviewablePrimaryKeys(reviewables)
	// Indexes each Reviewable object by its primary key
	var reviewableIndex map[interface{}][]IReviewable = indexReviewablesByPrimaryKey(reviewables)
	// Type: *[]versionType
	var versions = lib.ReflectMakeValPtr(reflect.MakeSlice(reflect.SliceOf(versionType), 0, 0))
	var primaryKeyColumnNamesInVersionTableAsCommaString = strings.Join(primaryKeyColumnNamesInVersionTable, ",")

	tx := db.
		// DISTINCT ON only works on PostgreSQL. When we want to support other databases, have a look at this alternative:
		// https://stackoverflow.com/a/3800572/20816
		Select("DISTINCT ON (organization_id, "+primaryKeyColumnNamesInVersionTableAsCommaString+") *").
		Where("organization_id = ? AND ("+primaryKeyColumnNamesInVersionTableAsCommaString+") IN (?) AND version_number IS NOT NULL",
			organizationID, reviewablePrimaryKeysGormValues).
		Order("organization_id, " + primaryKeyColumnNamesInVersionTableAsCommaString + ", version_number DESC").
		Find(versions.Interface())
	if tx.Error != nil {
		return tx.Error
	}

	nversions := versions.Elem().Len()
	for i := 0; i < nversions; i++ {
		version := lib.ReflectMakeValPtr(versions.Elem().Index(i)).Interface().(IReviewableVersion)
		reviewablePrimaryKey := version.GetReviewablePrimaryKey()
		matchingReviewables := reviewableIndex[reviewablePrimaryKey]

		version.AssociateWithReviewable(matchingReviewables[0])
		for _, reviewable := range matchingReviewables {
			reviewable.AssociateWithVersion(version)
		}
	}

	return nil
}

func indexReviewablesByPrimaryKey(reviewables []IReviewable) map[interface{}][]IReviewable {
	result := make(map[interface{}][]IReviewable, len(reviewables))
	for _, reviewable := range reviewables {
		primaryKey := reviewable.GetPrimaryKey()
		list, ok := result[primaryKey]
		if !ok {
			list = make([]IReviewable, 0, 1)
		}
		result[primaryKey] = append(list, reviewable)
	}
	return result
}

// LoadReviewableVersionsLatestAdjustments loads all Adjustments associated with the given Versions.
// For each found Adjustment, it calls `version.AssociateWithAdjustment(adjustment)` on an
// appropriate Version object.
//
// Parameters:
//
//  - `adjustmentType` is the concrete IReviewableAdjustment type.
//  - `versionIDColumnNameInAdjustmentTable` is the foreign key column, in the Adjustment's table, that refers to the Version's `ID` field.
func LoadReviewableVersionsLatestAdjustments(db *gorm.DB,
	organizationID string,
	versions []IReviewableVersion,
	adjustmentType reflect.Type,
	versionIDColumnNameInAdjustmentTable string) error {

	if len(versions) == 0 {
		return nil
	}

	// All the Version objects' IDs
	var versionIDs []interface{} = CollectReviewableVersionIDs(versions)
	// Indexes each Version object by its ID
	var versionIndex map[interface{}][]IReviewableVersion = indexReviewableVersionsByID(versions)

	// Type: *[]actualAdjustmentType
	adjustments := lib.ReflectMakeValPtr(reflect.MakeSlice(reflect.SliceOf(adjustmentType), 0, 0))
	tx := db.
		// DISTINCT ON only works on PostgreSQL. When we want to support other databases, have a look at this alternative:
		// https://stackoverflow.com/a/3800572/20816
		Select("DISTINCT ON (organization_id, "+versionIDColumnNameInAdjustmentTable+") *").
		Where("organization_id = ? AND "+versionIDColumnNameInAdjustmentTable+" IN ?",
			organizationID, versionIDs).
		Order("organization_id, " + versionIDColumnNameInAdjustmentTable + ", adjustment_number DESC").
		Find(adjustments.Interface())
	if tx.Error != nil {
		return tx.Error
	}

	nadjustments := adjustments.Elem().Len()
	for i := 0; i < nadjustments; i++ {
		adjustment := lib.ReflectMakeValPtr(adjustments.Elem().Index(i)).Interface().(IReviewableAdjustment)
		versionID := adjustment.GetVersionID()
		matchingVersions := versionIndex[versionID]
		adjustment.AssociateWithVersion(matchingVersions[0])
		for _, version := range matchingVersions {
			version.AssociateWithAdjustment(adjustment)
		}
	}

	return nil
}

func indexReviewableVersionsByID(versions []IReviewableVersion) map[interface{}][]IReviewableVersion {
	result := make(map[interface{}][]IReviewableVersion, len(versions))
	for _, version := range versions {
		id := version.GetID()
		list, ok := result[id]
		if !ok {
			list = make([]IReviewableVersion, 0, 1)
		}
		result[id] = append(list, version)
	}
	return result
}

//
// ******** Other functions ********
//

// CollectReviewablePrimaryKeys turns a `[]IReviewable` into a list of their primary keys.
// This is done by calling `reviewable.GetPrimaryKey()` for each element.
func CollectReviewablePrimaryKeys(reviewables []IReviewable) [][]interface{} {
	result := make([][]interface{}, 0, len(reviewables))
	for _, reviewable := range reviewables {
		primaryKey := reviewable.GetPrimaryKeyGormValue()
		result = append(result, primaryKey)
	}
	return result
}

// CollectReviewableVersionIDs turns a `[]IReviewableVersion` into a list of their IDs.
// This is done by calling `version.GetID()` for each element.
func CollectReviewableVersionIDs(versions []IReviewableVersion) []interface{} {
	result := make([]interface{}, 0, len(versions))
	for _, version := range versions {
		result = append(result, version.GetID())
	}
	return result
}

// FinalizeReviewableProposal transitions a Reviewable's proposal to either in the 'reviewing' state or the 'approved' state,
// depending on whether a review of this proposal is required. It is to be used inside PATCH API routes.
func FinalizeReviewableProposal(version *ReviewableVersionBase, adjustment *ReviewableAdjustmentBase, latestVersionNumber uint32, requiresReview bool) {
	// TODO: add support for comments
	if requiresReview {
		markProposalAsReviewing(adjustment)
	} else {
		markProposalAsApproved(version, latestVersionNumber+1, adjustment)
	}
}

func markProposalAsReviewing(adjustment *ReviewableAdjustmentBase) {
	adjustment.ReviewState = reviewstate.Reviewing
}

func markProposalAsApproved(version *ReviewableVersionBase, versionNumber uint32, adjustment *ReviewableAdjustmentBase) {
	adjustment.ReviewState = reviewstate.Approved
	version.VersionNumber = &versionNumber
	version.ApprovedAt = sql.NullTime{Time: time.Now(), Valid: true}
}

// SetReviewableAdjustmentReviewStateFromUnfinalizedProposalState sets an Adjustment's ReviewState to an appropriate value,
// based on a ProposalState given by a client. It is to be used inside PATCH API routes.
//
// Precondition: `state` is not `Final`.
func SetReviewableAdjustmentReviewStateFromUnfinalizedProposalState(adjustment *ReviewableAdjustmentBase, state proposalstate.State) {
	switch state {
	case proposalstate.Unset, proposalstate.Draft:
		adjustment.ReviewState = reviewstate.Draft
	case proposalstate.Final:
		panic("Not allowed to call this function on a proposalstate of 'Final'")
	case proposalstate.Abandon:
		adjustment.ReviewState = reviewstate.Abandoned
	default:
		panic("Unsupported proposal state '" + string(state) + "'")
	}
}
