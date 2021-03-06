package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-golang/lager/lagertest"

	"github.com/concourse/baggageclaim"
	"github.com/concourse/baggageclaim/api"
	"github.com/concourse/baggageclaim/uidjunk"
	"github.com/concourse/baggageclaim/volume"
	"github.com/concourse/baggageclaim/volume/driver"
)

var _ = Describe("Volume Server", func() {
	var (
		handler http.Handler

		volumeDir string
		tempDir   string
	)

	BeforeEach(func() {
		var err error

		tempDir, err = ioutil.TempDir("", fmt.Sprintf("baggageclaim_volume_dir_%d", GinkgoParallelNode()))
		Expect(err).NotTo(HaveOccurred())

		volumeDir = tempDir
	})

	JustBeforeEach(func() {
		logger := lagertest.NewTestLogger("volume-server")

		fs, err := volume.NewFilesystem(&driver.NaiveDriver{}, volumeDir)
		Expect(err).NotTo(HaveOccurred())

		repo := volume.NewRepository(
			logger,
			fs,
			volume.NewLockManager(),
		)

		strategerizer := volume.NewStrategerizer(&uidjunk.NoopNamespacer{})

		handler, err = api.NewHandler(logger, strategerizer, repo)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := os.RemoveAll(tempDir)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("listing the volumes", func() {
		var recorder *httptest.ResponseRecorder

		JustBeforeEach(func() {
			recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("GET", "/volumes", nil)

			handler.ServeHTTP(recorder, request)
		})

		Context("when there are no volumes", func() {
			It("returns an empty array", func() {
				Expect(recorder.Body).To(MatchJSON(`[]`))
			})
		})
	})

	Describe("querying for volumes with properties", func() {
		props := baggageclaim.VolumeProperties{
			"property-query": "value",
		}

		It("finds volumes that have a property", func() {
			body := &bytes.Buffer{}

			err := json.NewEncoder(body).Encode(baggageclaim.VolumeRequest{
				Strategy: encStrategy(map[string]string{
					"type": "empty",
				}),
				Properties: props,
			})
			Expect(err).NotTo(HaveOccurred())

			recorder := httptest.NewRecorder()
			request, _ := http.NewRequest("POST", "/volumes", body)
			handler.ServeHTTP(recorder, request)
			Expect(recorder.Code).To(Equal(201))

			body.Reset()
			err = json.NewEncoder(body).Encode(baggageclaim.VolumeRequest{
				Strategy: encStrategy(map[string]string{
					"type": "empty",
				}),
			})
			Expect(err).NotTo(HaveOccurred())

			recorder = httptest.NewRecorder()
			request, _ = http.NewRequest("POST", "/volumes", body)
			handler.ServeHTTP(recorder, request)
			Expect(recorder.Code).To(Equal(201))

			recorder = httptest.NewRecorder()
			request, _ = http.NewRequest("GET", "/volumes?property-query=value", nil)
			handler.ServeHTTP(recorder, request)
			Expect(recorder.Code).To(Equal(200))

			var volumes volume.Volumes
			err = json.NewDecoder(recorder.Body).Decode(&volumes)
			Expect(err).NotTo(HaveOccurred())

			Expect(volumes).To(HaveLen(1))
		})

		It("returns an error if an invalid set of properties are specified", func() {
			recorder := httptest.NewRecorder()
			request, _ := http.NewRequest("GET", "/volumes?property-query=value&property-query=another-value", nil)
			handler.ServeHTTP(recorder, request)

			Expect(recorder.Code).To(Equal(422))
		})
	})

	Describe("updating a volume", func() {
		It("can have it's properties updated", func() {
			body := &bytes.Buffer{}

			err := json.NewEncoder(body).Encode(baggageclaim.VolumeRequest{
				Strategy: encStrategy(map[string]string{
					"type": "empty",
				}),
				Properties: baggageclaim.VolumeProperties{
					"property-name": "property-val",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			recorder := httptest.NewRecorder()
			request, _ := http.NewRequest("POST", "/volumes", body)
			handler.ServeHTTP(recorder, request)
			Expect(recorder.Code).To(Equal(201))

			recorder = httptest.NewRecorder()
			request, _ = http.NewRequest("GET", "/volumes?property-name=property-val", nil)
			handler.ServeHTTP(recorder, request)
			Expect(recorder.Code).To(Equal(200))

			var volumes volume.Volumes
			err = json.NewDecoder(recorder.Body).Decode(&volumes)
			Expect(err).NotTo(HaveOccurred())
			Expect(volumes).To(HaveLen(1))

			err = json.NewEncoder(body).Encode(baggageclaim.PropertyRequest{
				Value: "other-val",
			})
			Expect(err).NotTo(HaveOccurred())

			recorder = httptest.NewRecorder()
			request, _ = http.NewRequest("PUT", fmt.Sprintf("/volumes/%s/properties/property-name", volumes[0].Handle), body)
			handler.ServeHTTP(recorder, request)
			Expect(recorder.Code).To(Equal(http.StatusNoContent))
			Expect(recorder.Body.String()).To(BeEmpty())

			recorder = httptest.NewRecorder()
			request, _ = http.NewRequest("GET", "/volumes?property-name=other-val", nil)
			handler.ServeHTTP(recorder, request)
			Expect(recorder.Code).To(Equal(200))

			err = json.NewDecoder(recorder.Body).Decode(&volumes)
			Expect(err).NotTo(HaveOccurred())

			Expect(volumes).To(HaveLen(1))
		})

		It("can have its ttl updated", func() {
			body := &bytes.Buffer{}

			err := json.NewEncoder(body).Encode(baggageclaim.VolumeRequest{
				Strategy: encStrategy(map[string]string{
					"type": "empty",
				}),
				TTLInSeconds: 1,
			})
			Expect(err).NotTo(HaveOccurred())

			recorder := httptest.NewRecorder()
			request, _ := http.NewRequest("POST", "/volumes", body)
			handler.ServeHTTP(recorder, request)
			Expect(recorder.Code).To(Equal(201))

			var firstVolume volume.Volume
			err = json.NewDecoder(recorder.Body).Decode(&firstVolume)
			Expect(err).NotTo(HaveOccurred())
			Expect(firstVolume.TTL).To(Equal(volume.TTL(1)))

			err = json.NewEncoder(body).Encode(baggageclaim.TTLRequest{
				Value: 2,
			})
			Expect(err).NotTo(HaveOccurred())

			recorder = httptest.NewRecorder()
			request, _ = http.NewRequest("PUT", fmt.Sprintf("/volumes/%s/ttl", firstVolume.Handle), body)
			handler.ServeHTTP(recorder, request)
			Expect(recorder.Code).To(Equal(http.StatusNoContent))
			Expect(recorder.Body.String()).To(BeEmpty())

			recorder = httptest.NewRecorder()
			request, _ = http.NewRequest("GET", "/volumes", body)
			handler.ServeHTTP(recorder, request)
			Expect(recorder.Code).To(Equal(200))

			var volumes volume.Volumes
			err = json.NewDecoder(recorder.Body).Decode(&volumes)
			Expect(err).NotTo(HaveOccurred())
			Expect(volumes).To(HaveLen(1))
			Expect(volumes[0].TTL).To(Equal(volume.TTL(2)))
			Expect(volumes[0].ExpiresAt).NotTo(Equal(firstVolume.ExpiresAt))
		})
	})

	Describe("creating a volume", func() {
		var (
			recorder *httptest.ResponseRecorder
			body     io.ReadWriter
		)

		JustBeforeEach(func() {
			recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("POST", "/volumes", body)

			handler.ServeHTTP(recorder, request)
		})

		Context("when there are properties given", func() {
			var properties baggageclaim.VolumeProperties

			Context("with valid properties", func() {
				BeforeEach(func() {
					properties = baggageclaim.VolumeProperties{
						"property-name": "property-value",
					}

					body = &bytes.Buffer{}
					json.NewEncoder(body).Encode(baggageclaim.VolumeRequest{
						Strategy: encStrategy(map[string]string{
							"type": "empty",
						}),
						Properties: properties,
					})
				})

				It("creates the properties file", func() {
					var response volume.Volume
					err := json.NewDecoder(recorder.Body).Decode(&response)
					Expect(err).NotTo(HaveOccurred())

					propertiesPath := filepath.Join(volumeDir, "live", response.Handle, "properties.json")
					Expect(propertiesPath).To(BeAnExistingFile())

					propertiesContents, err := ioutil.ReadFile(propertiesPath)
					Expect(err).NotTo(HaveOccurred())

					var storedProperties baggageclaim.VolumeProperties
					err = json.Unmarshal(propertiesContents, &storedProperties)
					Expect(err).NotTo(HaveOccurred())

					Expect(storedProperties).To(Equal(properties))
				})

				It("returns the properties in the response", func() {
					Expect(recorder.Body).To(ContainSubstring(`"property-name":"property-value"`))
				})
			})
		})

		Context("when there are no properties given", func() {
			BeforeEach(func() {
				body = &bytes.Buffer{}
				json.NewEncoder(body).Encode(baggageclaim.VolumeRequest{
					Strategy: encStrategy(map[string]string{
						"type": "empty",
					}),
				})
			})

			Context("when a new directory can be created", func() {
				It("writes a nice JSON response", func() {
					Expect(recorder.Body).To(ContainSubstring(`"path":`))
					Expect(recorder.Body).To(ContainSubstring(`"handle":`))
				})
			})

			Context("when invalid JSON is submitted", func() {
				BeforeEach(func() {
					body = bytes.NewBufferString("{{{{{{")
				})

				It("returns a 400 Bad Request response", func() {
					Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				})

				It("writes a nice JSON response", func() {
					Expect(recorder.Body).To(ContainSubstring(`"error":`))
				})

				It("does not create a volume", func() {
					getRecorder := httptest.NewRecorder()
					getReq, _ := http.NewRequest("GET", "/volumes", nil)
					handler.ServeHTTP(getRecorder, getReq)
					Expect(getRecorder.Body).To(MatchJSON("[]"))
				})
			})

			Context("when no strategy is submitted", func() {
				BeforeEach(func() {
					body = bytes.NewBufferString("{}")
				})

				It("returns a 422 Unprocessable Entity response", func() {
					Expect(recorder.Code).To(Equal(422))
				})

				It("writes a nice JSON response", func() {
					Expect(recorder.Body).To(ContainSubstring(`"error":`))
				})

				It("does not create a volume", func() {
					getRecorder := httptest.NewRecorder()
					getReq, _ := http.NewRequest("GET", "/volumes", nil)
					handler.ServeHTTP(getRecorder, getReq)
					Expect(getRecorder.Body).To(MatchJSON("[]"))
				})
			})

			Context("when an unrecognized strategy is submitted", func() {
				BeforeEach(func() {
					body = &bytes.Buffer{}
					json.NewEncoder(body).Encode(baggageclaim.VolumeRequest{
						Strategy: encStrategy(map[string]string{
							"type": "grime",
						}),
					})
				})

				It("returns a 422 Unprocessable Entity response", func() {
					Expect(recorder.Code).To(Equal(422))
				})

				It("writes a nice JSON response", func() {
					Expect(recorder.Body).To(ContainSubstring(`"error":`))
				})

				It("does not create a volume", func() {
					getRecorder := httptest.NewRecorder()
					getReq, _ := http.NewRequest("GET", "/volumes", nil)
					handler.ServeHTTP(getRecorder, getReq)
					Expect(getRecorder.Body).To(MatchJSON("[]"))
				})
			})

			Context("when the strategy is cow but not parent volume is provided", func() {
				BeforeEach(func() {
					body = &bytes.Buffer{}
					json.NewEncoder(body).Encode(baggageclaim.VolumeRequest{
						Strategy: encStrategy(map[string]string{
							"type": "cow",
						}),
					})
				})

				It("returns a 422 Unprocessable Entity response", func() {
					Expect(recorder.Code).To(Equal(422))
				})

				It("writes a nice JSON response", func() {
					Expect(recorder.Body).To(ContainSubstring(`"error":`))
				})

				It("does not create a volume", func() {
					getRecorder := httptest.NewRecorder()
					getReq, _ := http.NewRequest("GET", "/volumes", nil)
					handler.ServeHTTP(getRecorder, getReq)
					Expect(getRecorder.Body).To(MatchJSON("[]"))
				})
			})

			Context("when the strategy is cow and the parent volume does not exist", func() {
				BeforeEach(func() {
					body = &bytes.Buffer{}
					json.NewEncoder(body).Encode(baggageclaim.VolumeRequest{
						Strategy: encStrategy(map[string]string{
							"type":   "cow",
							"volume": "#pain",
						}),
					})
				})

				It("returns a 422 Unprocessable Entity response", func() {
					Expect(recorder.Code).To(Equal(422))
				})

				It("writes a nice JSON response", func() {
					Expect(recorder.Body).To(ContainSubstring(`"error":`))
				})

				It("does not create a volume", func() {
					getRecorder := httptest.NewRecorder()
					getReq, _ := http.NewRequest("GET", "/volumes", nil)
					handler.ServeHTTP(getRecorder, getReq)
					Expect(getRecorder.Body).To(MatchJSON("[]"))
				})
			})
		})
	})
})

func encStrategy(strategy map[string]string) *json.RawMessage {
	bytes, err := json.Marshal(strategy)
	Expect(err).NotTo(HaveOccurred())

	msg := json.RawMessage(bytes)

	return &msg
}
