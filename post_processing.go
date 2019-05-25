package goctpf

type DoneProcessor interface {
	Done()
}

type DiscardProcessor interface {
	Discard()
}
