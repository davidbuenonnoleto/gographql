package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/graphql-go/graphql"
	"github.com/mitchellh/mapstructure"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/go-playground/validator.v9"
)

type GraphQLPayload struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

var users []User = []User{
	{
		Id:        "user-1",
		Firstname: "David",
		Lastname:  "Noleto",
		Username:  "dbone",
		Password:  "senha132",
	},
}

var routes []Route = []Route{
	{
		Id:        "route-1",
		User:      "user-1",
		Zipcode:   "94015",
		Numberpkg: "15",
	},
}

var rootQuery *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "Query",
	Fields: graphql.Fields{
		"users": &graphql.Field{
			Type: graphql.NewList(userType),
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				return users, nil
			},
		},
		"user": &graphql.Field{
			Type: userType,
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				id := params.Args["id"].(string)
				for _, user := range users {
					if user.Id == id {
						return user, nil
					}
				}
				return nil, nil
			},
		},
		"routes": &graphql.Field{
			Type: graphql.NewList(routeType),
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				return routes, nil
			},
		},
		"route": &graphql.Field{
			Type: routeType,
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				id := params.Args["id"].(string)
				for _, route := range routes {
					if route.Id == id {
						return route, nil
					}
				}
				return nil, nil
			},
		},
	},
})

var rootMutation *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "Mutation",
	Fields: graphql.Fields{
		"deleteUser": &graphql.Field{
			Type: graphql.NewList(userType),
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				id := params.Args["id"].(string)
				for index, user := range users {
					if user.Id == id {
						users = append(users[:index], users[index+1:]...)
						return users, nil
					}
				}
				return nil, nil
			},
		},
		"updateUser": &graphql.Field{
			Type: graphql.NewList(userType),
			Args: graphql.FieldConfigArgument{
				"user": &graphql.ArgumentConfig{
					Type: userInputType,
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				var changes User
				mapstructure.Decode(params.Args["user"], &changes)
				validate := validator.New()
				for index, user := range users {
					if user.Id == changes.Id {
						if changes.Firstname != "" {
							user.Firstname = changes.Firstname
						}
						if changes.Lastname != "" {
							user.Lastname = changes.Lastname
						}
						if changes.Username != "" {
							user.Username = changes.Username
						}
						if changes.Password != "" {
							err := validate.Var(changes.Password, "gte=4")
							if err != nil {
								return nil, err
							}
							hash, _ := bcrypt.GenerateFromPassword([]byte(changes.Password), 10)
							user.Password = string(hash)
						}
						users[index] = user
						return users, nil
					}
				}
				return nil, nil
			},
		},
		"createRoute": &graphql.Field{
			Type: graphql.NewList(routeType),
			Args: graphql.FieldConfigArgument{
				"route": &graphql.ArgumentConfig{
					Type: routeInputType,
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				var route Route
				mapstructure.Decode(params.Args["route"], &route)
				validate := validator.New()
				err := validate.Struct(route)
				if err != nil {
					return nil, err
				}
				route.Id = uuid.Must(uuid.NewV4()).String()
				route.User = "dbone"
				routes = append(routes, route)
				return routes, nil
			},
		},
	},
})

func main() {
	fmt.Println("Starting the application")
	router := mux.NewRouter()
	schema, _ := graphql.NewSchema(graphql.SchemaConfig{
		Query:    rootQuery,
		Mutation: rootMutation,
	})
	router.HandleFunc("/register", RegisterEndpoint).Methods("POST")
	router.HandleFunc("/login", LoginEndpoint).Methods("POST")
	router.HandleFunc("/graphql", func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("content-type", "application/json")
		var payload GraphQLPayload
		json.NewDecoder(request.Body).Decode(&payload)
		result := graphql.Do(graphql.Params{
			Schema:         schema,
			RequestString:  payload.Query,
			VariableValues: payload.Variables,
		})
		json.NewEncoder(response).Encode(result)
	})
	http.ListenAndServe(":12345", router)
}
