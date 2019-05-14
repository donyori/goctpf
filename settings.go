package goctpf

import (
	"encoding/json"
	"io/ioutil"
	"time"
)

type WorkerSettings struct {
	Number         uint32        `json:"number"`
	SendErrTimeout time.Duration `json:"send_err_timeout,omitempty"`
}

func NewWorkerSettings() *WorkerSettings {
	return new(WorkerSettings)
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
