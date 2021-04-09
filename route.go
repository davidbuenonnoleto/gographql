package main

import "github.com/graphql-go/graphql"

type Route struct {
	Id        string `json:"id,omitempty"`
	User      string `json:"user,omitempty"`
	Zipcode   string `json:"zipcode,omitempty"`
	Numberpkg string `json:"numberpkg,omitempty"`
}

var routeType *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "Route",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"user": &graphql.Field{
			Type: graphql.String,
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
