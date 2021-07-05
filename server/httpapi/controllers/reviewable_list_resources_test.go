package controllers

import (
	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
)

type ReviewableListResourcesTestOptions struct {
	HTTPTestCtx *HTTPTestContext
	Path        string
	Setup       func()

	PrimaryKeyJSONFieldName string
	PrimaryKeyInitialValue  interface{}
}

type ReviewableListResourcesTestContext struct {
	MakeRequest func() gin.H
}

func IncludeReviewableListResourcesTest(options ReviewableListResourcesTestOptions) *ReviewableListResourcesTestContext {
	var rctx ReviewableListResourcesTestContext
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

	It("outputs all resources with their latest approved versions", func() {
		options.Setup()
		body := rctx.MakeRequest()

		Expect(body).To(HaveKeyWithValue("items", HaveLen(1)))

		items := body["items"].([]interface{})
		ruleset := items[0].(map[string]interface{})
		Expect(ruleset).To(HaveKeyWithValue(options.PrimaryKeyJSONFieldName, options.PrimaryKeyInitialValue))
		Expect(ruleset).To(HaveKeyWithValue("latest_approved_version", Not(BeEmpty())))

		version := ruleset["latest_approved_version"].(map[string]interface{})
		Expect(version).To(HaveKeyWithValue("version_number", BeNumerically("==", 1)))
	})

	return &rctx
}
