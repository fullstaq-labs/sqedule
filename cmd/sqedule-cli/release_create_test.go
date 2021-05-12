package main

import (
	encjson "encoding/json"
	"net/http"

	"github.com/fullstaq-labs/sqedule/lib/mocking"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/releasestate"
	"github.com/fullstaq-labs/sqedule/server/httpapi/json"

	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	viperPkg "github.com/spf13/viper"
)

var _ = Describe("release create", func() {
	const serverBaseURL = "http://server"
	const appID = "app1"

	var viper *viperPkg.Viper
	var printer mocking.FakePrinter

	isValidJSON := func(value string) bool {
		var data map[string]interface{}
		return encjson.Unmarshal([]byte(value), &data) == nil
	}

	BeforeEach(func() {
		httpmock.Reset()
		mockAuthToken()
		printer = mocking.FakePrinter{}

		viper = viperPkg.New()
		viper.Set("server-base-url", serverBaseURL)
		viper.Set("application-id", appID)
	})

	It("creates a release", func() {
		httpmock.RegisterResponder("POST", serverBaseURL+"/v1/applications/"+appID+"/releases", func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, json.ReleaseWithAssociations{})
			Expect(err).ToNot(HaveOccurred())
			return resp, nil
		})

		err := releaseCreateCmd_run(viper, &printer, true)
		Expect(err).ToNot(HaveOccurred())
		Expect(isValidJSON(printer.String())).To(BeTrue(), "it outputs the JSON response")
	})

	It("waits for the approval status to become final if --wait is set", func() {
		httpmock.RegisterResponder("POST", serverBaseURL+"/v1/applications/"+appID+"/releases", func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, json.ReleaseWithAssociations{
				Release: json.Release{
					ID:    1,
					State: string(releasestate.InProgress),
				},
			})
			Expect(err).ToNot(HaveOccurred())
			return resp, nil
		})

		viper.Set("wait", true)
		viper.Set("wait-interval", "1m")
		viper.Set("wait-timeout", "0")

		err := releaseCreateCmd_run(viper, &printer, true)
		Expect(err).ToNot(HaveOccurred())
		Expect(printer.String()).To(ContainSubstring("Waiting for the release's approval state to become final"))
	})
})
