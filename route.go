package main

import "github.com/graphql-go/graphql"

type Route struct {
	Id        string `json:"id,omitempty" validate:"omitempty,uuid"`
	User      string `json:"user,omitempty" validate:"isdefault"`
	Zipcode   string `json:"zipcode,omitempty" validate:"required"`
	Numberpkg string `json:"numberpkg,omitempty" validate:"required"`
}

var routeType *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "Route",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"user": &graphql.Field{
			Type: userType,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				route := params.Source.(Route)
				for _, user := range users {
					if user.Id == route.User {
						return user, nil
					}
				}
				return nil, nil
			},
		},
		"zipcode": &graphql.Field{
			Type: graphql.String,
		},
		"numberpkg": &graphql.Field{
			Type: graphql.String,
		},
	},
})

var routeInputType *graphql.InputObject = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "RouteInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"id": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"zipcode": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"numberpkg": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
	},
})
