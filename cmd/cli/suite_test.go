package main

import (
	"testing"
	"time"

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

func mockAuthToken() {
	cli.MockState.AuthToken = "test"
	cli.MockState.AuthTokenExpirationTime = time.Now().Add(1 * time.Hour)
}
