package controllers

import (
	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
)

type ReviewableReadAllProposalsTestOptions struct {
	HTTPTestCtx *HTTPTestContext
	Path        string
	Setup       func(approved bool)
}

type ReviewableReadAllProposalsTestContext struct {
	MakeRequest func() gin.H
}

func IncludeReviewableReadAllProposalsTest(options ReviewableReadAllProposalsTestOptions) *ReviewableReadAllProposalsTestContext {
	var rctx ReviewableReadAllProposalsTestContext
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

	It("outputs all proposals", func() {
		options.Setup(false)
		body := rctx.MakeRequest()

		Expect(body["items"]).To(HaveLen(1))

		items := body["items"].([]interface{})
		version := items[0].(map[string]interface{})
		Expect(version["id"]).ToNot(BeNil())
		Expect(version["version_state"]).To(Equal("proposal"))
		Expect(version["version_number"]).To(BeNil())
		Expect(version["approved_at"]).To(BeNil())
	})

	It("does not output approved versions", func() {
		options.Setup(true)
		body := rctx.MakeRequest()
		Expect(body["items"]).To(HaveLen(0))
	})

	return &rctx
}
