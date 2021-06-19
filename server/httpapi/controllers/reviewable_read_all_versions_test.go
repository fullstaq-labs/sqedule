package controllers

import (
	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
)

type ReviewableReadAllVersionsTestOptions struct {
	HTTPTestCtx *HTTPTestContext
	Path        string
	Setup       func(approved bool)
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

	It("outputs all versions", func() {
		options.Setup(true)
		body := rctx.MakeRequest()

		Expect(body["items"]).To(HaveLen(3))

		items := body["items"].([]interface{})
		version := items[0].(map[string]interface{})
		Expect(version["id"]).ToNot(BeNil())
		Expect(version["version_state"]).To(Equal("approved"))
		Expect(version["version_number"]).ToNot(BeNil())
		Expect(version["approved_at"]).ToNot(BeNil())
	})

	It("outputs versions in descending order", func() {
		options.Setup(true)
		body := rctx.MakeRequest()

		Expect(body["items"]).To(HaveLen(3))

		items := body["items"].([]interface{})

		version3 := items[0].(map[string]interface{})
		Expect(version3["version_number"]).To(BeNumerically("==", 3))
		version2 := items[1].(map[string]interface{})
		Expect(version2["version_number"]).To(BeNumerically("==", 2))
		version1 := items[2].(map[string]interface{})
		Expect(version1["version_number"]).To(BeNumerically("==", 1))
	})

	It("does not output proposed versions", func() {
		options.Setup(false)
		body := rctx.MakeRequest()

		Expect(body["items"]).To(HaveLen(0))
	})

	return &rctx
}
