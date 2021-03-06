package volume_test

import (
	"errors"

	. "github.com/concourse/baggageclaim/volume"
	"github.com/concourse/baggageclaim/volume/fakes"
	"github.com/pivotal-golang/lager/lagertest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("EmptyStrategy", func() {
	var (
		strategy Strategy
	)

	BeforeEach(func() {
		strategy = EmptyStrategy{}
	})

	Describe("Materialize", func() {
		var (
			fakeFilesystem *fakes.FakeFilesystem

			materializedVolume FilesystemInitVolume
			materializeErr     error
		)

		BeforeEach(func() {
			fakeFilesystem = new(fakes.FakeFilesystem)
		})

		JustBeforeEach(func() {
			materializedVolume, materializeErr = strategy.Materialize(
				lagertest.NewTestLogger("test"),
				"some-volume",
				fakeFilesystem,
			)
		})

		Context("when creating the new volume succeeds", func() {
			var fakeVolume *fakes.FakeFilesystemInitVolume

			BeforeEach(func() {
				fakeFilesystem.NewVolumeReturns(fakeVolume, nil)
			})

			It("succeeds", func() {
				Expect(materializeErr).ToNot(HaveOccurred())
			})

			It("returns it", func() {
				Expect(materializedVolume).To(Equal(fakeVolume))
			})

			It("created it with the correct handle", func() {
				handle := fakeFilesystem.NewVolumeArgsForCall(0)
				Expect(handle).To(Equal("some-volume"))
			})
		})

		Context("when creating the new volume fails", func() {
			disaster := errors.New("nope")

			BeforeEach(func() {
				fakeFilesystem.NewVolumeReturns(nil, disaster)
			})

			It("returns the error", func() {
				Expect(materializeErr).To(Equal(disaster))
			})
		})
	})
})
