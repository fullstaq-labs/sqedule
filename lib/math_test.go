package lib

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Math", func() {
	Describe("AddAndClamp", func() {
		When("there's no overflow", func() {
			Describe("for additions", func() {
				It("works", func() {
					Expect(AddAndClamp(0, 0, 0, MaxUint64)).To(Equal(uint64(0)))
					Expect(AddAndClamp(1, 2, 0, MaxUint64)).To(Equal(uint64(3)))
					Expect(AddAndClamp(uint64(MaxInt64)+1, MaxInt64, 0, MaxUint64)).To(Equal(MaxUint64))
				})

				It("clamps the result between min and max", func() {
					Expect(AddAndClamp(0, 0, 1, 3)).To(Equal(uint64(1)))
					Expect(AddAndClamp(0, 4, 1, 3)).To(Equal(uint64(3)))
				})
			})

			Describe("for substractions", func() {
				It("works", func() {
					Expect(AddAndClamp(2, -1, 0, MaxUint64)).To(Equal(uint64(1)))
					Expect(AddAndClamp(2, -2, 0, MaxUint64)).To(Equal(uint64(0)))
				})

				It("clamps the result between min and max", func() {
					Expect(AddAndClamp(10, -10, 1, 3)).To(Equal(uint64(1)))
					Expect(AddAndClamp(10, 0, 1, 3)).To(Equal(uint64(3)))
				})
			})
		})

		When("positive overflow could occur", func() {
			It("prevents overflowing", func() {
				Expect(AddAndClamp(MaxUint64, 1, 0, MaxUint64)).To(Equal(MaxUint64))
			})

			It("clamps the result between min and max", func() {
				Expect(AddAndClamp(MaxUint64, 1, 0, 10)).To(Equal(uint64(10)))
			})
		})

		When("negative overflow could occur", func() {
			It("prevents overflowing", func() {
				Expect(AddAndClamp(2, -3, 0, MaxUint64)).To(Equal(uint64(0)))
			})

			It("clamps the result between min and max", func() {
				Expect(AddAndClamp(2, -3, 1, 10)).To(Equal(uint64(1)))
			})
		})
	})
})
