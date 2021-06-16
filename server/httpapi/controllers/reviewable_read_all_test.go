package controllers

import (
	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
)

type ReviewableReadAllTestOptions struct {
	HTTPTestCtx *HTTPTestContext
	Path        string
	Setup       func()

	PrimaryKeyJSONFieldName string
	PrimaryKeyInitialValue  interface{}
}

type ReviewableReadAllTestContext struct {
	MakeRequest func() gin.H
}

func IncludeReviewableReadAllTest(options ReviewableReadAllTestOptions) *ReviewableReadAllTestContext {
	var rctx ReviewableReadAllTestContext
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

	It("outputs all resources with their latest approved versions", func() {
		options.Setup()
		body := rctx.MakeRequest()

		Expect(body["items"]).To(HaveLen(1))

		items := body["items"].([]interface{})
		ruleset := items[0].(map[string]interface{})
		Expect(ruleset).To(HaveKeyWithValue(options.PrimaryKeyJSONFieldName, options.PrimaryKeyInitialValue))
		Expect(ruleset).To(HaveKey("latest_approved_version"))
		Expect(ruleset["latest_approved_version"]).ToNot(BeNil())

		version := ruleset["latest_approved_version"].(map[string]interface{})
		Expect(version["version_number"]).To(BeNumerically("==", 1))
	})

	return &rctx
}
