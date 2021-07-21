package controllers

import (
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/reviewstate"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
)

type ReviewableReviewProposalTestOptions struct {
	HTTPTestCtx            *HTTPTestContext
	GetProposalPath        func() string
	GetApprovedVersionPath func() string
	Setup                  func(reviewState reviewstate.State)

	ResourceTypeNameInResponse string

	GetFirstProposalAndAdjustment  func() (dbmodels.IReviewableVersion, dbmodels.IReviewableAdjustment)
	GetSecondProposalAndAdjustment func() (dbmodels.IReviewableVersion, dbmodels.IReviewableAdjustment)

	PrimaryKeyJSONFieldName     string
	PrimaryKeyInitialValue      interface{}
	VersionedFieldJSONFieldName string
	VersionedFieldInitialValue  interface{}
}

type ReviewableReviewProposalTestContext struct {
	MakeRequest func(approved bool, reviewState string, expectedCode uint) gin.H
}

func IncludeReviewableReviewProposalTest(options ReviewableReviewProposalTestOptions) *ReviewableReviewProposalTestContext {
	var rctx ReviewableReviewProposalTestContext
	var hctx *HTTPTestContext = options.HTTPTestCtx

	rctx.MakeRequest = func(approved bool, reviewState string, expectedCode uint) gin.H {
		var path string
		if approved {
			path = options.GetApprovedVersionPath()
		} else {
			path = options.GetProposalPath()
		}

		req, err := hctx.NewRequestWithAuth("PUT", path, gin.H{"state": reviewState})
		Expect(err).ToNot(HaveOccurred())
		hctx.ServeHTTP(req)
		Expect(hctx.Recorder.Code).To(BeNumerically("==", expectedCode))

		body, err := hctx.BodyJSON()
		Expect(err).ToNot(HaveOccurred())

		return body
	}

	It("outputs non-versioned fields upon success", func() {
		options.Setup(reviewstate.Reviewing)
		body := rctx.MakeRequest(false, "approved", 200)

		Expect(body).To(HaveKeyWithValue(options.PrimaryKeyJSONFieldName, options.PrimaryKeyInitialValue))
	})

	Specify("approving works", func() {
		options.Setup(reviewstate.Reviewing)
		body := rctx.MakeRequest(false, "approved", 200)

		Expect(body).To(HaveKeyWithValue("version", Not(BeNil())))
		versionJSON := body["version"]

		Expect(versionJSON).To(HaveKeyWithValue("id", Not(BeNil())))
		Expect(versionJSON).To(HaveKeyWithValue("version_state", "approved"))
		Expect(versionJSON).To(HaveKeyWithValue("version_number", BeNumerically("==", 2)))
		Expect(versionJSON).To(HaveKeyWithValue("adjustment_state", "approved"))
		Expect(versionJSON).To(HaveKeyWithValue("approved_at", Not(BeNil())))
		Expect(versionJSON).To(HaveKeyWithValue(options.VersionedFieldJSONFieldName, options.VersionedFieldInitialValue))

		version, adjustment := options.GetFirstProposalAndAdjustment()
		Expect(*version.GetVersionNumber()).To(BeNumerically("==", 2))
		Expect(adjustment.GetReviewState()).To(Equal(reviewstate.Approved))
	})

	Specify("rejecting works", func() {
		options.Setup(reviewstate.Reviewing)
		body := rctx.MakeRequest(false, "rejected", 200)

		Expect(body).To(HaveKeyWithValue("version", Not(BeNil())))
		versionJSON := body["version"]

		Expect(versionJSON).To(HaveKeyWithValue("id", Not(BeNil())))
		Expect(versionJSON).To(HaveKeyWithValue("version_state", "proposal"))
		Expect(versionJSON).To(HaveKeyWithValue("version_number", BeNil()))
		Expect(versionJSON).To(HaveKeyWithValue("adjustment_state", "rejected"))
		Expect(versionJSON).To(HaveKeyWithValue("approved_at", BeNil()))
		Expect(versionJSON).To(HaveKeyWithValue(options.VersionedFieldJSONFieldName, options.VersionedFieldInitialValue))

		version, _ := options.GetFirstProposalAndAdjustment()
		Expect(version.GetVersionNumber()).To(BeNil())
	})

	It("requires an input state", func() {
		options.Setup(reviewstate.Reviewing)
		body := rctx.MakeRequest(false, "", 400)
		Expect(body).To(HaveKeyWithValue("error", ContainSubstring("Invalid review state input")))
	})

	It("rejects input state values other than 'approved' or 'rejected'", func() {
		options.Setup(reviewstate.Reviewing)
		body := rctx.MakeRequest(false, "foo", 400)
		Expect(body).To(HaveKeyWithValue("error", ContainSubstring("Invalid review state input")))
	})

	It("only works if the proposal is in the reviewing state", func() {
		options.Setup(reviewstate.Draft)
		body := rctx.MakeRequest(false, "approved", 422)
		Expect(body).To(HaveKeyWithValue("error", ContainSubstring("This proposal is not awaiting review")))
	})

	It("does not update approved versions", func() {
		options.Setup(reviewstate.Reviewing)
		body := rctx.MakeRequest(true, "approved", 404)
		Expect(body).To(HaveKeyWithValue("error", options.ResourceTypeNameInResponse+" not found"))
	})

	It("creates a CreationAuditRecord", func() {
		var count int64

		tx := hctx.Db.Model(&dbmodels.CreationAuditRecord{}).Count(&count)
		Expect(tx.Error).ToNot(HaveOccurred())
		Expect(count).To(BeNumerically("==", 0))

		options.Setup(reviewstate.Reviewing)
		rctx.MakeRequest(false, "approved", 200)

		tx = hctx.Db.Model(&dbmodels.CreationAuditRecord{}).Count(&count)
		Expect(tx.Error).ToNot(HaveOccurred())
		Expect(count).To(BeNumerically(">=", 1))
	})

	Specify("if the proposal is approved, then it puts all other proposals that are in the reviewing state, into the draft state", func() {
		options.Setup(reviewstate.Reviewing)

		version, adjustment := options.GetSecondProposalAndAdjustment()
		Expect(version.GetVersionNumber()).To(BeNil())
		Expect(adjustment.GetReviewState()).To(Equal(reviewstate.Reviewing))

		rctx.MakeRequest(false, "approved", 200)

		version, adjustment = options.GetSecondProposalAndAdjustment()
		Expect(version.GetVersionNumber()).To(BeNil())
		Expect(adjustment.GetReviewState()).To(Equal(reviewstate.Draft))
	})

	return &rctx
}
