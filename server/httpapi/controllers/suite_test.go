package controllers

import (
	"testing"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	Describe    = ginkgo.Describe
	PDescribe   = ginkgo.PDescribe
	FDescribe   = ginkgo.FDescribe
	BeforeEach  = ginkgo.BeforeEach
	BeforeSuite = ginkgo.BeforeSuite
	It          = ginkgo.It
	PIt         = ginkgo.PIt
	FIt         = ginkgo.FIt
	Specify     = ginkgo.Specify
	PSpecify    = ginkgo.PSpecify
	FSpecify    = ginkgo.FSpecify
)

func TestControllers(t *testing.T) {
	RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Controllers Suite")
}
