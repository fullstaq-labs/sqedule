package controllers

import (
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/proposalstate"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
)

type ReviewableReviewProposalTestOptions struct {
	HTTPTestCtx            *HTTPTestContext
	GetProposalPath        func() string
	GetApprovedVersionPath func() string
	Setup                  func(proposalState proposalstate.State)

	ResourceTypeNameInResponse string

	GetFirstProposalAndAdjustment  func() (dbmodels.IReviewableVersion, dbmodels.IReviewableAdjustment)
	GetSecondProposalAndAdjustment func() (dbmodels.IReviewableVersion, dbmodels.IReviewableAdjustment)

	AssertNonVersionedJSONFieldsExist func(resource map[string]interface{})
	VersionedFieldJSONFieldName       string
	VersionedFieldInitialValue        interface{}
}

type ReviewableReviewProposalTestContext struct {
	MakeRequest func(approved bool, proposalState string, expectedCode uint) gin.H
}

func IncludeReviewableReviewProposalTest(options ReviewableReviewProposalTestOptions) *ReviewableReviewProposalTestContext {
	var rctx ReviewableReviewProposalTestContext
	var hctx *HTTPTestContext = options.HTTPTestCtx

	rctx.MakeRequest = func(approved bool, proposalState string, expectedCode uint) gin.H {
		var path string
		if approved {
			path = options.GetApprovedVersionPath()
		} else {
			path = options.GetProposalPath()
		}

		req, err := hctx.NewRequestWithAuth("PUT", path, gin.H{"state": proposalState})
		Expect(err).ToNot(HaveOccurred())
		hctx.ServeHTTP(req)
		Expect(hctx.Recorder.Code).To(BeNumerically("==", expectedCode))

		body, err := hctx.BodyJSON()
		Expect(err).ToNot(HaveOccurred())

		return body
	}

	It("outputs non-versioned fields upon success", func() {
		options.Setup(proposalstate.Reviewing)
		body := rctx.MakeRequest(false, "approved", 200)

		if options.AssertNonVersionedJSONFieldsExist != nil {
			options.AssertNonVersionedJSONFieldsExist(body)
		}
	})

	Specify("approving works", func() {
		options.Setup(proposalstate.Reviewing)
		body := rctx.MakeRequest(false, "approved", 200)

		Expect(body).To(HaveKeyWithValue("version", Not(BeNil())))
		versionJSON := body["version"]

		Expect(versionJSON).To(HaveKeyWithValue("id", Not(BeNil())))
		Expect(versionJSON).To(HaveKeyWithValue("version_state", "approved"))
		Expect(versionJSON).To(HaveKeyWithValue("version_number", BeNumerically("==", 2)))
		Expect(versionJSON).To(HaveKeyWithValue("proposal_state", "approved"))
		Expect(versionJSON).To(HaveKeyWithValue("approved_at", Not(BeNil())))
		Expect(versionJSON).To(HaveKeyWithValue(options.VersionedFieldJSONFieldName, options.VersionedFieldInitialValue))

		version, adjustment := options.GetFirstProposalAndAdjustment()
		Expect(*version.GetVersionNumber()).To(BeNumerically("==", 2))
		Expect(adjustment.GetProposalState()).To(Equal(proposalstate.Approved))
	})

	Specify("rejecting works", func() {
		options.Setup(proposalstate.Reviewing)
		body := rctx.MakeRequest(false, "rejected", 200)

		Expect(body).To(HaveKeyWithValue("version", Not(BeNil())))
		versionJSON := body["version"]

		Expect(versionJSON).To(HaveKeyWithValue("id", Not(BeNil())))
		Expect(versionJSON).To(HaveKeyWithValue("version_state", "proposal"))
		Expect(versionJSON).To(HaveKeyWithValue("version_number", BeNil()))
		Expect(versionJSON).To(HaveKeyWithValue("proposal_state", "rejected"))
		Expect(versionJSON).To(HaveKeyWithValue("approved_at", BeNil()))
		Expect(versionJSON).To(HaveKeyWithValue(options.VersionedFieldJSONFieldName, options.VersionedFieldInitialValue))

		version, _ := options.GetFirstProposalAndAdjustment()
		Expect(version.GetVersionNumber()).To(BeNil())
	})

	It("requires an input state", func() {
		options.Setup(proposalstate.Reviewing)
		body := rctx.MakeRequest(false, "", 400)
		Expect(body).To(HaveKeyWithValue("error", ContainSubstring("Invalid review state input")))
	})

	It("rejects input state values other than 'approved' or 'rejected'", func() {
		options.Setup(proposalstate.Reviewing)
		body := rctx.MakeRequest(false, "foo", 400)
		Expect(body).To(HaveKeyWithValue("error", ContainSubstring("Invalid review state input")))
	})

	It("only works if the proposal is in the reviewing state", func() {
		options.Setup(proposalstate.Draft)
		body := rctx.MakeRequest(false, "approved", 422)
		Expect(body).To(HaveKeyWithValue("error", ContainSubstring("This proposal is not awaiting review")))
	})

	It("does not update approved versions", func() {
		options.Setup(proposalstate.Reviewing)
		body := rctx.MakeRequest(true, "approved", 404)
		Expect(body).To(HaveKeyWithValue("error", options.ResourceTypeNameInResponse+" not found"))
	})

	It("creates a CreationAuditRecord", func() {
		var count int64

		tx := hctx.Db.Model(&dbmodels.CreationAuditRecord{}).Count(&count)
		Expect(tx.Error).ToNot(HaveOccurred())
		Expect(count).To(BeNumerically("==", 0))

		options.Setup(proposalstate.Reviewing)
		rctx.MakeRequest(false, "approved", 200)

		tx = hctx.Db.Model(&dbmodels.CreationAuditRecord{}).Count(&count)
		Expect(tx.Error).ToNot(HaveOccurred())
		Expect(count).To(BeNumerically(">=", 1))
	})

	Specify("if the proposal is approved, then it puts all other proposals that are in the reviewing state, into the draft state", func() {
		options.Setup(proposalstate.Reviewing)

		version, adjustment := options.GetSecondProposalAndAdjustment()
		Expect(version.GetVersionNumber()).To(BeNil())
		Expect(adjustment.GetProposalState()).To(Equal(proposalstate.Reviewing))

		rctx.MakeRequest(false, "approved", 200)

		version, adjustment = options.GetSecondProposalAndAdjustment()
		Expect(version.GetVersionNumber()).To(BeNil())
		Expect(adjustment.GetProposalState()).To(Equal(proposalstate.Draft))
	})

	return &rctx
}
