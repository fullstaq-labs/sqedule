package controllers

import (
	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
)

type ReviewableReadResourceTestOptions struct {
	HTTPTestCtx *HTTPTestContext
	Path        string
	Setup       func()

	PrimaryKeyJSONFieldName string
	PrimaryKeyInitialValue  interface{}
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

		Expect(body).To(HaveKeyWithValue(options.PrimaryKeyJSONFieldName, options.PrimaryKeyInitialValue))
		Expect(body).To(HaveKeyWithValue("latest_approved_version", Not(BeEmpty())))

		version := body["latest_approved_version"].(map[string]interface{})
		Expect(version).To(HaveKeyWithValue("version_number", BeNumerically("==", 1)))
	})

	return &rctx
}
