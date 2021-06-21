package controllers

import (
	"reflect"

	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/reviewstate"
	"github.com/fullstaq-labs/sqedule/server/dbutils"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
)

type ReviewableUpdateProposalTestOptions struct {
	HTTPTestCtx *HTTPTestContext
	GetPath     func(approved bool) string
	Setup       func(reviewState reviewstate.State)

	Input gin.H

	AdjustmentType reflect.Type

	GetPrimaryKey           func(resource interface{}) interface{}
	PrimaryKeyJSONFieldName string
	PrimaryKeyInitialValue  interface{}

	GetResourceVersionAndLatestAdjustment func() (dbmodels.IReviewableVersion, dbmodels.IReviewableAdjustment)
	VersionedFieldJSONFieldName           string
	VersionedFieldUpdatedValue            interface{}
}

type ReviewableUpdateProposalTestContext struct {
	MakeRequest func(approved bool, proposalState string, expectedCode int) gin.H
}

func IncludeReviewableUpdateProposalTest(options ReviewableUpdateProposalTestOptions) *ReviewableUpdateProposalTestContext {
	var rctx ReviewableUpdateProposalTestContext
	var hctx *HTTPTestContext = options.HTTPTestCtx

	rctx.MakeRequest = func(approved bool, proposalState string, expectedCode int) gin.H {
		input := gin.H{}
		for k, v := range options.Input {
			input[k] = v
		}
		if len(proposalState) > 0 {
			input["proposal_state"] = proposalState
		}

		req, err := hctx.NewRequestWithAuth("PATCH", options.GetPath(approved), input)
		Expect(err).ToNot(HaveOccurred())
		hctx.ServeHTTP(req)

		Expect(hctx.HttpRecorder.Code).To(Equal(expectedCode))
		body, err := hctx.BodyJSON()
		Expect(err).ToNot(HaveOccurred())

		return body
	}

	It("outputs non-versioned fields", func() {
		options.Setup(reviewstate.Draft)
		body := rctx.MakeRequest(false, "", 200)

		Expect(body).To(HaveKeyWithValue(options.PrimaryKeyJSONFieldName, options.PrimaryKeyInitialValue))
	})

	It("patches the proposal", func() {
		options.Setup(reviewstate.Draft)
		body := rctx.MakeRequest(false, "", 200)

		Expect(body).To(HaveKeyWithValue("version", Not(BeEmpty())))
		version := body["version"].(map[string]interface{})
		Expect(version).To(HaveKeyWithValue(options.VersionedFieldJSONFieldName, options.VersionedFieldUpdatedValue))
	})

	It("does not allow patching approved versions", func() {
		options.Setup(reviewstate.Draft)
		rctx.MakeRequest(true, "", 404)
	})

	It("keeps the proposal as a draft by default", func() {
		options.Setup(reviewstate.Draft)
		body := rctx.MakeRequest(false, "", 200)

		Expect(body).To(HaveKeyWithValue("version", Not(BeEmpty())))
		version := body["version"].(map[string]interface{})
		Expect(version).To(HaveKeyWithValue("version_state", "proposal"))
		Expect(version).To(HaveKeyWithValue("version_number", BeNil()))
		Expect(version).To(HaveKeyWithValue("adjustment_state", "draft"))
		Expect(version).To(HaveKeyWithValue("approved_at", BeNil()))

		adjustment := reflect.New(options.AdjustmentType)
		tx := hctx.Db.Where("review_state = 'draft'").Take(adjustment.Interface())
		Expect(dbutils.CreateFindOperationError(tx)).ToNot(HaveOccurred())
	})

	It("keeps the proposal as a draft if proposal_state is draft", func() {
		options.Setup(reviewstate.Draft)
		body := rctx.MakeRequest(false, "draft", 200)

		Expect(body).To(HaveKeyWithValue("version", Not(BeEmpty())))
		version := body["version"].(map[string]interface{})
		Expect(version).To(HaveKeyWithValue("version_state", "proposal"))
		Expect(version).To(HaveKeyWithValue("version_number", BeNil()))
		Expect(version).To(HaveKeyWithValue("adjustment_state", "draft"))
		Expect(version).To(HaveKeyWithValue("approved_at", BeNil()))

		adjustment := reflect.New(options.AdjustmentType)
		tx := hctx.Db.Where("review_state = 'draft'").Take(adjustment.Interface())
		Expect(dbutils.CreateFindOperationError(tx)).ToNot(HaveOccurred())
	})

	It("abandons the proposal if proposal_state is abandon", func() {
		options.Setup(reviewstate.Draft)
		body := rctx.MakeRequest(false, "abandon", 200)

		Expect(body).To(HaveKeyWithValue("version", Not(BeEmpty())))
		version := body["version"].(map[string]interface{})
		Expect(version).To(HaveKeyWithValue("version_state", "proposal"))
		Expect(version).To(HaveKeyWithValue("version_number", BeNil()))
		Expect(version).To(HaveKeyWithValue("adjustment_state", "abandoned"))
		Expect(version).To(HaveKeyWithValue("approved_at", BeNil()))

		adjustment := reflect.New(options.AdjustmentType)
		tx := hctx.Db.Where("review_state = 'abandoned'").Take(adjustment.Interface())
		Expect(dbutils.CreateFindOperationError(tx)).ToNot(HaveOccurred())
	})

	It("submits the proposal for approval if proposal_state is final", func() {
		options.Setup(reviewstate.Draft)
		body := rctx.MakeRequest(false, "final", 200)

		Expect(body).To(HaveKeyWithValue("version", Not(BeEmpty())))
		versionJSON := body["version"].(map[string]interface{})
		Expect(versionJSON).To(SatisfyAny(
			HaveKeyWithValue("version_state", "proposal"),
			HaveKeyWithValue("version_state", "approved"),
		))
		if versionJSON["version_state"] == "proposal" {
			Expect(versionJSON).To(HaveKeyWithValue("version_number", BeNil()))
			Expect(versionJSON).To(HaveKeyWithValue("adjustment_state", Or(
				Equal("draft"),
				Equal("reviewing"),
				Equal("rejected"),
				Equal("abandoned"),
			)))
			Expect(versionJSON).To(HaveKeyWithValue("approved_at", BeNil()))

			adjustment := reflect.New(options.AdjustmentType)
			tx := hctx.Db.Where("review_state != 'approved'").Take(adjustment.Interface())
			Expect(dbutils.CreateFindOperationError(tx)).ToNot(HaveOccurred())
		} else {
			Expect(versionJSON).To(HaveKeyWithValue("version_number", BeNumerically("==", 2)))
			Expect(versionJSON).To(HaveKeyWithValue("adjustment_state", "approved"))
			Expect(versionJSON).To(HaveKeyWithValue("approved_at", Not(BeNil())))

			version, adjustment := options.GetResourceVersionAndLatestAdjustment()
			Expect(version).ToNot(BeNil())
			Expect(version.GetVersionNumber()).ToNot(BeNil())
			Expect(*version.GetVersionNumber()).To(BeNumerically("==", 2))
			Expect(adjustment.GetReviewState()).To(Equal(reviewstate.Approved))
		}
	})

	It("returns a 422 if the requested adjustment is not allowed in the current proposal state", func() {
		options.Setup(reviewstate.Reviewing)
		body := rctx.MakeRequest(false, "final", 422)
		Expect(body).To(HaveKeyWithValue("error", ContainSubstring("Cannot finalize a proposal which is already being reviewed")))
	})

	return &rctx
}
