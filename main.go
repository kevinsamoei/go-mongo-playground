package main

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	Name string
	Age  int
}

func (t *User) String() string {
	return fmt.Sprintf("%v, %v", t.Name, t.Age)
}

func main() {
	ctx := context.Background()
	opts := options.Client().ApplyURI("mongodb://localhost:27017")

	// connect mongodb
	mClient, err := mongo.Connect(ctx, opts)
	if err != nil {
		log.Fatal(err)
	}

	// ensure we are connected
	err = mClient.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("could not connect to mongo: %v", err)
	}
	defer mClient.Disconnect(ctx)
	fmt.Println("Connected to the db")

	// create collection users
	clctn := mClient.Database("test").Collection("users")

	// create a unique index on the "name" field
	i, err := clctn.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.M{"name": -1}, Options: options.Index().SetUnique(true)})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Index created: ", i)

	// Insert a record
	res, err := clctn.InsertOne(ctx, User{"Kevin", 20})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Insert result: ", res)

	// find by name
	var result User
	err = clctn.FindOne(ctx, bson.D{{"name", "Kevin"}}).Decode(&result)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("----------------------------")
	fmt.Println("Name: ", result.Name)
	fmt.Println("Age: ", result.Age)

	fmt.Println("===============Using cursor============")

	cur, err := clctn.Find(ctx, bson.D{{}})
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(ctx)

	var users []*User
	for cur.Next(ctx) {
		var t User
		err := cur.Decode(&t)
		if err != nil {
			log.Fatal(err)
		}
		users = append(users, &t)
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("All items: ", users)

	// test unique index
	_, err = clctn.InsertOne(ctx, User{"Kevin", 60})
	if err != nil {
		fmt.Println("Failed to insert: ", err) // expected
	}

	// truncate after each run
	clctn.Drop(ctx)
	// delete the test db
	mClient.Database("test").Drop(ctx)
}
