package controllers

import (
	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
)

type ReviewableReadProposalTestOptions struct {
	HTTPTestCtx *HTTPTestContext
	GetPath     func() string
	Setup       func(approved bool)

	ResourceTypeNameInResponse string

	PrimaryKeyJSONFieldName string
	PrimaryKeyInitialValue  interface{}
}

type ReviewableReadProposalTestContext struct {
	MakeRequest func(expectedCode uint) gin.H
}

func IncludeReviewableReadProposalTest(options ReviewableReadProposalTestOptions) *ReviewableReadProposalTestContext {
	var rctx ReviewableReadProposalTestContext
	var hctx *HTTPTestContext = options.HTTPTestCtx

	rctx.MakeRequest = func(expectedCode uint) gin.H {
		req, err := hctx.NewRequestWithAuth("GET", options.GetPath(), nil)
		Expect(err).ToNot(HaveOccurred())
		hctx.ServeHTTP(req)
		Expect(hctx.Recorder.Code).To(BeNumerically("==", expectedCode))

		body, err := hctx.BodyJSON()
		Expect(err).ToNot(HaveOccurred())

		return body
	}

	It("outputs non-versioned fields", func() {
		options.Setup(false)
		body := rctx.MakeRequest(200)

		Expect(body).To(HaveKeyWithValue(options.PrimaryKeyJSONFieldName, options.PrimaryKeyInitialValue))
	})

	It("outputs the requested proposal", func() {
		options.Setup(false)
		body := rctx.MakeRequest(200)

		Expect(body).To(HaveKeyWithValue("version", Not(BeEmpty())))
		version := body["version"]

		Expect(version).To(HaveKeyWithValue("id", Not(BeNil())))
		Expect(version).To(HaveKeyWithValue("version_state", "proposal"))
		Expect(version).To(HaveKeyWithValue("version_number", BeNil()))
		Expect(version).To(HaveKeyWithValue("approved_at", BeNil()))
	})

	It("does not find approved versions", func() {
		options.Setup(true)
		body := rctx.MakeRequest(404)
		Expect(body).To(HaveKeyWithValue("error", options.ResourceTypeNameInResponse+" not found"))
	})

	return &rctx
}
