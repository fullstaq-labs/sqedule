package controllers

import (
	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
)

type ReviewableReadTestOptions struct {
	HTTPTestCtx *HTTPTestContext
	Path        string
	Setup       func()

	PrimaryKeyJSONFieldName string
	PrimaryKeyInitialValue  interface{}
}

type ReviewableReadTestContext struct {
	MakeRequest func() gin.H
}

func IncludeReviewableReadTest(options ReviewableReadTestOptions) *ReviewableReadTestContext {
	var rctx ReviewableReadTestContext
	var hctx *HTTPTestContext = options.HTTPTestCtx

	rctx.MakeRequest = func() gin.H {
		req, err := hctx.NewRequestWithAuth("GET", options.Path, nil)
		Expect(err).ToNot(HaveOccurred())
		hctx.ServeHTTP(req)
		Expect(hctx.HttpRecorder.Code).To(Equal(200))

		body, err := hctx.BodyJSON()
		Expect(err).ToNot(HaveOccurred())

		return body
	}

	It("outputs the latest approved version", func() {
		options.Setup()
		body := rctx.MakeRequest()

		Expect(body).To(HaveKeyWithValue(options.PrimaryKeyJSONFieldName, options.PrimaryKeyInitialValue))
		Expect(body).To(HaveKey("latest_approved_version"))
		Expect(body["latest_approved_version"]).ToNot(BeNil())

		version := body["latest_approved_version"].(map[string]interface{})
		Expect(version["version_number"]).To(BeNumerically("==", 1))
	})

	return &rctx
}
