package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	PORT        = "8081"
	MongoServer = "mongodb://mongodb:27017"
)

type Middleware func(http.HandlerFunc) http.HandlerFunc

// Logging logs all requests with its path and the time it took to process
func Logging() Middleware {

	// Create a new Middleware
	return func(f http.HandlerFunc) http.HandlerFunc {

		// Define the http.HandlerFunc
		return func(w http.ResponseWriter, r *http.Request) {

			// Do middleware things
			start := time.Now()
			defer func() { log.Println(r.URL.Path, time.Since(start)) }()

			// Call the next middleware/handler in chain
			f(w, r)
		}
	}
}

// Method ensures that url can only be requested with a specific method, else returns a 400 Bad Request
func Method(m string) Middleware {

	// Create a new Middleware
	return func(f http.HandlerFunc) http.HandlerFunc {

		// Define the http.HandlerFunc
		return func(w http.ResponseWriter, r *http.Request) {

			// Do middleware things
			if r.Method != m {
				http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				return
			}

			// Call the next middleware/handler in chain
			f(w, r)
		}
	}
}

// Chain applies middlewares to a http.HandlerFunc
func Chain(f http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	for _, m := range middlewares {
		f = m(f)
	}
	return f
}

func getAllPatientNames(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	client, err := mongo.NewClient(options.Client().ApplyURI(MongoServer))
	if err != nil {
		log.Fatal(err)
	}

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	healthcareDatabase := client.Database("healthcare")
	patientsCollection := healthcareDatabase.Collection("patients")

	cursor, err := patientsCollection.Find(ctx, bson.M{})
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	defer cursor.Close(ctx)

	var myData = make(map[string]string)
	for cursor.Next(ctx) {
		var patient bson.M
		if err = cursor.Decode(&patient); err != nil {
			log.Fatal(err)
		}
		myData[fmt.Sprintf("%v", patient["_id"])] = fmt.Sprintf("%v", patient["first_name"])
	}

	myjson, err := json.Marshal(myData)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Fprintln(w, string(myjson))
}

func getPatient(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	client, err := mongo.NewClient(options.Client().ApplyURI(MongoServer))
	if err != nil {
		log.Fatal(err)
	}

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	healthcareDatabase := client.Database("healthcare")
	patientsCollection := healthcareDatabase.Collection("patients")

	r.ParseForm()

	fmt.Println(r.Form)

	var patient bson.M
	if err = patientsCollection.FindOne(ctx, bson.M{"first_name": r.Form.Get("first_name")}).Decode(&patient); err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusBadRequest)+" Missing first_name value in parameters", http.StatusBadRequest)
		return
	}

	myjson, _ := json.Marshal(patient)
	fmt.Fprintln(w, string(myjson))
}

func addPatient(w http.ResponseWriter, r *http.Request) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	client, err := mongo.NewClient(options.Client().ApplyURI(MongoServer))
	if err != nil {
		log.Fatal(err)
	}

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	healthcareDatabase := client.Database("healthcare")
	patientsCollection := healthcareDatabase.Collection("patients")

	r.ParseForm()

	data := make(map[string]interface{})
	for k, v := range r.PostForm {
		if len(v) == 1 {
			data[k] = v[0]
		} else {
			data[k] = v
		}
	}

	_, err = patientsCollection.InsertOne(ctx, data)
	if err != nil {
		log.Fatal(err)
	}

}

func main() {
	http.HandleFunc("/user/add",
		Chain(
			addPatient,
			Method("POST"),
			Logging()))
	http.HandleFunc("/user/view",
		Chain(
			getPatient,
			Method("GET"),
			Logging()))
	http.HandleFunc("/user/all",
		Chain(
			getAllPatientNames,
			Method("GET"),
			Logging()))

	fs := http.FileServer(http.Dir("static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Printf("Starting Server on Port: %s", PORT)
	log.Fatalln(http.ListenAndServe(":"+PORT, nil))
}
