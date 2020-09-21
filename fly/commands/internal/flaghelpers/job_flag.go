package flaghelpers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jessevdk/go-flags"

	"github.com/concourse/concourse/atc"
	"github.com/concourse/concourse/fly/rc"
)

type JobFlag struct {
	PipelineRef atc.PipelineRef
	JobName     string
}

func (flag *JobFlag) UnmarshalFlag(value string) error {
	flag.PipelineRef = atc.PipelineRef{}

	jobNameIdx := strings.LastIndex(value, "/")
	if jobNameIdx == -1 {
		return errors.New("argument format should be <pipeline>/<job>")
	}

	flag.JobName = value[jobNameIdx+1:]
	if warnings := atc.ValidateIdentifier(flag.JobName, "job"); warnings != nil {
		return errors.New("argument format should be <pipeline>/<key:value>/<job>")
	}

	vs := strings.SplitN(value[:jobNameIdx], "/", 2)
	flag.PipelineRef.Name = vs[0]
	if len(vs) == 2 {
		instanceVars := atc.InstanceVars{}
		for _, instanceVar := range strings.Split(vs[1], ",") {
			kv := strings.SplitN(strings.TrimSpace(instanceVar), ":", 2)
			if len(kv) == 2 {
				instanceVars[kv[0]] = kv[1]
			} else {
				return errors.New("argument format should be <pipeline>/<key:value>/<job>")
			}
		}
		flag.PipelineRef.InstanceVars = instanceVars
	}

	return nil
}

func (flag *JobFlag) Complete(match string) []flags.Completion {
	fly := parseFlags()

	target, err := rc.LoadTarget(fly.Target, false)
	if err != nil {
		return []flags.Completion{}
	}

	err = target.Validate()
	if err != nil {
		return []flags.Completion{}
	}

	team := target.Team()
	var comps []flags.Completion
	vs := strings.SplitN(match, "/", 3)

	if len(vs) == 1 {
		pipelines, err := team.ListPipelines()
		if err != nil {
			return comps
		}

		for _, pipeline := range pipelines {
			if strings.HasPrefix(pipeline.Name, vs[0]) {
				comps = append(comps, flags.Completion{Item: pipeline.Name + "/"})
			}
		}
	} else if len(vs) == 2 {
		pipelines, err := team.ListPipelines()
		if err != nil {
			return comps
		}
		instanceVars := atc.InstanceVars{}
		for _, instanceVar := range strings.Split(vs[1], ",") {
			kv := strings.SplitN(strings.TrimSpace(instanceVar), ":", 2)
			if len(kv) == 2 {
				instanceVars[kv[0]] = kv[1]
			}
		}

		pipelineRef := atc.PipelineRef{Name: vs[0], InstanceVars: instanceVars}
		for _, pipeline := range pipelines {
			if strings.HasPrefix(pipeline.Ref().String(), pipelineRef.String()) {
				comps = append(comps, flags.Completion{Item: pipeline.Ref().String() + "/"})
			}
		}

		if comps == nil {
			pipelineRef := atc.PipelineRef{Name: vs[0]}
			jobs, err := team.ListJobs(pipelineRef)
			if err != nil {
				return comps
			}
			for _, job := range jobs {
				if strings.HasPrefix(job.Name, vs[1]) {
					comps = append(comps, flags.Completion{Item: fmt.Sprintf("%s/%s", pipelineRef.String(), job.Name)})
				}
			}
		}
	} else if len(vs) == 3 {
		instanceVars := atc.InstanceVars{}
		for _, instanceVar := range strings.Split(vs[1], ",") {
			kv := strings.SplitN(strings.TrimSpace(instanceVar), ":", 2)
			if len(kv) == 2 {
				instanceVars[kv[0]] = kv[1]
			}
		}
		pipelineRef := atc.PipelineRef{Name: vs[0], InstanceVars: instanceVars}
		jobs, err := team.ListJobs(pipelineRef)
		if err != nil {
			return comps
		}
		for _, job := range jobs {
			if strings.HasPrefix(job.Name, vs[2]) {
				comps = append(comps, flags.Completion{Item: fmt.Sprintf("%s/%s", pipelineRef.String(), job.Name)})
			}
		}
	}

	return comps
}
