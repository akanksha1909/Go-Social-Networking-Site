package mongo

import (
	"fmt"
	"strings"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"Go-Social/pkg"
)

type UserService struct {
	collection *mgo.Collection
}

func NewUserService(session *mgo.Session, config *root.MongoConfig) *UserService {
	collection := session.DB(config.DbName).C("user")
	collection.EnsureIndex(userModelIndex())
	return &UserService{collection}
}

func (p *UserService) CreateUser(u *root.User) error {
	user, err := newUserModel(u)
	if err != nil {
		return err
	}
	return p.collection.Insert(&user)
}

func (p *UserService) CheckUserName(username string) bool {
	model := userModel{}
	// use regex
	p.collection.Find(bson.M{"username": username}).One(&model)
	if model.Username != "" {
		return false
	}
	return true
}

func (p *UserService) CheckEmail(email string) bool {
	model := userModel{}
	email = strings.ToLower(email)
	p.collection.Find(bson.M{"email": strings.ToLower(email)}).One(&model)
	if model.Email != "" {
		return false
	}
	return true
}

func (p *UserService) HandleSecret(secret string) (root.User, error) {
	model := userModel{}
	condition := bson.M{
		"$and": []bson.M{
			bson.M{"verificationsecret": secret},
			bson.M{"verified": false},
		},
	}
	change := bson.M{"$set": bson.M{"verified": true, "verifiedon": time.Now(), "updatedat": time.Now()}}
	err := p.collection.Update(condition, change)
	if err != nil {
		fmt.Println("err")
		fmt.Println(err)
		return root.User{
			ID:        model.ID.Hex(),
			Username:  model.Username,
			Password:  "-",
			UpdatedAt: model.UpdatedAt}, err
	}
	err1 := p.collection.Find(bson.M{"verificationsecret": secret}).One(&model)
	if err1 != nil {
		fmt.Println("err1")
		fmt.Println(err1)
		return root.User{
			ID:        model.ID.Hex(),
			Username:  model.Username,
			Password:  "-",
			UpdatedAt: model.UpdatedAt}, err1
	}
	return root.User{
		ID:        model.ID.Hex(),
		Username:  model.Username,
		Password:  "-",
		UpdatedAt: model.UpdatedAt}, nil
}

func (p *UserService) UpdateUser(fields []string, VerificationSecret string, email string) error {
	// make it generic

	email = strings.ToLower(email)

	condition := bson.M{
		"$and": []bson.M{
			bson.M{"email": email},
			bson.M{"verified": false},
		},
	}
	change := bson.M{
		"$set": bson.M{
			"verificationsecret": VerificationSecret,
			"updatedat":          time.Now(),
		},
	}
	err := p.collection.Update(condition, change)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (p *UserService) GetUserByUsername(username string) (root.User, error) {
	model := userModel{}
	err := p.collection.Find(bson.M{"username": username}).One(&model)
	return root.User{
		ID:       model.ID.Hex(),
		Username: model.Username,
		Password: "-"}, err
}
