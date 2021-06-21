package controllers

import (
	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
)

type ReviewableReadProposalTestOptions struct {
	HTTPTestCtx *HTTPTestContext
	GetPath     func() string
	Setup       func()

	PrimaryKeyJSONFieldName string
	PrimaryKeyInitialValue  interface{}
}

type ReviewableReadProposalTestContext struct {
	MakeRequest func() gin.H
}

func IncludeReviewableReadProposalTest(options ReviewableReadProposalTestOptions) *ReviewableReadProposalTestContext {
	var rctx ReviewableReadProposalTestContext
	var hctx *HTTPTestContext = options.HTTPTestCtx

	rctx.MakeRequest = func() gin.H {
		req, err := hctx.NewRequestWithAuth("GET", options.GetPath(), nil)
		Expect(err).ToNot(HaveOccurred())
		hctx.ServeHTTP(req)
		Expect(hctx.HttpRecorder.Code).To(Equal(200))

		body, err := hctx.BodyJSON()
		Expect(err).ToNot(HaveOccurred())

		return body
	}

	It("outputs the requested proposal", func() {
		options.Setup()
		body := rctx.MakeRequest()

		Expect(body).To(HaveKeyWithValue(options.PrimaryKeyJSONFieldName, options.PrimaryKeyInitialValue))

		Expect(body).To(HaveKeyWithValue("version", Not(BeEmpty())))
		version := body["version"]

		Expect(version).To(HaveKeyWithValue("id", Not(BeNil())))
		Expect(version).To(HaveKeyWithValue("version_state", "proposal"))
		Expect(version).To(HaveKeyWithValue("version_number", BeNil()))
		Expect(version).To(HaveKeyWithValue("approved_at", BeNil()))
	})

	return &rctx
}
