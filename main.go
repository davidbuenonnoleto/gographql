package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/graphql-go/graphql"
)

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

func main() {
	fmt.Println("Starting the application")
	router := mux.NewRouter()
	schema, _ := graphql.NewSchema(graphql.SchemaConfig{
		Query: rootQuery,
	})
	router.HandleFunc("/graphql", func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("content-type", "application/json")
		result := graphql.Do(graphql.Params{
			Schema: schema,
		})
		json.NewEncoder(response).Encode(result)
	})
}
