package goctpf

import "strings"

type Source int8
type Purpose int8

const (
	FromApp Source = iota + 1
	FromWorkers
	FromMe
	FromOthers
)

const (
	ForWorkers Purpose = iota + 1
	ForMe
	ForOthers
)

var sourceStrings = [...]string{
	"Unknown",
	"FromApp",
	"FromWorkers",
	"FromMe",
	"FromOthers",
}

var purposeStrings = [...]string{
	"Unknown",
	"ForWorkers",
	"ForMe",
	"ForOthers",
}

func ParseSource(s string) Source {
	for i := range sourceStrings {
		if strings.EqualFold(s, sourceStrings[i]) {
			return Source(i)
		}
	}
	return 0 // Stands for "Unknown".
}

func (s Source) String() string {
	if s < FromApp || s > FromOthers {
		return sourceStrings[0]
	}
	return sourceStrings[s]
}

func (s Source) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *Source) UnmarshalText(text []byte) error {
	*s = ParseSource(string(text))
	return nil
}

func ParsePurpose(s string) Purpose {
	for i := range purposeStrings {
		if strings.EqualFold(s, purposeStrings[i]) {
			return Purpose(i)
		}
	}
	return 0 // Stands for "Unknown".
}

func (p Purpose) String() string {
	if p < ForWorkers || p > ForOthers {
		return purposeStrings[0]
	}
	return purposeStrings[p]
}

func (p Purpose) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

func (p *Purpose) UnmarshalText(text []byte) error {
	*p = ParsePurpose(string(text))
	return nil
}
