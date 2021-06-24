package controllers

import (
	"reflect"

	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/reviewstate"
	"github.com/fullstaq-labs/sqedule/server/dbutils"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
)

type ReviewableUpdateUnversionedDataTestOptions struct {
	HTTPTestCtx *HTTPTestContext
	Path        string
	Setup       func()

	UnversionedInput gin.H

	ResourceType reflect.Type

	GetPrimaryKey           func(resource interface{}) interface{}
	PrimaryKeyJSONFieldName string
	PrimaryKeyUpdatedValue  interface{}
}

type ReviewableUpdateUnversionedDataTestContext struct {
	MakeRequest func(expectedCode uint) gin.H
}

func IncludeReviewableUpdateUnversionedDataTest(options ReviewableUpdateUnversionedDataTestOptions) *ReviewableUpdateUnversionedDataTestContext {
	var rctx ReviewableUpdateUnversionedDataTestContext
	var hctx *HTTPTestContext = options.HTTPTestCtx

	rctx.MakeRequest = func(expectedCode uint) gin.H {
		input := gin.H{}
		for k, v := range options.UnversionedInput {
			input[k] = v
		}

		req, err := hctx.NewRequestWithAuth("PATCH", options.Path, input)
		Expect(err).ToNot(HaveOccurred())
		hctx.ServeHTTP(req)

		Expect(hctx.HttpRecorder.Code).To(BeNumerically("==", expectedCode))
		body, err := hctx.BodyJSON()
		Expect(err).ToNot(HaveOccurred())

		return body
	}

	It("patches the resource", func() {
		options.Setup()
		body := rctx.MakeRequest(200)

		Expect(body).To(HaveKeyWithValue(options.PrimaryKeyJSONFieldName, options.PrimaryKeyUpdatedValue))

		resource := reflect.New(options.ResourceType)
		tx := hctx.Db.Take(resource.Interface())
		Expect(dbutils.CreateFindOperationError(tx)).ToNot(HaveOccurred())
		Expect(options.GetPrimaryKey(resource.Interface())).To(Equal(options.PrimaryKeyUpdatedValue))
	})

	return &rctx
}

type ReviewableUpdateVersionedDataTestOptions struct {
	HTTPTestCtx *HTTPTestContext
	Path        string
	Setup       func()

	VersionedInput gin.H

	AdjustmentType reflect.Type

	GetLatestResourceVersionAndAdjustment func() (dbmodels.IReviewableVersion, dbmodels.IReviewableAdjustment)
	VersionedFieldJSONFieldName           string
	VersionedFieldUpdatedValue            interface{}
}

type ReviewableUpdateVersionedDataTestContext struct {
	MakeRequest func(proposalState string, expectedCode uint) gin.H
}

func IncludeReviewableUpdateVersionedDataTest(options ReviewableUpdateVersionedDataTestOptions) *ReviewableUpdateVersionedDataTestContext {
	var rctx ReviewableUpdateVersionedDataTestContext
	var hctx *HTTPTestContext = options.HTTPTestCtx

	rctx.MakeRequest = func(proposalState string, expectedCode uint) gin.H {
		versionInput := gin.H{}
		for k, v := range options.VersionedInput {
			versionInput[k] = v
		}
		if len(proposalState) > 0 {
			versionInput["proposal_state"] = proposalState
		}

		req, err := hctx.NewRequestWithAuth("PATCH", options.Path, gin.H{"version": versionInput})
		Expect(err).ToNot(HaveOccurred())
		hctx.ServeHTTP(req)

		Expect(hctx.HttpRecorder.Code).To(BeNumerically("==", expectedCode))
		body, err := hctx.BodyJSON()
		Expect(err).ToNot(HaveOccurred())

		return body
	}

	It("outputs the created resource version", func() {
		options.Setup()
		body := rctx.MakeRequest("", 200)

		Expect(body).To(HaveKeyWithValue("version", Not(BeEmpty())))
		version := body["version"].(map[string]interface{})
		Expect(version).To(HaveKeyWithValue(options.VersionedFieldJSONFieldName, options.VersionedFieldUpdatedValue))
	})

	It("creates a draft proposal by default", func() {
		options.Setup()
		body := rctx.MakeRequest("", 200)

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

	It("creates a draft proposal if proposal_state is draft", func() {
		options.Setup()
		body := rctx.MakeRequest("draft", 200)

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

	It("creates an abandoned proposal if proposal_state is abandon", func() {
		options.Setup()
		body := rctx.MakeRequest("abandon", 200)

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

	It("submits the version for approval if proposal_state is final", func() {
		options.Setup()
		body := rctx.MakeRequest("final", 200)

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

			version, adjustment := options.GetLatestResourceVersionAndAdjustment()
			Expect(version).ToNot(BeNil())
			Expect(version.GetVersionNumber()).ToNot(BeNil())
			Expect(*version.GetVersionNumber()).To(BeNumerically("==", 2))
			Expect(adjustment.GetReviewState()).To(Equal(reviewstate.Approved))
		}
	})

	return &rctx
}
