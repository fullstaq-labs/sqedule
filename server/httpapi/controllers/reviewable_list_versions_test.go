package controllers

import (
	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
)

type ReviewableListVersionsTestOptions struct {
	HTTPTestCtx *HTTPTestContext
	Path        string
	Setup       func(approved bool)
}

type ReviewableListVersionsTestContext struct {
	MakeRequest func() gin.H
}

func IncludeReviewableListVersionsTest(options ReviewableListVersionsTestOptions) *ReviewableListVersionsTestContext {
	var rctx ReviewableListVersionsTestContext
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

		Expect(body).To(HaveKeyWithValue("items", HaveLen(3)))

		items := body["items"].([]interface{})
		version := items[0].(map[string]interface{})
		Expect(version).To(HaveKeyWithValue("id", Not(BeNil())))
		Expect(version).To(HaveKeyWithValue("version_state", "approved"))
		Expect(version).To(HaveKeyWithValue("version_number", Not(BeNil())))
		Expect(version).To(HaveKeyWithValue("approved_at", Not(BeNil())))
	})

	It("outputs versions in descending order", func() {
		options.Setup(true)
		body := rctx.MakeRequest()

		Expect(body).To(HaveKeyWithValue("items", HaveLen(3)))

		items := body["items"].([]interface{})

		version3 := items[0].(map[string]interface{})
		Expect(version3).To(HaveKeyWithValue("version_number", BeNumerically("==", 3)))
		version2 := items[1].(map[string]interface{})
		Expect(version2).To(HaveKeyWithValue("version_number", BeNumerically("==", 2)))
		version1 := items[2].(map[string]interface{})
		Expect(version1).To(HaveKeyWithValue("version_number", BeNumerically("==", 1)))
	})

	It("does not output proposed versions", func() {
		options.Setup(false)
		body := rctx.MakeRequest()

		Expect(body).To(HaveKeyWithValue("items", HaveLen(0)))
	})

	return &rctx
}
