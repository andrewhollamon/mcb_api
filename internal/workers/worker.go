package workers

type Result int

var ResultEnum = struct {
	Success Result
	Failure Result
}{
	Success: 0,
	Failure: 1,
}

type QueueConsumerResult struct {
	Result       Result
	NumProcessed int
}
