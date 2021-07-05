package controllers

import (
	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
)

type ReviewableDeleteProposalTestOptions struct {
	HTTPTestCtx                *HTTPTestContext
	GetProposalPath            func() string
	GetApprovedVersionPath     func() string
	Setup                      func()
	ResourceTypeNameInResponse string
	CountProposals             func() uint
	CountProposalAdjustments   func() uint
}

type ReviewableDeleteProposalTestContext struct {
	MakeRequest func(approved bool, expectedCode uint) gin.H
}

func IncludeReviewableDeleteProposalTest(options ReviewableDeleteProposalTestOptions) *ReviewableDeleteProposalTestContext {
	var rctx ReviewableDeleteProposalTestContext
	var hctx *HTTPTestContext = options.HTTPTestCtx

	rctx.MakeRequest = func(approved bool, expectedCode uint) gin.H {
		var path string
		if approved {
			path = options.GetApprovedVersionPath()
		} else {
			path = options.GetProposalPath()
		}

		req, err := hctx.NewRequestWithAuth("DELETE", path, nil)
		Expect(err).ToNot(HaveOccurred())
		hctx.ServeHTTP(req)
		Expect(hctx.Recorder.Code).To(BeNumerically("==", expectedCode))

		body, err := hctx.BodyJSON()
		Expect(err).ToNot(HaveOccurred())

		return body
	}

	It("deletes the proposal and all its adjustments", func() {
		options.Setup()
		Expect(options.CountProposals()).To(BeNumerically("==", 1))
		Expect(options.CountProposalAdjustments()).To(BeNumerically(">", 1))
		rctx.MakeRequest(false, 200)
		Expect(options.CountProposals()).To(BeNumerically("==", 0))
		Expect(options.CountProposalAdjustments()).To(BeNumerically("==", 0))
	})

	It("does not delete approved versions", func() {
		options.Setup()
		body := rctx.MakeRequest(true, 404)
		Expect(body).To(HaveKeyWithValue("error", options.ResourceTypeNameInResponse+" not found"))
	})

	return &rctx
}
