package controllers

import (
	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
)

type ReviewableReadAllVersionsTestOptions struct {
	HTTPTestCtx *HTTPTestContext
	Path        string
	Setup       func()
}

type ReviewableReadAllVersionsTestContext struct {
	MakeRequest func() gin.H
}

func IncludeReviewableReadAllVersionsTest(options ReviewableReadAllVersionsTestOptions) *ReviewableReadAllVersionsTestContext {
	var rctx ReviewableReadAllVersionsTestContext
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

	It("outputs versions", func() {
		options.Setup()
		body := rctx.MakeRequest()

		Expect(body["items"]).To(HaveLen(1))

		items := body["items"].([]interface{})
		version := items[0].(map[string]interface{})
		Expect(version["version_number"]).To(BeNumerically("==", 1))
	})

	return &rctx
}
