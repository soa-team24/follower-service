package main

import (
	"context"
	"follower-service/handler"
	"follower-service/model"
	"follower-service/repository"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	//"os/signal"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"follower-service/proto/follower"
)

func seedProfiles(store *repository.FollowRepo) error {
	profiles := []*model.Profile{
		{
			FirstName:      "Andjela",
			LastName:       "Djordjevic",
			ProfilePicture: "profile1.jpg",
			UserID:         1,
			Followers:      nil,
		},
		{
			FirstName:      "Mika",
			LastName:       "Mikic",
			ProfilePicture: "profile2.jpg",
			UserID:         2,
			Followers:      nil,
		},
		{
			FirstName:      "Pera",
			LastName:       "Peric",
			ProfilePicture: "profile2.jpg",
			UserID:         3,
			Followers:      nil,
		},
		{
			FirstName:      "Nina",
			LastName:       "Batranovic",
			ProfilePicture: "profile2.jpg",
			UserID:         4,
			Followers:      nil,
		},
		{
			FirstName:      "Tamara",
			LastName:       "Miljevic",
			ProfilePicture: "profile2.jpg",
			UserID:         5,
			Followers:      nil,
		},

		// Add more profiles as needed
	}

	err1 := store.EmptyBase()
	if err1 != nil {
		return err1
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

	//err = seedProfiles(store)
	if err != nil {
		logger.Fatal("Failed to seed profiles:", err)
	}

	//Initialize the handler and inject said logger
	followsHandler := handler.NewFollowHandler(logger, store)

	/*getAllProfiles := router.Methods(http.MethodGet).Subrouter()
	getAllProfiles.HandleFunc("/profiles", followsHandler.GetAllProfiles)

	postFollowNode := router.Methods(http.MethodPost).Subrouter()
	postFollowNode.HandleFunc("/follow", followsHandler.AddFollow)
	postFollowNode.Use(followsHandler.MiddlewareFollowDeserialization)

	getAllFollowersForUser := router.Methods(http.MethodGet).Subrouter()
	getAllFollowersForUser.HandleFunc("/userFollowers/{userId}", followsHandler.GetAllFollowersForUser)

	getAllFollowersOfMyFollowers := router.Methods(http.MethodGet).Subrouter()
	getAllFollowersOfMyFollowers.HandleFunc("/userSuggestedFollowers/{userId}", followsHandler.GetAllFollowersOfMyFollowers)

	getAllBlogs := router.Methods(http.MethodGet).Subrouter()
	getAllBlogs.HandleFunc("/checkIfFollows/{followerID}/{userID}", followsHandler.CheckIfUserFollows)*/

	lis, err := net.Listen("tcp", ":8000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)

	follower.RegisterFollowServiceServer(grpcServer, followsHandler)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal("server error: ", err)
		}
	}()

	stopCh := make(chan os.Signal)
	signal.Notify(stopCh, syscall.SIGTERM)

	<-stopCh

	grpcServer.Stop()
}
