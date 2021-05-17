package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/fullstaq-labs/sqedule/lib/mocking"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/releasestate"
	"github.com/fullstaq-labs/sqedule/server/httpapi/json"

	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	viperPkg "github.com/spf13/viper"
)

var _ = Describe("release wait", func() {
	const serverBaseURL = "http://server"
	const appID = "app1"
	const releaseID = 1

	var viper *viperPkg.Viper
	var printer mocking.FakePrinter
	var clock mocking.FakeClock

	BeforeEach(func() {
		httpmock.Reset()
		mockAuthToken()
		printer = mocking.FakePrinter{}
		clock = mocking.FakeClock{Value: time.Now()}

		viper = viperPkg.New()
		viper.Set("server-base-url", serverBaseURL)
		viper.Set("application-id", appID)
		viper.Set("release-id", releaseID)
		viper.Set("wait-min-duration", "1s")
		viper.Set("wait-max-duration", "1s")
		viper.Set("wait-timeout", "0")
	})

	It("waits until the approval status is final", func() {
		var count = 0

		url := fmt.Sprintf("%s/v1/applications/%s/releases/%d", serverBaseURL, appID, releaseID)
		httpmock.RegisterResponder("GET", url, func(req *http.Request) (*http.Response, error) {
			var state releasestate.State
			if count == 0 {
				state = releasestate.InProgress
			} else {
				state = releasestate.Approved
			}
			count++

			resp, err := httpmock.NewJsonResponse(200, json.ReleaseWithAssociations{
				Release: json.Release{
					ID:    releaseID,
					State: string(state),
				},
			})
			Expect(err).ToNot(HaveOccurred())
			return resp, nil
		})

		state, err := releaseWaitCmd_run(viper, &printer, &clock, true)
		Expect(err).ToNot(HaveOccurred())
		Expect(state).To(Equal(releasestate.Approved))

		output := printer.String()
		Expect(output).To(ContainSubstring("Current state: in_progress"))
		Expect(output).To(ContainSubstring("Current state: approved"))
	})

	It("times out when the approval status doesn't become final", func() {
		url := fmt.Sprintf("%s/v1/applications/%s/releases/%d", serverBaseURL, appID, releaseID)
		httpmock.RegisterResponder("GET", url, func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, json.ReleaseWithAssociations{
				Release: json.Release{
					ID:    releaseID,
					State: string(releasestate.InProgress),
				},
			})
			Expect(err).ToNot(HaveOccurred())
			return resp, nil
		})

		viper.Set("wait-timeout", "1s")
		_, err := releaseWaitCmd_run(viper, &printer, &clock, true)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("Timeout"))
	})
})
