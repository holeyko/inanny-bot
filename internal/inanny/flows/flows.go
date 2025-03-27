package flows

type FlowName string

type FlowStep string

// type Flow interface {
// 	SupportedFlow() FlowName
// 	DataType[CONTEXT any, DATA any]() DATA
// 	HandleNextStep(state FlowState[DATA], context CONTEXT, nextStep FlowStep) (err error)
// }

// func foo[CONTEXT any, DATA any]() {

// }

// type FlowState[DATA any] struct {
// 	tgId int64
// 	flow FlowName
// 	step FlowStep
// 	data DATA
// }

const (
	Initial     FlowStep = "INITIAL"
	Termination FlowStep = "TERMINATION"
)
