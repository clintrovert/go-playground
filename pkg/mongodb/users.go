package mongodb

import (
	"context"
	"errors"
	"strings"

	"github.com/readactedworks/go-http-server/api/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	userId = "_id"
)

// UserDatabase provides access to User-specific actions in Mongo database.
type UserDatabase struct {
	Database
}

// NewUserDatabase creates a new instance of UserDatabase.
func NewUserDatabase(collection *mongo.Collection) *UserDatabase {
	return &UserDatabase{Database{collection: collection}}
}

// GetUser retrieves a user from the Mongo database.
func (u *UserDatabase) GetUser(
	ctx context.Context,
	id string,
) (*model.User, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("id was missing")
	}

	// convert id string to ObjectId
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := bson.D{{userId, objectId}}
	var user model.User
	err = u.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// CreateUser generates a new user in the Mongo database.
func (u *UserDatabase) CreateUser(ctx context.Context, user *model.User) error {
	objectId := primitive.NewObjectID()
	user.Id = objectId.Hex()
	_, err := u.collection.InsertOne(ctx, convertUserToDoc(user))
	if err != nil {
		return err
	}

	return nil
}

// UpdateUser updates a user in the Mongo database.
func (u *UserDatabase) UpdateUser(ctx context.Context, user *model.User) error {
	objectId, err := primitive.ObjectIDFromHex(user.Id)
	if err != nil {
		return err
	}
	filter := bson.D{{userId, objectId}}
	update := bson.D{{"$set", convertUserToDoc(user)}}
	res, err := u.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if res.ModifiedCount != 1 {
		return errors.New("user not updated")
	}

	return nil
}

// DeleteUser deletes a user from the Mongo database.
func (u *UserDatabase) DeleteUser(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("id was missing")
	}

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.D{{userId, objectId}}
	res, err := u.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if res.DeletedCount != 1 {
		return errors.New("user not deleted")
	}

	return nil
}

func convertUserToDoc(user *model.User) bson.D {
	objId, err := primitive.ObjectIDFromHex(user.Id)
	if err != nil {
		return nil
	}
	return bson.D{
		{userId, objId},
		{"name", user.Name},
		{"email", user.Email},
		{"password", user.Password},
	}
}
