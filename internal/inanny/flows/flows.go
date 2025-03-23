package flows

type FLowName string

type FlowStep string

type Flow[CONTEXT any, DATA any] interface {
	HandleStep(curStep FlowStep, context CONTEXT, DATA data) (nextStep FlowStep, err error)
}

type FlowState[DATA any] struct {
	flowName FLowName
	step     FlowStep
	data 	 DATA
}

const (
	AddReqularPoll FLowName = "AddReqularPoll"
)

type CommonStep struct {
	Initial FlowStep = "Initial"
}
