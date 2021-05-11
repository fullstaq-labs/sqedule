package main_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/fullstaq-labs/sqedule/cli"
	. "github.com/fullstaq-labs/sqedule/cmd/sqedule-cli"
)

var _ = Describe("Logout", func() {
	It("erases the authentication token", func() {
		cli.MockState.AuthToken = "abcd"
		cli.MockState.AuthTokenExpirationTime = time.Now()
		err := LogoutCmd_Run(nil, nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(cli.MockState.AuthToken).To(BeEmpty())
		Expect(cli.MockState.AuthTokenExpirationTime).To(Equal(time.Time{}))
	})
})
