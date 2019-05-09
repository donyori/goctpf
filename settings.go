package goctpf

import (
	"encoding/json"
	"io/ioutil"
	"runtime"
	"time"
)

type WorkerSettings struct {
	Number         uint32        `json:"number"`
	SendErrTimeout time.Duration `json:"send_err_timeout,omitempty"`
}

func NewWorkerSettings() *WorkerSettings {
	maxProcs := runtime.GOMAXPROCS(0)
	var n uint32
	if maxProcs >= 0 {
		n = uint32(maxProcs)
	} else {
		n = 0
	}
	return &WorkerSettings{
		Number:         n,
		SendErrTimeout: 0,
	}
}

func LoadWorkerSettings(filename string) (settings *WorkerSettings, err error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	settings = NewWorkerSettings()
	err = json.Unmarshal(data, settings)
	if err != nil {
		return nil, err
	}
	return settings, nil
}
