package controllers

import (
	"reflect"

	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/reviewstate"
	"github.com/fullstaq-labs/sqedule/server/dbutils"
	"github.com/fullstaq-labs/sqedule/server/httpapi/json/proposalstate"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
)

type ReviewableCreateTestOptions struct {
	HTTPTestCtx      *HTTPTestContext
	Path             string
	UnversionedInput gin.H
	VersionedInput   gin.H

	ResourceType   reflect.Type
	AdjustmentType reflect.Type

	GetPrimaryKey               func(resource interface{}) interface{}
	PrimaryKeyJSONFieldName     string
	PrimaryKeyInitialValue      interface{}
	GetVersionedField           func(adjustment interface{}) interface{}
	VersionedFieldJSONFieldName string
	VersionedFieldInitialValue  interface{}
}

type ReviewableCreateTestContext struct {
	MakeRequest func(proposalState string, expectedCode int) gin.H
}

func IncludeReviewableCreateTest(options ReviewableCreateTestOptions) *ReviewableCreateTestContext {
	var rctx ReviewableCreateTestContext
	var hctx *HTTPTestContext = options.HTTPTestCtx

	rctx.MakeRequest = func(proposalState string, expectedCode int) gin.H {
		input := gin.H{}
		for k, v := range options.UnversionedInput {
			input[k] = v
		}

		versionInput := gin.H{}
		for k, v := range options.VersionedInput {
			versionInput[k] = v
		}
		if len(proposalState) > 0 {
			versionInput["proposal_state"] = proposalState
		}
		input["version"] = versionInput

		req, err := hctx.NewRequestWithAuth("POST", options.Path, input)
		Expect(err).ToNot(HaveOccurred())
		hctx.ServeHTTP(req)

		Expect(hctx.HttpRecorder.Code).To(Equal(expectedCode))
		body, err := hctx.BodyJSON()
		Expect(err).ToNot(HaveOccurred())

		return body
	}

	It("outputs the created resource", func() {
		body := rctx.MakeRequest("", 201)

		Expect(body).To(HaveKeyWithValue(options.PrimaryKeyJSONFieldName, options.PrimaryKeyInitialValue))
		Expect(body["version"]).ToNot(BeNil())

		version := body["version"].(map[string]interface{})
		Expect(version).To(HaveKeyWithValue(options.VersionedFieldJSONFieldName, options.VersionedFieldInitialValue))
	})

	It("creates a new resource", func() {
		rctx.MakeRequest("", 201)

		resource := reflect.New(options.ResourceType)
		tx := hctx.Db.Take(resource.Interface())
		Expect(dbutils.CreateFindOperationError(tx)).ToNot(HaveOccurred())
		Expect(options.GetPrimaryKey(resource.Interface())).To(Equal(options.PrimaryKeyInitialValue))

		adjustment := reflect.New(options.AdjustmentType)
		tx = hctx.Db.Take(adjustment.Interface())
		Expect(dbutils.CreateFindOperationError(tx)).ToNot(HaveOccurred())
		Expect(options.GetVersionedField(adjustment.Interface())).To(Equal(options.VersionedFieldInitialValue))
	})

	It("creates a draft proposal by default", func() {
		body := rctx.MakeRequest("", 201)

		Expect(body["version"]).ToNot(BeNil())
		version := body["version"].(map[string]interface{})
		Expect(version).To(HaveKeyWithValue("version_state", "proposal"))
		Expect(version["version_number"]).To(BeNil())
		Expect(version["approved_at"]).To(BeNil())

		adjustment := reflect.New(options.AdjustmentType)
		tx := hctx.Db.Take(adjustment.Interface())
		Expect(dbutils.CreateFindOperationError(tx)).ToNot(HaveOccurred())
		Expect(adjustment.Interface().(dbmodels.IReviewableAdjustment).GetReviewState()).To(Equal(reviewstate.Draft))
	})

	It("creates a draft proposal if proposal_state is draft", func() {
		body := rctx.MakeRequest("draft", 201)

		Expect(body["version"]).ToNot(BeNil())
		version := body["version"].(map[string]interface{})
		Expect(version).To(HaveKeyWithValue("version_state", "proposal"))
		Expect(version["version_number"]).To(BeNil())
		Expect(version["approved_at"]).To(BeNil())

		adjustment := reflect.New(options.AdjustmentType)
		tx := hctx.Db.Take(adjustment.Interface())
		Expect(dbutils.CreateFindOperationError(tx)).ToNot(HaveOccurred())
		Expect(adjustment.Interface().(dbmodels.IReviewableAdjustment).GetReviewState()).To(Equal(reviewstate.Draft))
	})

	It("submits the version for approval if proposal_state is final", func() {
		body := rctx.MakeRequest("final", 201)

		Expect(body["version"]).ToNot(BeNil())
		version := body["version"].(map[string]interface{})
		Expect(version).To(SatisfyAny(
			HaveKeyWithValue("version_state", "reviewing"),
			HaveKeyWithValue("version_state", "approved"),
		))
		if version["version_state"] == "reviewing" {
			Expect(version["version_number"]).To(BeNil())
			Expect(version["approved_at"]).To(BeNil())

			adjustment := reflect.New(options.AdjustmentType)
			tx := hctx.Db.Where("review_state = 'reviewing'").Take(adjustment.Interface())
			Expect(dbutils.CreateFindOperationError(tx)).ToNot(HaveOccurred())
		} else {
			Expect(version["version_number"]).To(BeNumerically("==", 1))
			Expect(version["approved_at"]).ToNot(BeNil())

			adjustment := reflect.New(options.AdjustmentType)
			tx := hctx.Db.Take(adjustment.Interface())
			Expect(dbutils.CreateFindOperationError(tx)).ToNot(HaveOccurred())
			Expect(adjustment.Interface().(dbmodels.IReviewableAdjustment).GetReviewState()).To(
				SatisfyAny(Equal(reviewstate.Reviewing), Equal(reviewstate.Approved)))
		}
	})

	It("requires the 'version' input field", func() {
		input := gin.H{}
		for k, v := range options.UnversionedInput {
			input[k] = v
		}

		req, err := hctx.NewRequestWithAuth("POST", options.Path, input)
		Expect(err).ToNot(HaveOccurred())
		hctx.ServeHTTP(req)

		Expect(hctx.HttpRecorder.Code).To(Equal(400))
		body, err := hctx.BodyJSON()
		Expect(err).ToNot(HaveOccurred())

		Expect(body).To(HaveKeyWithValue("error", ContainSubstring("'version' field must be set")))
	})

	It("rejects proposal_state values other than unset, draft and final", func() {
		body := rctx.MakeRequest(string(proposalstate.Abandon), 400)
		Expect(body).To(HaveKeyWithValue("error", ContainSubstring("version.proposal_state must be either draft or final ('abandon' given)")))
	})

	return &rctx
}
