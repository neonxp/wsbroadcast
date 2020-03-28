/*
Copyright © 2020 Alexander Kiryukhin <a.kiryukhin@mail.ru>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

type Manager struct {
	collection *mongo.Collection
}

func NewManager(collection *mongo.Collection, indexes map[string]bool) (*Manager, error) {
	if indexes != nil {
		indexModels := make([]mongo.IndexModel, 0, len(indexes))
		for name, isUnique := range indexes {
			indexModels = append(indexModels, mongo.IndexModel{
				Keys:    bsonx.Doc{{name, bsonx.Int32(-1)}},
				Options: (options.Index()).SetBackground(true).SetSparse(true).SetUnique(isUnique),
			})
		}
		_, err := collection.Indexes().CreateMany(context.Background(), indexModels)
		if err != nil {
			return nil, err
		}
	}
	return &Manager{collection: collection}, nil
}

func (m *Manager) Add(s interface{}) (primitive.ObjectID, error) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	r, err := m.collection.InsertOne(ctx, s)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return r.InsertedID.(primitive.ObjectID), nil
}

func (m *Manager) Update(ID primitive.ObjectID, s interface{}) error {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	_, err := m.collection.UpdateOne(ctx, bson.M{"_id": ID}, bson.M{"$set": s})
	return err
}

func (m *Manager) Remove(ID primitive.ObjectID) error {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	_, err := m.collection.DeleteOne(ctx, bson.M{"_id": ID})
	return err
}

func (m *Manager) FindOne(filter bson.M, v interface{}) error {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result := m.collection.FindOne(ctx, filter)
	if err := result.Err(); err != nil {
		return err
	}
	if err := result.Decode(v); err != nil {
		return err
	}
	return nil
}

func (m *Manager) Find(filter bson.M, sort map[string]int, pagination Pagination) (*mongo.Cursor, error) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result, err := m.collection.Find(
		ctx,
		filter,
		new(options.FindOptions).
			SetSort(sort).
			SetSkip(pagination.Offset).
			SetLimit(pagination.Limit),
	)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		} else {
			return nil, err
		}
	}
	return result, nil
}

type Pagination struct {
	Offset int64
	Limit  int64
}

func DefaultPagination() Pagination {
	return Pagination{Offset: 0, Limit: 20}
}

func DefaultSort() map[string]int {
	return map[string]int{
		"_id": -1,
	}
}
