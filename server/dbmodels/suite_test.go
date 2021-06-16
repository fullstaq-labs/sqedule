package dbmodels

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDbmodels(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dbmodels Suite")
}
