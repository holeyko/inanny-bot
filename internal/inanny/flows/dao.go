package flows

// import (
// 	ctx "context"
// 	"encoding/json"
// 	"log"

// 	"github.com/holeyko/innany-tgbot/internal/inanny/db"
// 	"github.com/jackc/pgx/v5/pgxpool"
// )

const (
	INSERT_FLOW string = `
		INSERT INTO flows(tg_id, "name", step, "data")
		VALUES ($1, $2, $3, $4);
	`

	SELECT_FLOW string = `
		SELECT tg_id, "name", step, "data" FROM flows
		WHERE tg_id = $1
			AND "name" = $2;
	`
)

// type FlowDao interface {
// 	SaveFlow(state *FlowState[any]) (err error)
// 	FindFlow(tgId int64, flowName FlowName) *FlowState[any]
// }

// type DbFlowDao struct{}

// func NewFlowDao() FlowDao {
// 	return &DbFlowDao{}
// }

// func (dao *DbFlowDao) SaveFlow(state *FlowState[any]) (err error) {
// 	data, err := json.Marshal(state.data)

// 	if err != nil {
// 		log.Println("Can't serializea context for", state.flow, "flow")
// 		return err
// 	}

// 	err = db.Execute(func(pool *pgxpool.Pool) error {
// 		_, err = pool.Exec(
// 			ctx.Background(),
// 			INSERT_FLOW,
// 			state.tgId,
// 			state.flow,
// 			state.step,
// 			data,
// 		)

// 		return err
// 	})

// 	return
// }

// // func (dao *DbFlowDao) FindFlow(tgId int64, flowName FlowName) *FlowState[any] {
// // 	state := db.Execute(func(pool *pgxpool.Pool) *FlowState {
// // 		state := FlowState[any]{}

// // 		err := pool.QueryRow(
// // 			ctx.Background(),
// // 			SELECT_FLOW,
// // 			tgId,
// // 			flowName,
// // 		).Scan(
// // 			&state.tgId,
// // 			&state.flow,
// // 			&state.step,
// // 			&state.data,
// // 		)

// // 		if err != nil {
// // 			log.Println("err:", err)
// // 			return nil
// // 		}

// // 		return &state
// // 	})

// // 	return state
// // }
