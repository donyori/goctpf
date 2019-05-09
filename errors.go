package goctpf

import (
	"errors"
	"fmt"
)

type UnknownSourceError struct {
	source interface{}
}

type UnknownPurposeError struct {
	purpose interface{}
}

var ErrNoMoreTask error = errors.New("goctpf: no more task")

func NewUnknownSourceError(source interface{}) error {
	switch source.(type) {
	case Source, string:
		return &UnknownSourceError{source: source}
	default:
		panic(fmt.Errorf(
			"goctpf: type of source should be Source or string, but got %T",
			source))
	}
}

func (use *UnknownSourceError) Error() string {
	switch use.source.(type) {
	case Source:
		return fmt.Sprintf("goctpf: source (%d) is unknown", use.source)
	default:
		return fmt.Sprintf("goctpf: source (%v) is unknown", use.source)
	}
}

func NewUnknownPurposeError(purpose interface{}) error {
	switch purpose.(type) {
	case Purpose, string:
		return &UnknownPurposeError{purpose: purpose}
	default:
		panic(fmt.Errorf(
			"goctpf: type of purpose should be Purpose or string, but got %T",
			purpose))
	}
}

func (upe *UnknownPurposeError) Error() string {
	switch upe.purpose.(type) {
	case Purpose:
		return fmt.Sprintf("goctpf: purpose (%d) is unknown", upe.purpose)
	default:
		return fmt.Sprintf("goctpf: purpose (%v) is unknown", upe.purpose)
	}
}
