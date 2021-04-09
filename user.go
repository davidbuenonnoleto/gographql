package main

import (
	"encoding/json"
	"net/http"

	"github.com/graphql-go/graphql"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/go-playground/validator.v9"
)

type User struct {
	Id        string `json:"id,omitempty" validate:"omitempty,uuid"`
	Firstname string `json:"firstname,omitempty" validate:"required"`
	Lastname  string `json:"lastname,omitempty" validate:"required"`
	Username  string `json:"username,omitempty" validate:"required"`
	Password  string `json:"password,omitempty" validate:"required,gte=4"`
}

/* defining a graphql object */
var userType *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "User",
	Fields: &graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"firstname": &graphql.Field{
			Type: graphql.String,
		},
		"lastname": &graphql.Field{
			Type: graphql.String,
		},
		"username": &graphql.Field{
			Type: graphql.String,
		},
		"password": &graphql.Field{
			Type: graphql.String,
		},
	},
})

/* defining graphql input object configuration */
var userInputType *graphql.InputObject = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "UserInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"id": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"firstname": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"lastname": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"username": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"password": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
	},
})

func RegisterEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	var user User
	json.NewDecoder(request.Body).Decode(&user)
	validate := validator.New()
	err := validate.Struct(user)
	if err != nil {
		response.WriteHeader(500)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	user.Id = uuid.Must(uuid.NewV4()).String()
	hash, _ := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
	user.Password = string(hash)
	users = append(users, user)
	json.NewEncoder(response).Encode(users)
}

func LoginEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	var data User
	json.NewDecoder(request.Body).Decode(&data)
	validate := validator.New()
	/* it ignores the anotations validation on the struct */
	err := validate.StructExcept(data, "Firstname", "Lastname")
	if err != nil {
		response.WriteHeader(500)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	for _, user := range users {
		if user.Username == data.Username {
			err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(data.Password))
			if err != nil {
				response.WriteHeader(500)
				response.Write([]byte(`{ "message": "invalid password" }`))
				return
			}
			json.NewEncoder(response).Encode(user)
		}
	}
	json.NewEncoder(response).Encode(User{})
}
