package json

import (
	"time"

	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/fullstaq-labs/sqedule/server/httpapi/json/proposalstate"
)

//
// ******** Types, constants & variables ********
//

type ReviewableBase struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ReviewableVersionBase struct {
	ID              uint64     `json:"id"`
	VersionState    string     `json:"version_state"`
	VersionNumber   *uint32    `json:"version_number"`
	AdjustmentState string     `json:"adjustment_state"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	ApprovedAt      *time.Time `json:"approved_at"`
}

type ReviewableVersionInputBase struct {
	ProposalState proposalstate.State `json:"proposal_state"`
	Comments      *string             `json:"comments"`
}

//
// ******** Constructor functions ********
//

func createReviewableBase(base dbmodels.ReviewableBase) ReviewableBase {
	return ReviewableBase{
		CreatedAt: base.CreatedAt,
		UpdatedAt: base.UpdatedAt,
	}
}

func createReviewableVersionBase(versionBase dbmodels.ReviewableVersionBase, latestAdjustmentBase dbmodels.ReviewableAdjustmentBase) ReviewableVersionBase {
	var versionState string
	if versionBase.VersionNumber == nil {
		versionState = "proposal"
	} else {
		versionState = "approved"
	}
	return ReviewableVersionBase{
		ID:              versionBase.ID,
		VersionState:    versionState,
		VersionNumber:   versionBase.VersionNumber,
		AdjustmentState: string(latestAdjustmentBase.ReviewState),
		CreatedAt:       versionBase.CreatedAt,
		UpdatedAt:       latestAdjustmentBase.CreatedAt,
		ApprovedAt:      getSqlTimeContentsOrNil(versionBase.ApprovedAt),
	}
}
