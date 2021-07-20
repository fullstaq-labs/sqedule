package controllers

import (
	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
)

type ReviewableReadResourceTestOptions struct {
	HTTPTestCtx *HTTPTestContext
	Path        string
	Setup       func()

	AssertBaseResourceValid func(resource map[string]interface{})
	AssertVersionValid      func(version map[string]interface{})
}

type ReviewableReadResourceTestContext struct {
	MakeRequest func() gin.H
}

func IncludeReviewableReadResourceTest(options ReviewableReadResourceTestOptions) *ReviewableReadResourceTestContext {
	var rctx ReviewableReadResourceTestContext
	var hctx *HTTPTestContext = options.HTTPTestCtx

	rctx.MakeRequest = func() gin.H {
		req, err := hctx.NewRequestWithAuth("GET", options.Path, nil)
		Expect(err).ToNot(HaveOccurred())
		hctx.ServeHTTP(req)
		Expect(hctx.Recorder.Code).To(Equal(200))

		body, err := hctx.BodyJSON()
		Expect(err).ToNot(HaveOccurred())

		return body
	}

	It("outputs the latest approved version", func() {
		options.Setup()
		body := rctx.MakeRequest()

		Expect(body).To(HaveKeyWithValue("latest_approved_version", Not(BeEmpty())))

		version := body["latest_approved_version"].(map[string]interface{})
		Expect(version).To(HaveKeyWithValue("version_number", BeNumerically("==", 1)))

		if options.AssertVersionValid != nil {
			options.AssertVersionValid(version)
		}
		if options.AssertBaseResourceValid != nil {
			options.AssertBaseResourceValid(body)
		}
	})

	return &rctx
}
