package dbmodels

import (
	"database/sql"
	"reflect"
	"strings"
	"time"

	"github.com/fullstaq-labs/sqedule/server/dbmodels/reviewstate"
	"gorm.io/gorm"
)

type IReviewable interface {
	GetPrimaryKey() interface{}
	SetLatestVersion(version IReviewableVersion)
	SetLatestAdjustment(adjustment IReviewableAdjustment)
}

type IReviewableVersion interface {
	GetID() interface{}
	GetReviewablePrimaryKey() interface{}
	AssociateWithReviewable(reviewable IReviewable)
}

type IReviewableAdjustment interface {
	GetVersionID() interface{}
	AssociateWithVersion(version IReviewableVersion)
}

type IReviewableCompositeKey interface {
	GormValue() interface{}
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

// LoadReviewablesLatestVersions loads the latest Version and Adjustment records associated
// with the given IReviewable records.
//
// For each found Version and Adjustment record, this function calls `SetLatestVersion()`
// and `SetLatestAdjustment()` on the appropriate IReviewable records.
//
// Parameters:
//
//  - `primaryKeyType` is the type of the IReviewable's (possibly composite) primary key (excluding OrganizationID).
//    When the primary key is singular, this is the type of the `ID` field.
//    When the primary key is composite, then this should be a struct that contains the two keys. Furthermore, this struct must implement `IReviewableCompositeKey`.
//  - `primaryKeyColumnNamesInVersionTable` are the (possibly composite) foreign key columns, in the Version's table, that refer to the IReviewable's primary key (excluding OrganizationID).
//  - `primaryKeyGormValueType` is the type of the IReviewable's (possibly composite) primary key when passed as a value to a GORM `Where()` clause.
//    When the primary key is singular, this should be the same as `primaryKeyType`.
//    When the primary key is composite, this should be the type that's returned by primaryKeyType's `IReviewableCompositeKey.GormValue()` method.
//  - `versionType` is the type of the Version struct. It must implement IReviewableVersion.
//  - `versionIDColumnType` is the type of the Version's `ID` field.
//  - `versionIDColumnNameInAdjustmentTable` is the foreign key column, in the Adjustment's table, that refers to the Version's `ID` field.
//  - `adjustmentType` is the type of the Adjustment struct. It must implement IReviewableAdjustment.
//  - `organizationID` is the organization in which you want to query the records.
//  - `reviewables` is the list of IReviewables for which you want to load their latest version and adjustment records.
func LoadReviewablesLatestVersions(db *gorm.DB,
	primaryKeyType reflect.Type,
	primaryKeyColumnNamesInVersionTable []string,
	primaryKeyGormValueType reflect.Type,
	versionType reflect.Type,
	versionIDColumnType reflect.Type,
	versionIDColumnNameInAdjustmentTable string,
	adjustmentType reflect.Type,
	organizationID string,
	reviewables []IReviewable) error {

	var nversions int

	if len(reviewables) == 0 {
		return nil
	}

	// reviewableIndex type: map[actualPrimaryKeyType][]*Application
	// reviewableIds type  : []actualPrimaryKeyGormValueType
	// reviewableIds length: len(applications)
	reviewableIndex, reviewableIds := buildReviewablesIndexAndIDList(primaryKeyType, primaryKeyGormValueType,
		len(primaryKeyColumnNamesInVersionTable) > 1, reviewables)

	/****** Load associated versions ******/

	// Type: *[]actualVersionType
	versions := reflectMakePtr(reflect.MakeSlice(reflect.SliceOf(versionType), 0, 0))
	// Type: map[actualVersionIDColumnType]*actualVersionType
	versionIndex := reflect.MakeMap(reflect.MapOf(versionIDColumnType, reflect.PtrTo(versionType)))
	// Type: []actualVersionIDColumnType
	versionIds := reflect.MakeSlice(reflect.SliceOf(versionIDColumnType), 0, 0)
	primaryKeyColumnNamesInVersionTableAsCommaString := strings.Join(primaryKeyColumnNamesInVersionTable, ",")
	tx := db.
		// DISTINCT ON only works on PostgreSQL. When we want to support other databases, have a look at this alternative:
		// https://stackoverflow.com/a/3800572/20816
		Select("DISTINCT ON (organization_id, "+primaryKeyColumnNamesInVersionTableAsCommaString+") *").
		Where("organization_id = ? AND ("+primaryKeyColumnNamesInVersionTableAsCommaString+") IN (?) AND version_number IS NOT NULL",
			organizationID, reviewableIds.Interface()).
		Order("organization_id, " + primaryKeyColumnNamesInVersionTableAsCommaString + ", version_number DESC").
		Find(versions.Interface())
	if tx.Error != nil {
		return tx.Error
	}
	nversions = versions.Elem().Len()
	for i := 0; i < nversions; i++ {
		version := reflectMakePtr(versions.Elem().Index(i)).Interface().(IReviewableVersion)
		versionID := reflect.ValueOf(version.GetID())
		reviewableID := version.GetReviewablePrimaryKey()
		// Example type: []*Application
		reviewables := reviewableIndex.MapIndex(reflect.ValueOf(reviewableID))

		for i := 0; i < reviewables.Len(); i++ {
			reviewable := reviewables.Index(i).Interface().(IReviewable)
			reviewable.SetLatestVersion(version)
		}

		firstMatch := reviewables.Index(0).Interface().(IReviewable)
		version.AssociateWithReviewable(firstMatch)

		versionIndex.SetMapIndex(versionID, reflect.ValueOf(version))
		versionIds = reflect.Append(versionIds, versionID)
	}

	/****** Load associated Adjustments ******/

	// Type: *[]actualAdjustmentType
	adjustments := reflectMakePtr(reflect.MakeSlice(reflect.SliceOf(adjustmentType), 0, 0))
	tx = db.
		Select("DISTINCT ON (organization_id, "+versionIDColumnNameInAdjustmentTable+") *").
		Where("organization_id = ? AND "+versionIDColumnNameInAdjustmentTable+" IN ?",
			organizationID, versionIds.Interface()).
		Order("organization_id, " + versionIDColumnNameInAdjustmentTable + ", adjustment_number DESC").
		Find(adjustments.Interface())
	if tx.Error != nil {
		return tx.Error
	}
	nversions = adjustments.Elem().Len()
	for i := 0; i < nversions; i++ {
		// Type: actualAdjustmentType
		adjustment := reflectMakePtr(adjustments.Elem().Index(i)).Interface().(IReviewableAdjustment)

		versionID := reflect.ValueOf(adjustment.GetVersionID())
		// Type: actualVersionType
		version := versionIndex.MapIndex(versionID).Interface().(IReviewableVersion)
		adjustment.AssociateWithVersion(version)

		reviewablePrimaryKey := reflect.ValueOf(version.GetReviewablePrimaryKey())
		// Example type: []*Application
		reviewables := reviewableIndex.MapIndex(reviewablePrimaryKey)
		for i := 0; i < reviewables.Len(); i++ {
			reviewable := reviewables.Index(i).Interface().(IReviewable)
			reviewable.SetLatestAdjustment(adjustment)
		}
	}

	return nil
}

// buildReviewablesIndexAndIdList returns two things:
// 1. An index, that maps each IReviewable primary key, to a list of matching IReviewables.
// 2. A list of all unique IReviewable primary keys.
//
// For example, given a list of `dbmodels.Application`s, it returns something like
// `(map[string][]*dbmodels.Application, []string)`
func buildReviewablesIndexAndIDList(primaryKeyType reflect.Type, primaryKeyGormValueType reflect.Type, primaryKeyIsComposite bool, reviewables []IReviewable) (reflect.Value, reflect.Value) {
	index := reflect.MakeMap(reflect.MapOf(primaryKeyType, reflect.TypeOf([]IReviewable{})))
	ids := reflect.MakeSlice(reflect.SliceOf(primaryKeyGormValueType), 0, len(reviewables))

	for _, reviewable := range reviewables {
		id := reflect.ValueOf(reviewable.GetPrimaryKey())

		// Type: []IReviewable
		indexElem := index.MapIndex(id)
		if !indexElem.IsValid() {
			indexElem = reflect.ValueOf(make([]IReviewable, 0))
			index.SetMapIndex(id, indexElem)
			if primaryKeyIsComposite {
				compositeKey := id.Interface().(IReviewableCompositeKey)
				ids = reflect.Append(ids, reflect.ValueOf(compositeKey.GormValue()))
			} else {
				ids = reflect.Append(ids, id)
			}
		}

		indexElem = reflect.Append(indexElem, reflect.ValueOf(reviewable))
		index.SetMapIndex(id, indexElem)
	}

	return index, ids
}

func reflectMakePtr(val reflect.Value) reflect.Value {
	ptr := reflect.New(val.Type())
	ptr.Elem().Set(val)
	return ptr
}

// FinalizeReviewableProposal transitions a Reviewable's proposal to either in the 'reviewing' state or the 'approved' state,
// depending on whether a review of this proposal is required.
func FinalizeReviewableProposal(version *ReviewableVersionBase, adjustment *ReviewableAdjustmentBase, latestVersionNumber uint32, requiresReview bool) {
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
