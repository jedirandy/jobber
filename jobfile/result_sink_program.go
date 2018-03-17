package jobfile

import (
	"encoding/json"
	"os/exec"

	"github.com/dshearer/jobber/common"
)

const _PROGRAM_RESULT_SINK_NAME = "program"

type ProgramResultSink struct {
	Path string `yaml:"path"`
}

func (self ProgramResultSink) String() string {
	return _PROGRAM_RESULT_SINK_NAME
}

func (self ProgramResultSink) Equals(other ResultSink) bool {
	otherProgResultSink, ok := other.(ProgramResultSink)
	if !ok {
		return false
	}
	if otherProgResultSink.Path != self.Path {
		return false
	}
	return true
}

func (self ProgramResultSink) Validate() error {
	if len(self.Path) == 0 {
		return &common.Error{What: "Program result sink needs 'path' param"}
	}
	return nil
}

func (self ProgramResultSink) Handle(rec RunRec) {
	/*
	 Here we make a JSON document with the data in rec, and then pass it
	 to a user-specified program.
	*/

	var timeFormat string = "Jan _2 15:04:05 2006"

	// make job JSON
	jobJson := map[string]interface{}{
		"name":    rec.Job.Name,
		"command": rec.Job.Cmd,
		"time":    rec.Job.FullTimeSpec.String(),
		"status":  rec.NewStatus.String()}

	// make rec JSON
	recJson := map[string]interface{}{
		"job":       jobJson,
		"user":      rec.Job.User,
		"startTime": rec.RunTime.Format(timeFormat),
		"succeeded": rec.Succeeded}
	if rec.Stdout == nil {
		recJson["stdout"] = nil
	} else {
		stdoutStr, stdoutBase64 := SafeBytesToStr(rec.Stdout)
		recJson["stdout"] = stdoutStr
		recJson["stdout_base64"] = stdoutBase64
	}
	if rec.Stderr == nil {
		recJson["stderr"] = nil
	} else {
		stderrStr, stderrBase64 := SafeBytesToStr(rec.Stderr)
		recJson["stderr"] = stderrStr
		recJson["stderr_base64"] = stderrBase64
	}
	recJsonStr, err := json.Marshal(recJson)
	if err != nil {
		common.ErrLogger.Printf("Failed to make RunRec JSON: %v\n", err)
		return
	}

	// call program
	execResult, err2 := common.ExecAndWait(exec.Command(self.Path),
		&recJsonStr)
	if err2 != nil {
		common.ErrLogger.Printf("Failed to call %v: %v\n", self.Path, err2)
	} else if !execResult.Succeeded {
		errMsg, _ := SafeBytesToStr(execResult.Stderr)
		common.ErrLogger.Printf(
			"%v failed: %v\n",
			self.Path,
			errMsg,
		)
	}
}
