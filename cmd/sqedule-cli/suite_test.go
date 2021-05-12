package main

import (
	"testing"

	"github.com/fullstaq-labs/sqedule/cli"
	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSqeduleCli(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SqeduleCLI Suite")
}

var _ = BeforeSuite(func() {
	cli.MockState = &cli.State{}
	cli.MockHttpClientFunc = httpmock.ActivateNonDefault
})

var _ = AfterSuite(func() {
	httpmock.DeactivateAndReset()
	cli.MockHttpClientFunc = nil
	cli.MockState = nil
})

var _ = BeforeEach(func() {
	cli.MockState = &cli.State{}
})
