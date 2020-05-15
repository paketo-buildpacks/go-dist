package main_test

import (
	"testing"
	"time"

	main "github.com/paketo-buildpacks/go-compiler"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testClock(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	context("Now", func() {
		it("returns the value from the given Now function", func() {
			now := time.Now()

			clock := main.NewClock(func() time.Time {
				return now
			})

			Expect(clock.Now()).To(Equal(now))
		})
	})
}
