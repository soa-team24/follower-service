package main

import (
	"context"
	"follower-service/handler"
	"follower-service/model"
	"follower-service/repository"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func seedProfiles(store *repository.FollowRepo) error {
	profiles := []*model.Profile{
		{
			ID:             5,
			FirstName:      "Marko",
			LastName:       "Filipovic",
			ProfilePicture: "profile1.jpg",
			UserID:         2,
			Follows:        nil,
		},
		{
			ID:             4,
			FirstName:      "Maja",
			LastName:       "Petrovic",
			ProfilePicture: "profile2.jpg",
			UserID:         1,
			Follows:        nil,
		},
		{
			ID:             6,
			FirstName:      "Mina",
			LastName:       "Jovanovic",
			ProfilePicture: "profile2.jpg",
			UserID:         3,
			Follows:        nil,
		},
		{
			ID:             7,
			FirstName:      "Petar",
			LastName:       "Ivanovic",
			ProfilePicture: "profile2.jpg",
			UserID:         4,
			Follows:        nil,
		},

		// Add more profiles as needed
	}

	// Iterate over profiles and save them to the database
	for _, profile := range profiles {
		err := store.WriteProfile(profile)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	//Reading from environment, if not set we will default it to 8080.
	//This allows flexibility in different environments (for eg. when running multiple docker api's and want to override the default port)
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8083"
	}

	// Initialize context
	timeoutContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	//Initialize the logger we are going to use, with prefix and datetime for every log
	logger := log.New(os.Stdout, "[follow-api] ", log.LstdFlags)
	storeLogger := log.New(os.Stdout, "[follow-store] ", log.LstdFlags)

	// NoSQL: Initialize Movie Repository store
	store, err := repository.New(storeLogger)
	if err != nil {
		logger.Fatal(err)
	}
	defer store.CloseDriverConnection(timeoutContext)
	store.CheckConnection()

	err = seedProfiles(store)
	if err != nil {
		logger.Fatal("Failed to seed profiles:", err)
	}

	//Initialize the handler and inject said logger
	followsHandler := handler.NewFollowHandler(logger, store)

	//Initialize the router and add a middleware for all the requests
	router := mux.NewRouter()

	router.Use(followsHandler.MiddlewareContentTypeSet)

	getAllProfiles := router.Methods(http.MethodGet).Subrouter()
	getAllProfiles.HandleFunc("/profiles", followsHandler.GetAllProfiles)

	//postPersonNode.Use(moviesHandler.MiddlewarePersonDeserialization)

	cors := gorillaHandlers.CORS(gorillaHandlers.AllowedOrigins([]string{"*"}))

	//Initialize the server
	server := http.Server{
		Addr:         ":" + port,
		Handler:      cors(router),
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	logger.Println("Server listening on port", port)
	//Distribute all the connections to goroutines
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			logger.Fatal(err)
		}
	}()

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt)
	signal.Notify(sigCh, os.Kill)

	sig := <-sigCh
	logger.Println("Received terminate, graceful shutdown", sig)

	//Try to shutdown gracefully
	if server.Shutdown(timeoutContext) != nil {
		logger.Fatal("Cannot gracefully shutdown...")
	}
	logger.Println("Server stopped")
}
