package dbmodels

import (
	"database/sql"
	"reflect"
	"strings"
	"time"

	"github.com/fullstaq-labs/sqedule/server/dbmodels/reviewstate"
	"gorm.io/gorm"
)

// IReviewable ...
type IReviewable interface {
	GetPrimaryKey() interface{}
	SetLatestMajorVersion(majorVersion IReviewableMajorVersion)
	SetLatestMinorVersion(minorVersion IReviewableMinorVersion)
}

// IReviewableMajorVersion ...
type IReviewableMajorVersion interface {
	GetID() interface{}
	GetReviewablePrimaryKey() interface{}
	AssociateWithReviewable(reviewable IReviewable)
}

// IReviewableMinorVersion ...
type IReviewableMinorVersion interface {
	GetMajorVersionID() interface{}
	AssociateWithMajorVersion(majorVersion IReviewableMajorVersion)
}

// IReviewableCompositeKey ...
type IReviewableCompositeKey interface {
	GormValue() interface{}
}

type ReviewableBase struct {
	CreatedAt time.Time `gorm:"not null"`
}

type ReviewableVersionBase struct {
	ID            uint64    `gorm:"primaryKey; autoIncrement; not null"`
	VersionNumber *uint32   `gorm:"type:int; check:(version_number > 0)"`
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`
}

type ReviewableAdjustmentBase struct {
	VersionNumber  uint32            `gorm:"type:int; primaryKey; not null; check:(version_number > 0)"`
	ReviewState    reviewstate.State `gorm:"type:review_state; not null"`
	ReviewComments sql.NullString
	CreatedAt      time.Time `gorm:"not null"`
}

// LoadReviewablesLatestVersions loads the latest MajorVersion and MinorVersion records associated
// with the given IReviewable records.
//
// For each found MajorVersion and MinorVersion record, this function calls `SetLatestMajorVersion()`
// and `SetLatestMinorVersion()` on the appropriate IReviewable records.
//
// Parameters:
//
//  - `primaryKeyType` is the type of the IReviewable's (possibly composite) primary key (excluding OrganizationID).
//    When the primary key is singular, this is the type of the `ID` field.
//    When the primary key is composite, then this should be a struct that contains the two keys. Furthermore, this struct must implement `IReviewableCompositeKey`.
//  - `primaryKeyColumnNamesInMajorVersionTable` are the (possibly composite) foreign key columns, in the MajorVersion's table, that refer to the IReviewable's primary key (excluding OrganizationID).
//  - `primaryKeyGormValueType` is the type of the IReviewable's (possibly composite) primary key when passed as a value to a GORM `Where()` clause.
//    When the primary key is singular, this should be the same as `primaryKeyType`.
//    When the primary key is composite, this should be the type that's returned by primaryKeyType's `IReviewableCompositeKey.GormValue()` method.
//  - `majorVersionType` is the type of the MajorVersion struct. It must implement IReviewableMajorVersion.
//  - `majorVersionIDColumnType` is the type of the MajorVersion's `ID` field.
//  - `majorVersionIDColumnNameInMinorVersionTable` is the foreign key column, in the MinorVersion's table, that refers to the MajorVersion's `ID` field.
//  - `minorVersionType` is the type of the MinorVersion struct. It must implement IReviewableMinorVersion.
//  - `organizationID` is the organization in which you want to query the records.
//  - `reviewables` is the list of IReviewables for which you want to load their latest major and minor version records.
func LoadReviewablesLatestVersions(db *gorm.DB,
	primaryKeyType reflect.Type,
	primaryKeyColumnNamesInMajorVersionTable []string,
	primaryKeyGormValueType reflect.Type,
	majorVersionType reflect.Type,
	majorVersionIDColumnType reflect.Type,
	majorVersionIDColumnNameInMinorVersionTable string,
	minorVersionType reflect.Type,
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
		len(primaryKeyColumnNamesInMajorVersionTable) > 1, reviewables)

	/****** Load associated major versions ******/

	// Type: *[]actualMajorVersionType
	majorVersions := reflectMakePtr(reflect.MakeSlice(reflect.SliceOf(majorVersionType), 0, 0))
	// Type: map[actualMajorVersionIDColumnType]*actualMajorVersionType
	majorIndex := reflect.MakeMap(reflect.MapOf(majorVersionIDColumnType, reflect.PtrTo(majorVersionType)))
	// Type: []actualMajorVersionIDColumnType
	majorIds := reflect.MakeSlice(reflect.SliceOf(majorVersionIDColumnType), 0, 0)
	primaryKeyColumnNamesInMajorVersionTableAsCommaString := strings.Join(primaryKeyColumnNamesInMajorVersionTable, ",")
	tx := db.
		// DISTINCT ON only works on PostgreSQL. When we want to support other databases, have a look at this alternative:
		// https://stackoverflow.com/a/3800572/20816
		Select("DISTINCT ON (organization_id, "+primaryKeyColumnNamesInMajorVersionTableAsCommaString+") *").
		Where("organization_id = ? AND ("+primaryKeyColumnNamesInMajorVersionTableAsCommaString+") IN (?) AND version_number IS NOT NULL",
			organizationID, reviewableIds.Interface()).
		Order("organization_id, " + primaryKeyColumnNamesInMajorVersionTableAsCommaString + ", version_number DESC").
		Find(majorVersions.Interface())
	if tx.Error != nil {
		return tx.Error
	}
	nversions = majorVersions.Elem().Len()
	for i := 0; i < nversions; i++ {
		majorVersion := reflectMakePtr(majorVersions.Elem().Index(i)).Interface().(IReviewableMajorVersion)
		majorVersionID := reflect.ValueOf(majorVersion.GetID())
		reviewableID := majorVersion.GetReviewablePrimaryKey()
		// Example type: []*Application
		reviewables := reviewableIndex.MapIndex(reflect.ValueOf(reviewableID))

		for i := 0; i < reviewables.Len(); i++ {
			reviewable := reviewables.Index(i).Interface().(IReviewable)
			reviewable.SetLatestMajorVersion(majorVersion)
		}

		firstMatch := reviewables.Index(0).Interface().(IReviewable)
		majorVersion.AssociateWithReviewable(firstMatch)

		majorIndex.SetMapIndex(majorVersionID, reflect.ValueOf(majorVersion))
		majorIds = reflect.Append(majorIds, majorVersionID)
	}

	/****** Load associated Minor versions ******/

	// Type: *[]actualMinorVersionType
	minorVersions := reflectMakePtr(reflect.MakeSlice(reflect.SliceOf(minorVersionType), 0, 0))
	tx = db.
		Select("DISTINCT ON (organization_id, "+majorVersionIDColumnNameInMinorVersionTable+") *").
		Where("organization_id = ? AND "+majorVersionIDColumnNameInMinorVersionTable+" IN ?",
			organizationID, majorIds.Interface()).
		Order("organization_id, " + majorVersionIDColumnNameInMinorVersionTable + ", version_number DESC").
		Find(minorVersions.Interface())
	if tx.Error != nil {
		return tx.Error
	}
	nversions = minorVersions.Elem().Len()
	for i := 0; i < nversions; i++ {
		// Type: actualMinorVersionType
		minorVersion := reflectMakePtr(minorVersions.Elem().Index(i)).Interface().(IReviewableMinorVersion)

		majorVersionID := reflect.ValueOf(minorVersion.GetMajorVersionID())
		// Type: actualMajorVersionType
		majorVersion := majorIndex.MapIndex(majorVersionID).Interface().(IReviewableMajorVersion)
		minorVersion.AssociateWithMajorVersion(majorVersion)

		reviewablePrimaryKey := reflect.ValueOf(majorVersion.GetReviewablePrimaryKey())
		// Example type: []*Application
		reviewables := reviewableIndex.MapIndex(reviewablePrimaryKey)
		for i := 0; i < reviewables.Len(); i++ {
			reviewable := reviewables.Index(i).Interface().(IReviewable)
			reviewable.SetLatestMinorVersion(minorVersion)
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
