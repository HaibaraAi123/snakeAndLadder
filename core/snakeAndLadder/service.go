package snakeAndLadder

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"math/rand"
	pb "snakeAndLadder/api"
	"snakeAndLadder/core/dao"
	"snakeAndLadder/edge/ws"
)

type Service struct {
	db *mongo.Database
}

type GameConfig struct {
	N            int
	MaxUserCnt   int
	MaxSnakeCnt  int
	MaxLadderCnt int
	MaxStep      int
}

type Snakes struct {
	Tails    []int
	Heads    []int
	IndexSet map[int]int //head-index
}

type Ladders struct {
	Tails    []int
	Heads    []int
	IndexSet map[int]int //tail-index
}

func NewService() *Service {
	mongoDb, _ := dao.NewMongoClient()
	return &Service{
		db: mongoDb,
	}
}

func (s *Service) NewGame(_ context.Context, req *pb.NewGame_Req) (*pb.NewGame_Resp, error) {
	resp := new(pb.NewGame_Resp)
	resp.Status = new(pb.Status)

	/*
		check param
	*/

	/*
		generate snake and ladder
		首尾不重复
	*/
	meta := &pb.GameMeta{}
	/*
		insert game
	*/
	g := &dao.Game{} // 填充参数
	err := g.Insert(s.db, context.TODO())
	if err != nil {
		resp.Status.Code = -3
		resp.Status.ErrMsg = err.Error()
		return resp, nil
	}
	resp.GameId = g.Id.Hex()
	resp.GameMeta = meta
	return resp, nil
}

func (s *Service) EndGame(_ context.Context, req *pb.EndGame_Req) (*pb.EndGame_Resp, error) {
	return nil, nil
}

func (s *Service) MoveForward(_ context.Context, req *pb.MoveForward_Req) (*pb.MoveForward_Resp, error) {
	resp := new(pb.MoveForward_Resp)
	resp.Status = new(pb.Status)
	/*
		check param
	*/

	/*
		lock in redis or etcd
	*/
	/*
		find game
	*/
	g := &dao.Game{}
	err := g.FindOne(s.db, context.TODO(), bson.D{{Key: "id", Value: req.GameId}})
	if err != nil {
		resp.Status = &pb.Status{Code: -3, ErrMsg: err.Error()}
		return resp, nil
	}
	/*
		必须按顺序 投掷骰子
	*/
	if g.UserIds[int(g.NextUserIndex)%len(g.UserIds)] != req.UserId {
		resp.Status = &pb.Status{Code: -1, ErrMsg: "not in order"}
		return resp, nil
	}
	g.NextUserIndex++
	/*
		find first
	*/
	gr := &dao.GameRecord{}
	err = gr.FindOne(s.db, context.TODO(), bson.D{{Key: "id", Value: req.UserId}, {Key: "gameId", Value: req.GameId}})
	if err != nil && err != mongo.ErrNoDocuments {
		resp.Status = &pb.Status{Code: -3, ErrMsg: err.Error()}
		return resp, nil
	}
	/*
		get random step
	*/
	step := s.randIntN(int(g.MaxStep))
	End := step
	if err == nil {
		End = gr.Positions[len(gr.Positions)-1] + step
	}
	/*
		check if touch snakeHead or ladderTail
	*/
	if End > int(g.N*g.N) {
		End = 2*int(g.N*g.N) - End
	}
	gr.Steps = append(gr.Steps, step)
	gr.Positions = append(gr.Positions, End)

	if End == int(g.N*g.N) {
		// end game
	}

	/*
		if  exist
	*/
	session, sessionErr := s.db.Client().StartSession()
	if sessionErr != nil {

	}
	defer session.EndSession(context.TODO())
	sessionErr = session.StartTransaction()
	if sessionErr != nil {

	}
	sessionCtx := mongo.NewSessionContext(context.TODO(), session)
	if err == mongo.ErrNoDocuments {
		// insert gameRecord
		// fill g
		err = gr.Insert(s.db, sessionCtx)
		if err != nil {
			resp.Status = &pb.Status{Code: -3, ErrMsg: err.Error()}
			return resp, nil
		}
	} else {
		err = gr.UpdateOne(s.db, sessionCtx, nil, nil)
		if err != nil {
			resp.Status = &pb.Status{Code: -3, ErrMsg: err.Error()}
			return resp, nil
		}
	}
	if err = g.UpdateOne(s.db, sessionCtx, nil, nil); err != nil {

	}
	sessionErr = session.CommitTransaction(context.TODO())
	if err != nil {

	}
	/*
		use websocket conn to notify all users
	*/
	_ = ws.WebsocketServer.BroadcastMsg(g.UserIds, &ws.Msg{})
	return resp, nil
}

func (s *Service) FetchReplay(_ context.Context, req *pb.FetchReplay_Req) (*pb.FetchReplay_Resp, error) {
	return nil, nil
}

func (s *Service) randIntN(n int) int {
	//r := rand.NewSource(time.Now().UnixNano())
	return rand.Intn(n)
}
