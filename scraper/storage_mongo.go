package scraper

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Mongo implements colly storage
type Mongo struct {
	Client     *mongo.Client
	Database   string
	Collection string
	db         *mongo.Database
	visited    *mongo.Collection
	cookies    *mongo.Collection
	queue      *mongo.Collection
}

func (m *Mongo) Init() error {
	if m.Client == nil {
		return fmt.Errorf("Mongo storage client not set")
	}

	if m.Database == "" || m.Collection == "" {
		return fmt.Errorf("Mongo database or collection not set")
	}

	m.db = m.Client.Database(m.Database)
	m.visited = m.db.Collection(fmt.Sprintf("%s_%s", m.Collection, "visited"))
	m.cookies = m.db.Collection(fmt.Sprintf("%s_%s", m.Collection, "cookies"))
	m.queue = m.db.Collection(fmt.Sprintf("%s_%s", m.Collection, "queue"))

	return nil
}

func (m *Mongo) Visited(requestID uint64) error {
	_, err := m.visited.InsertOne(context.TODO(), bson.D{
		{"requestID", strconv.FormatUint(requestID, 10)},
		{"visited", true},
	})
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (m *Mongo) IsVisited(requestID uint64) (bool, error) {
	var result bson.D
	err := m.visited.FindOne(context.TODO(), bson.D{
		{"requestID", strconv.FormatUint(requestID, 10)},
	}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}

		log.Println(err)
		return false, err
	}

	return true, nil
}

func (m *Mongo) Cookies(u *url.URL) string {
	var result struct {
		Host    string
		Cookies string
	}

	err := m.cookies.FindOne(context.TODO(), bson.D{
		{"host", u.Host},
	}).Decode(&result)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			log.Println(err)
		}
		return ""
	}

	return result.Cookies
}

func (m *Mongo) SetCookies(u *url.URL, cookies string) {
	_, err := m.cookies.InsertOne(context.TODO(), bson.D{
		{"host", u.Host},
		{"cookies", cookies},
	})
	if err != nil {
		log.Println(err)
	}
}

func (m *Mongo) Clear() error {
	err := m.Client.Database(m.Database).Collection(fmt.Sprintf("%s_%s", m.Collection, "visited")).Drop(context.TODO())
	if err != nil {
		log.Println(err)
		return err
	}

	err = m.Client.Database(m.Database).Collection(fmt.Sprintf("%s_%s", m.Collection, "cookies")).Drop(context.TODO())
	if err != nil {
		log.Println(err)
		return err
	}

	err = m.Client.Database(m.Database).Collection(fmt.Sprintf("%s_%s", m.Collection, "queue")).Drop(context.TODO())
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (m *Mongo) AddRequest(req []byte) error {
	_, err := m.queue.InsertOne(context.TODO(), bson.D{
		{"request", req},
	})
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (m *Mongo) GetRequest() ([]byte, error) {
	type result struct {
		ID      primitive.ObjectID `bson:"_id"`
		Request []byte             `bson:"request"`
	}

	// First find the first inserted, FIFO
	findOptions := options.Find()
	findOptions.SetSort(bson.D{
		{"_id", 1},
	})
	findOptions.SetLimit(1)

	cur, err := m.queue.Find(context.TODO(), bson.D{}, findOptions)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer cur.Close(context.TODO())

	var results []result
	for cur.Next(context.TODO()) {
		var res result
		err := cur.Decode(&res)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		results = append(results, res)
	}
	if err := cur.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("Empty queue")
	}

	id := results[0].ID
	req := results[0].Request

	// Then delete it
	_, err = m.queue.DeleteOne(context.TODO(), bson.D{
		{"_id", id},
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return req, nil
}

func (m *Mongo) QueueSize() (int, error) {
	count, err := m.queue.CountDocuments(context.TODO(), bson.D{})
	if err != nil {
		log.Println(err)
		return 0, err
	}

	return int(count), nil
}
