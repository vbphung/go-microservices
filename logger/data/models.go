package data

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func New(mgClient *mongo.Client) Models {
	client = mgClient

	return Models{
		LogEntry: LogEntry{},
	}
}

type Models struct {
	LogEntry LogEntry
}

type LogEntry struct {
	ID        string    `bson:"_id,omitempty" json:"id,omitempty"`
	Name      string    `bson:"name" json:"name"`
	Data      string    `bson:"data" json:"data"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

func Insert(entr LogEntry) error {
	clt := client.Database("logs").Collection("logs")

	if _, err := clt.InsertOne(context.TODO(), LogEntry{
		Name:      entr.Name,
		Data:      entr.Data,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}); err != nil {
		log.Println("Error log insert: ", err)
		return err
	}

	return nil
}

func All() ([]*LogEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	clt := client.Database("logs").Collection("logs")

	opts := options.Find()
	opts.SetSort(bson.D{{Key: "created_at", Value: -1}})

	cs, err := clt.Find(context.TODO(), bson.D{}, opts)
	if err != nil {
		log.Println("Error logs getAll: ", err)
		return nil, err

	}

	defer cs.Close(ctx)

	var logs []*LogEntry
	for cs.Next(ctx) {
		var logItem LogEntry

		if err = cs.Decode(&logItem); err != nil {
			log.Println("Error log decode: ", err)
			return nil, err
		}

		logs = append(logs, &logItem)

	}

	return logs, nil
}

func GetOne(id string) (*LogEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	clt := client.Database("logs").Collection("logs")

	docID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Println("Error log getOne: ", err)
		return nil, err
	}

	var logEntr LogEntry

	err = clt.FindOne(ctx, bson.M{"_id": docID}).Decode(&logEntr)
	if err != nil {
		log.Println("Error log decode: ", err)
		return nil, err
	}

	return &logEntr, nil
}

func DropClt() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	clt := client.Database("logs").Collection("logs")

	if err := clt.Drop(ctx); err != nil {
		log.Println("Error logs dropClt: ", err)
		return err
	}

	return nil
}

func Update(logEntr LogEntry) (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	clt := client.Database("logs").Collection("logs")

	docID, err := primitive.ObjectIDFromHex(logEntr.ID)
	if err != nil {
		log.Println("Error log getDocID: ", err)
		return nil, err
	}

	res, err := clt.UpdateOne(ctx, bson.M{"_id": docID}, bson.D{{
		Key: "$set", Value: bson.D{
			{Key: "name", Value: logEntr.Name},
			{Key: "data", Value: logEntr.Data},
			{Key: "updated_at", Value: time.Now()},
		},
	}})
	if err != nil {
		log.Println("Error log update: ", err)
		return nil, err
	}

	return res, nil
}
