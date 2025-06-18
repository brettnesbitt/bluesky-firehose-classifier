package repositories

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"stockseer.ai/blueksy-firehose/internal/models"
)

// MongoRepository implements the Repository interface using MongoDB.
type MongoRepository struct {
	collection *mongo.Collection
}

// NewMongoRepository creates a new instance of MongoRepository.
func NewMongoRepository(client *mongo.Client, dbName, collectionName string) *MongoRepository {
	collection := client.Database(dbName).Collection(collectionName)
	return &MongoRepository{collection: collection}
}

// Insert inserts a document into the MongoDB collection.
func (r *MongoRepository) Insert(data interface{}) error {
	switch v := data.(type) {
	case *models.ProtoMessage:
		// handle ProtoMessage
		_data, _ := v.WithDateTime()
		_, err := r.collection.InsertOne(context.Background(), _data)
		return err
	case models.CategoryMetrics:
		// handle CategoryMetrics
		_, err := r.collection.InsertOne(context.Background(), data)
		return err
	default:
		return fmt.Errorf("unsupported data type: %T", data)
	}
}

// FindAll retrieves all documents from the collection.
func (r *MongoRepository) FindAll() ([]interface{}, error) {
	cur, err := r.collection.Find(context.Background(), bson.D{})
	if err != nil {
		return nil, err
	}
	var results []interface{}
	if err = cur.All(context.Background(), &results); err != nil {
		return nil, err
	}
	return results, nil
}

// FindByID retrieves a document by its ID.
func (r *MongoRepository) FindByID(id string) (interface{}, error) {
	var result interface{}
	err := r.collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&result)
	return result, err
}

// Update updates a document in the MongoDB collection.
func (r *MongoRepository) Update(id string, data interface{}) error {
	_, err := r.collection.UpdateOne(context.Background(), bson.M{"_id": id}, bson.M{"$set": data})
	return err
}

// DeleteByID deletes a document by its ID.
func (r *MongoRepository) Delete(id string) error {
	_, err := r.collection.DeleteOne(context.Background(), bson.M{"_id": id})
	return err
}
