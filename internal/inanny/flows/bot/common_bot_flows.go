package botflows

import (
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/holeyko/innany-tgbot/internal/inanny/flows"
)

type BotFlowContext struct {
	bot *tgbot.BotAPI
}

type BotFlowHandler struct {
	flowDao flows.FlowDao
}

func NewBotFlowHandler() *BotFlowHandler {
	return &BotFlowHandler{
		flowDao: flows.NewFlowDao(),
	}
}

var botFlows = [...]flows.Flow[BotFlowContext, any]{}

func (handler *BotFlowHandler) HandleFlowStep(
	bot *tgbot.BotAPI,
	tgId int64,
	flowName flows.FlowName,
	nextStep flows.FlowStep,
) (err error) {

}

func findSuitableFlow[DATA any](flowName flows.FlowName) flows.Flow[BotFlowContext, DATA] {
	for _, flow := range botFlows {
		if flow.SupportedFlow() == flowName {
			return flow
		}
	}

	return nil
}
