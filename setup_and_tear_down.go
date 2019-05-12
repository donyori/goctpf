package goctpf

type SetupAndTearDown struct {
	Setup, TearDown func(workerNo int)
}
