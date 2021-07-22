package controllers

import (
	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
)

type ReviewableListProposalsTestOptions struct {
	HTTPTestCtx *HTTPTestContext
	Path        string
	Setup       func(versionIsApproved bool)
}

type ReviewableListProposalsTestContext struct {
	MakeRequest func() gin.H
}

func IncludeReviewableListProposalsTest(options ReviewableListProposalsTestOptions) *ReviewableListProposalsTestContext {
	var rctx ReviewableListProposalsTestContext
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

	It("outputs all proposals", func() {
		options.Setup(false)
		body := rctx.MakeRequest()

		Expect(body).To(HaveKeyWithValue("items", HaveLen(1)))

		items := body["items"].([]interface{})
		version := items[0].(map[string]interface{})
		Expect(version).To(HaveKeyWithValue("id", Not(BeNil())))
		Expect(version).To(HaveKeyWithValue("version_state", "proposal"))
		Expect(version).To(HaveKeyWithValue("version_number", BeNil()))
		Expect(version).To(HaveKeyWithValue("approved_at", BeNil()))
	})

	It("does not output approved versions", func() {
		options.Setup(true)
		body := rctx.MakeRequest()
		Expect(body).To(HaveKeyWithValue("items", HaveLen(0)))
	})

	return &rctx
}
