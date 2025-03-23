package db

import (
	cxt "context"
	"encoding/json"
	"log"

	"github.com/holeyko/innany-tgbot/internal/inanny/flows"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	INSERT_FLOW string = `
		INSERT INTO flows(tg_id, "name", step, context)
		VALUES ($1, $2, $3, $4);
	`
)

func SaveFlow[CONTEXT any](tgId int64, flowName flows.FLowName, context CONTEXT) error {
	jsonContext, err := json.Marshal(context)

	if err != nil {
		log.Println("Can't serializea context for", flowName, "flow")
		return err
	}

	err = Execute(func(pool *pgxpool.Pool) error {
		_, err = pool.Exec(
			cxt.Background(),
			INSERT_FLOW,
			tgId,
			flowName,
			flows.Initial,
			jsonContext,
		)

		return err
	})

	return err
}

func RetrieveFlow[CONTEXT any](tgId int64, flowName flows.FLowName) (flows.FlowState[CONTEXT], error) {

}
