package flows

type FlowStep string

type Flow[CONTEXT any] interface {
	HandleStep(curStep FlowStep, context CONTEXT) (nextStep FlowStep, err error)
}
