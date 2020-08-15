package flaghelpers_test

import (
	. "github.com/concourse/concourse/atc"
	. "github.com/concourse/concourse/fly/commands/internal/flaghelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("JobFlag", func() {
	var (
		flag *JobFlag
	)

	BeforeEach(func() {
		flag = &JobFlag{}
	})

	Describe("UnmarshalFlag", func() {

		Context("when there is only a pipeline specified", func() {
			It("displays an error message", func() {
				err := flag.UnmarshalFlag("pipeline")
				Expect(err).To(MatchError("argument format should be <pipeline>/<job>"))
			})
		})

		Context("when the job name is not specified", func() {
			It("displays an error message", func() {
				err := flag.UnmarshalFlag("some-pipeline/")
				Expect(err).To(MatchError("argument format should be <pipeline>/<job>"))
			})
		})

		Context("when a pipeline instance is specified", func() {

			Context("when a pipeline ref has a single instance var", func() {
				It("unmarshal the flag successfully", func() {
					err := flag.UnmarshalFlag("some-pipeline/branch:master/job-name")
					Expect(err).ToNot(HaveOccurred())
					Expect(flag.JobName).To(Equal("job-name"))
					Expect(flag.PipelineRef).To(Equal(PipelineRef{
						Name:         "some-pipeline",
						InstanceVars: InstanceVars{"branch": "master"},
					}))
				})
			})

			Context("when a pipeline ref has a multiple instance vars", func() {
				It("unmarshal the flag successfully", func() {
					err := flag.UnmarshalFlag("some-pipeline/branch:master,ref:some-ref/job-name")
					Expect(err).ToNot(HaveOccurred())
					Expect(flag.JobName).To(Equal("job-name"))
					Expect(flag.PipelineRef).To(Equal(PipelineRef{
						Name:         "some-pipeline",
						InstanceVars: InstanceVars{"branch": "master", "ref": "some-ref"},
					}))
				})
			})

			Context("when a pipeline ref has '/' character in an instance vars", func() {
				It("unmarshal the flag successfully", func() {
					err := flag.UnmarshalFlag("some-pipeline/branch:feature/bar,ref:some/ref/job-name")
					Expect(err).ToNot(HaveOccurred())
					Expect(flag.JobName).To(Equal("job-name"))
					Expect(flag.PipelineRef).To(Equal(PipelineRef{
						Name:         "some-pipeline",
						InstanceVars: InstanceVars{"branch": "feature/bar", "ref": "some/ref"},
					}))
				})
			})

			Context("when an instance vars is complex", func() {
				It("unmarshal the pipeline ref with instance vars correctly", func() {
					err := flag.UnmarshalFlag("some-pipeline/foo.bar.baz:1,foo.bar.qux:2,bar.0:1,bar.1:\"2\"/job-name")

					Expect(err).ToNot(HaveOccurred())
					Expect(flag.JobName).To(Equal("job-name"))
					Expect(flag.PipelineRef).To(Equal(PipelineRef{
						Name: "some-pipeline",
						InstanceVars: InstanceVars{
							"bar": []interface{}{1, "2"},
							"foo": map[string]interface{}{
								"bar": map[string]interface{}{
									"baz": 1,
									"qux": 2,
								},
							},
						},
					}))
				})
			})

			Context("when the instance var is malformed", func() {
				It("displays an error message", func() {
					err := flag.UnmarshalFlag("some-pipeline/branch=master/job-name")
					Expect(err).To(MatchError("argument format should be <pipeline>/<key:value>/<job>"))
				})
			})

		})
	})
})
