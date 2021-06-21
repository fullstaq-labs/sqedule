package controllers

import (
	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
)

type ReviewableReadVersionTestOptions struct {
	HTTPTestCtx *HTTPTestContext
	Path        string
	Setup       func()

	PrimaryKeyJSONFieldName string
	PrimaryKeyInitialValue  interface{}
}

type ReviewableReadVersionTestContext struct {
	MakeRequest func() gin.H
}

func IncludeReviewableReadVersionTest(options ReviewableReadVersionTestOptions) *ReviewableReadVersionTestContext {
	var rctx ReviewableReadVersionTestContext
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

	It("outputs the requested version", func() {
		options.Setup()
		body := rctx.MakeRequest()

		Expect(body).To(HaveKeyWithValue(options.PrimaryKeyJSONFieldName, options.PrimaryKeyInitialValue))

		Expect(body).To(HaveKeyWithValue("version", Not(BeEmpty())))
		version := body["version"]

		Expect(version).To(HaveKeyWithValue("id", Not(BeNil())))
		Expect(version).To(HaveKeyWithValue("version_state", "approved"))
		Expect(version).To(HaveKeyWithValue("version_number", BeNumerically("==", 1)))
		Expect(version).To(HaveKeyWithValue("approved_at", Not(BeNil())))
	})

	return &rctx
}
