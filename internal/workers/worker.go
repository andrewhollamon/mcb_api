package workers

type Result int

const (
	ResultSuccess Result = iota
	ResultFailure
)

// Legacy enum for backward compatibility
var ResultEnum = struct {
	Success Result
	Failure Result
}{
	Success: ResultSuccess,
	Failure: ResultFailure,
}

type QueueConsumerResult struct {
	Result       Result
	NumProcessed int
}
