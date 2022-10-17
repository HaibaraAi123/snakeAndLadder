package dao

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"time"
)

const (
	mongoUrl = ""
	mongoDb  = "chenfeng123"
)

func NewMongoClient() (*mongo.Database, error) {
	/*
		mongodb://user:pwd@localhost:27017/?authSource=
	*/
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoUrl))
	if err != nil {
		return nil, err
	}
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}
	return client.Database(mongoDb), nil
}

type Game struct {
	Id         primitive.ObjectID
	CreateTime int64
	ModifyTime int64
	UserIds    []string
	N          int32
	MaxStep    int32

	SnakeTails    []int32
	SnakeHeads    []int32
	LadderTails   []int32
	LadderHeads   []int32
	NextUserIndex int32

	Round int32
}

func (a *Game) Index() []mongo.IndexModel {
	return []mongo.IndexModel{}
}
func (a *Game) Collection() string {
	return "game"
}
func (a *Game) CreateIndexes(db *mongo.Database) error {
	opts := options.CreateIndexes()
	names, err := db.Collection(a.Collection()).Indexes().CreateMany(context.TODO(), a.Index(), opts)
	if err != nil {
		return err
	}
	log.Printf("createIndexes, db:%v, collection:%v, names:%v", db.Name(), a.Collection(), names)
	return nil
}

func (a *Game) Insert(db *mongo.Database, _ context.Context) error {
	return nil
}

func (a *Game) FindOne(db *mongo.Database, _ context.Context, filter bson.D) error {
	return nil
}

func (a *Game) UpdateOne(db *mongo.Database, _ context.Context, filter, data bson.D) error {
	return nil
}

/*

 */

type GameRecord struct {
	Id         primitive.ObjectID
	GameId     primitive.ObjectID
	CreateTime int64
	ModifyTime int64
	Positions  []int
	Steps      []int
}

func (a *GameRecord) Index() []mongo.IndexModel {
	return []mongo.IndexModel{}
}
func (a *GameRecord) Collection() string {
	return "gameRecord"
}
func (a *GameRecord) CreateIndexes(db *mongo.Database) error {
	opts := options.CreateIndexes()
	names, err := db.Collection(a.Collection()).Indexes().CreateMany(context.TODO(), a.Index(), opts)
	if err != nil {
		return err
	}
	log.Printf("createIndexes, db:%v, collection:%v, names:%v", db.Name(), a.Collection(), names)
	return nil
}

func (a *GameRecord) Insert(db *mongo.Database, _ context.Context) error {
	return nil
}

func (a *GameRecord) FindOne(db *mongo.Database, _ context.Context, filter bson.D) error {
	return nil
}
func (a *GameRecord) UpdateOne(db *mongo.Database, _ context.Context, filter, data bson.D) error {
	return nil
}
