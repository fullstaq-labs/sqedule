package main_test

import (
	"encoding/json"
	"net/http"

	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	viperPkg "github.com/spf13/viper"

	. "github.com/fullstaq-labs/sqedule/cmd/sqedule-cli"
)

var _ = Describe("Login", func() {
	const serverBaseURL = "http://server"
	const orgID = "org1"
	const password = "123456"

	var viper *viperPkg.Viper

	BeforeEach(func() {
		httpmock.Reset()

		viper = viperPkg.New()
		viper.Set("server-base-url", serverBaseURL)
		viper.Set("organization-id", orgID)
		viper.Set("password", password)
	})

	It("works on user accounts", func() {
		viper.Set("email", "a@a.com")

		httpmock.RegisterResponder("POST", serverBaseURL+"/v1/auth/login", func(req *http.Request) (*http.Response, error) {
			input := make(map[string]interface{})
			err := json.NewDecoder(req.Body).Decode(&input)
			Expect(err).ToNot(HaveOccurred())

			Expect(input["organization_id"]).To(Equal(orgID))
			Expect(input["email"]).To(Equal("a@a.com"))
			Expect(input["password"]).To(Equal(password))

			resp, err := httpmock.NewJsonResponse(200, map[string]interface{}{
				"code":   200,
				"expire": "2021-05-11T16:14:58+02:00",
				"token":  "my token",
			})
			Expect(err).ToNot(HaveOccurred())
			return resp, nil
		})

		err := LoginCmd_Run(nil, nil, viper)
		Expect(err).ToNot(HaveOccurred())
	})

	It("works on service accounts", func() {
		viper.Set("service-account-name", "sa")

		httpmock.RegisterResponder("POST", serverBaseURL+"/v1/auth/login", func(req *http.Request) (*http.Response, error) {
			input := make(map[string]interface{})
			err := json.NewDecoder(req.Body).Decode(&input)
			Expect(err).ToNot(HaveOccurred())

			Expect(input["organization_id"]).To(Equal(orgID))
			Expect(input["service_account_name"]).To(Equal("sa"))
			Expect(input["password"]).To(Equal(password))

			resp, err := httpmock.NewJsonResponse(200, map[string]interface{}{
				"code":   200,
				"expire": "2021-05-11T16:14:58+02:00",
				"token":  "my token",
			})
			Expect(err).ToNot(HaveOccurred())
			return resp, nil
		})

		err := LoginCmd_Run(nil, nil, viper)
		Expect(err).ToNot(HaveOccurred())
	})
})
