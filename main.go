package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/graphql-go/graphql"
	"github.com/mitchellh/mapstructure"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/couchbase/gocb.v1"
	"gopkg.in/go-playground/validator.v9"
)

type GraphQLPayload struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type CustomJWTClaims struct {
	Id string `json:"id"`
	jwt.StandardClaims
}

var JWT_SECRET []byte = []byte("beyondmonkeys")
var bucket *gocb.Bucket

func ValidateJWT(t string) (interface{}, error) {
	token, err := jwt.Parse(t, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", token.Header["alg"])
		}
		return JWT_SECRET, nil
	})
	if err != nil {
		return nil, errors.New(`{ "message": "` + err.Error() + `" }`)
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		var tokenData CustomJWTClaims
		mapstructure.Decode(claims, &tokenData)
		return tokenData, nil
	} else {
		return nil, errors.New(`{ "message": "invalid token" }`)
	}
}

var rootQuery *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "Query",
	Fields: graphql.Fields{
		"users": &graphql.Field{
			Type: graphql.NewList(userType),
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				var users []User
				query := gocb.NewN1qlQuery(`SELECT ` + bucket.Name() + `.* FROM ` + bucket.Name() + ` WHERE type = 'manager'`)
				rows, err := bucket.ExecuteN1qlQuery(query, nil)
				if err != nil {
					return nil, err
				}
				var row User
				for rows.Next(&row) {
					users = append(users, row)
				}
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
				var user User
				_, err := bucket.Get(id, &user)
				if err != nil {
					return nil, err
				}
				return user, nil
			},
		},
		"routes": &graphql.Field{
			Type: graphql.NewList(routeType),
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				var routes []Route
				query := gocb.NewN1qlQuery(`SELECT ` + bucket.Name() + `.* FROM ` + bucket.Name() + ` WHERE type = 'article'`)
				rows, err := bucket.ExecuteN1qlQuery(query, nil)
				if err != nil {
					return nil, err
				}
				var row Route
				for rows.Next(&row) {
					routes = append(routes, row)
				}
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
				var route Route
				_, err := bucket.Get(id, &route)
				if err != nil {
					return nil, err
				}
				return route, nil
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
				_, err := bucket.Remove(id, 0)
				if err != nil {
					return nil, err
				}
				return id, nil
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
				mutation := bucket.MutateIn(changes.Id, 0, 0)
				if changes.Firstname != "" {
					mutation.Upsert("firstname", changes.Firstname, true)
				}
				if changes.Lastname != "" {
					mutation.Upsert("lastname", changes.Lastname, true)
				}
				if changes.Username != "" {
					mutation.Upsert("username", changes.Username, true)
				}
				if changes.Password != "" {
					err := validate.Var(changes.Password, "gte=4")
					if err != nil {
						return nil, err
					}
					hash, _ := bcrypt.GenerateFromPassword([]byte(changes.Password), 10)
					mutation.Upsert("password", string(hash), true)
				}
				mutation.Execute()
				return changes, nil
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
				mapstructure.Decode(params.Args["article"], &route)
				decoded, err := ValidateJWT(params.Context.Value("token").(string))
				if err != nil {
					return nil, err
				}
				validate := validator.New()
				err = validate.Struct(route)
				if err != nil {
					return nil, err
				}
				route.Id = uuid.Must(uuid.NewV4()).String()
				route.User = decoded.(CustomJWTClaims).Id
				route.Type = "peninisula"
				bucket.Insert(route.Id, route, 0)
				return route, nil
			},
		},
	},
})

func main() {
	fmt.Println("Starting the application")

	/* couchbase connect*/
	cluster, _ := gocb.Connect("couchbase://localhost")
	cluster.Authenticate(gocb.PasswordAuthenticator{
		Username: "demo",
		Password: "123456",
	})
	/* bucket selection */
	bucket, _ = cluster.OpenBucket("graphql", "")

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
			Context:        context.WithValue(context.Background(), "token", request.URL.Query().Get("token")),
		})
		json.NewEncoder(response).Encode(result)
	})
	headers := handlers.AllowedHeaders(
		[]string{
			"Content-type",
			"Authorization",
			"X-Requested-With",
		},
	)
	methods := handlers.AllowedMethods(
		[]string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
		},
	)
	origins := handlers.AllowedOrigins(
		[]string{
			"*",
		},
	)
	http.ListenAndServe(
		":8080",
		handlers.CORS(headers, methods, origins)(router),
	)
}
