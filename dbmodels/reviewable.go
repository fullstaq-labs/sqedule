package dbmodels

import (
	"reflect"

	"gorm.io/gorm"
)

// IReviewable ...
type IReviewable interface {
	GetID() interface{}
	SetLatestMajorVersion(majorVersion IReviewableMajorVersion)
	SetLatestMinorVersion(minorVersion IReviewableMinorVersion)
}

// IReviewableMajorVersion ...
type IReviewableMajorVersion interface {
	GetID() interface{}
	GetReviewableID() interface{}
	AssociateWithReviewable(reviewable IReviewable)
}

// IReviewableMinorVersion ...
type IReviewableMinorVersion interface {
	GetMajorVersionID() interface{}
	AssociateWithMajorVersion(majorVersion IReviewableMajorVersion)
}

// LoadReviewablesLatestVersions loads the latest MajorVersion and MinorVersion records associated
// with the given IReviewable records.
//
// For each found MajorVersion and MinorVersion record, this function calls `SetLatestMajorVersion()`
// and `SetLatestMinorVersion()` on the appropriate IReviewable records.
//
// Parameters:
//
//  - `idColumnType` is the type of the IReviewable's `ID` field.
//  - `idColumnNameInMajorVersionTable` is the foreign key column, in the MajorVersion's table, that refers to the IReviewable's `ID` field.
//  - `majorVersionType` is the type of the MajorVersion struct. It must implement IReviewableMajorVersion.
//  - `majorVersionIDColumnType` is the type of the MajorVersion's `ID` field.
//  - `majorVersionIDColumnNameInMinorVersionTable` is the foreign key column, in the MinorVersion's table, that refers to the MajorVersion's `ID` field.
//  - `minorVersionType` is the type of the MinorVersion struct. It must implement IReviewableMinorVersion.
//  - `organizationID` is the organization in which you want to query the records.
//  - `reviewables` is the list of IReviewables for which you want to load their latest major and minor version records.
func LoadReviewablesLatestVersions(db *gorm.DB,
	idColumnType reflect.Type,
	idColumnNameInMajorVersionTable string,
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

	// Example reviewableIndex type: make(map[string][]*Application)
	// Example reviewableIds type: make([]string, 0, len(applications))
	reviewableIndex, reviewableIds := buildReviewablesIndexAndIDList(idColumnType, reviewables)

	/****** Load associated major versions ******/

	// Example type: *[]ApplicationMajorVersion
	majorVersions := reflectMakePtr(reflect.MakeSlice(reflect.SliceOf(majorVersionType), 0, 0))
	// Example type: map[uint64]*ApplicationMajorVersion
	majorIndex := reflect.MakeMap(reflect.MapOf(majorVersionIDColumnType, reflect.PtrTo(majorVersionType)))
	// Example type: []uint64
	majorIds := reflect.MakeSlice(reflect.SliceOf(majorVersionIDColumnType), 0, 0)
	tx := db.
		Select("DISTINCT ON (organization_id, "+idColumnNameInMajorVersionTable+") *").
		Where("organization_id = ? AND "+idColumnNameInMajorVersionTable+" IN ? AND version_number IS NOT NULL",
			organizationID, reviewableIds.Interface()).
		Order("organization_id, " + idColumnNameInMajorVersionTable + ", version_number DESC").
		Find(majorVersions.Interface())
	if tx.Error != nil {
		return tx.Error
	}
	nversions = majorVersions.Elem().Len()
	for i := 0; i < nversions; i++ {
		majorVersion := reflectMakePtr(majorVersions.Elem().Index(i)).Interface().(IReviewableMajorVersion)
		majorVersionID := reflect.ValueOf(majorVersion.GetID())
		reviewableID := majorVersion.GetReviewableID()
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

	// Example type: *[]ApplicationMinorVersion
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
		minorVersion := reflectMakePtr(minorVersions.Elem().Index(i)).Interface().(IReviewableMinorVersion)

		majorVersion := majorIndex.MapIndex(reflect.ValueOf(minorVersion.GetMajorVersionID())).Interface().(IReviewableMajorVersion)
		minorVersion.AssociateWithMajorVersion(majorVersion)

		// Example type: []*Application
		reviewables := reviewableIndex.MapIndex(reflect.ValueOf(majorVersion.GetReviewableID()))
		for i := 0; i < reviewables.Len(); i++ {
			reviewable := reviewables.Index(i).Interface().(IReviewable)
			reviewable.SetLatestMinorVersion(minorVersion)
		}
	}

	return nil
}

// buildReviewablesIndexAndIdList returns two things:
// 1. An index, that maps each IReviewable ID, to a list of matching IReviewables.
// 2. A list of all unique IReviewable IDs.
//
// For example, given a list of `dbmodels.Application`s, it returns something like
// `(map[string][]*dbmodels.Application, []string)`
func buildReviewablesIndexAndIDList(idColumnType reflect.Type, reviewables []IReviewable) (reflect.Value, reflect.Value) {
	index := reflect.MakeMap(reflect.MapOf(idColumnType, reflect.TypeOf([]IReviewable{})))
	ids := reflect.MakeSlice(reflect.SliceOf(idColumnType), 0, len(reviewables))

	for _, reviewable := range reviewables {
		id := reflect.ValueOf(reviewable.GetID())

		// Type: []IReviewable
		indexElem := index.MapIndex(id)
		if !indexElem.IsValid() {
			indexElem = reflect.ValueOf(make([]IReviewable, 0))
			index.SetMapIndex(id, indexElem)
			ids = reflect.Append(ids, id)
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
