package approvalrulesprocessing

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestApprovalRulesProcessing(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ApprovalRulesProcessing Suite")
}
