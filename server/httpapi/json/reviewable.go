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
	ID            uint64     `json:"id"`
	VersionState  string     `json:"version_state"`
	VersionNumber *uint32    `json:"version_number"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	ApprovedAt    *time.Time `json:"approved_at"`
}

type ReviewableVersionInputBase struct {
	ProposalState proposalstate.ProposalState `json:"proposal_state"`
	Comments      *string                     `json:"comments"`
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
	var state string
	if versionBase.VersionNumber == nil {
		state = "proposal"
	} else {
		state = "approved"
	}
	return ReviewableVersionBase{
		ID:            versionBase.ID,
		VersionState:  state,
		VersionNumber: versionBase.VersionNumber,
		CreatedAt:     versionBase.CreatedAt,
		UpdatedAt:     latestAdjustmentBase.CreatedAt,
		ApprovedAt:    getSqlTimeContentsOrNil(versionBase.ApprovedAt),
	}
}
