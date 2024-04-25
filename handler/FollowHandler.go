package handler

import (
	"context"
	"follower-service/model"
	"follower-service/repository"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type KeyProduct struct{}

type FollowHandler struct {
	logger *log.Logger
	// NoSQL: injecting movie repository
	repo *repository.FollowRepo
}

// Injecting the logger makes this code much more testable.
func NewFollowHandler(l *log.Logger, r *repository.FollowRepo) *FollowHandler {
	return &FollowHandler{l, r}
}

func (m *FollowHandler) AddFollow(rw http.ResponseWriter, h *http.Request) {
	follow := h.Context().Value(KeyProduct{}).(*model.Follow)
	err := m.repo.WriteFollow(follow)
	if err != nil {
		m.logger.Print("Database exception: ", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusCreated)
}

func (m *FollowHandler) GetAllFollowersForUser(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	userID, err := strconv.Atoi(vars["userId"])
	if err != nil {
		m.logger.Printf("Expected integer, got: %s", vars["userId"])
		http.Error(rw, "Unable to convert user ID to integer", http.StatusBadRequest)
		return
	}

	userFollowers, err := m.repo.GetAllFollowersForUser(uint32(userID))
	if err != nil {
		m.logger.Print("Database exception: ", err)
	}

	if userFollowers == nil {
		return
	}

	err = userFollowers.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		m.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (m *FollowHandler) GetAllFollowersOfMyFollowers(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	userID, err := strconv.Atoi(vars["userId"])
	if err != nil {
		m.logger.Printf("Expected integer, got: %s", vars["userId"])
		http.Error(rw, "Unable to convert user ID to integer", http.StatusBadRequest)
		return
	}

	userFollowers, err := m.repo.GetAllFollowersOfMyFollowers(uint32(userID))
	if err != nil {
		m.logger.Print("Database exception: ", err)
	}

	if userFollowers == nil {
		return
	}

	err = userFollowers.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		m.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

/*
	func (m *FollowHandler) GetAllFollowers(rw http.ResponseWriter, h *http.Request) {
		vars := mux.Vars(h)
		limit, err := strconv.Atoi(vars["limit"])
		if err != nil {
			m.logger.Printf("Expected integer, got: %d", limit)
			http.Error(rw, "Unable to convert limit to integer", http.StatusBadRequest)
			return
		}

		movies, err := m.repo.GetAllNodesWithFollowLabel()
		if err != nil {
			m.logger.Print("Database exception: ", err)
		}

		if movies == nil {
			return
		}

		err = movies.ToJSON(rw)
		if err != nil {
			http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
			m.logger.Fatal("Unable to convert to json :", err)
			return
		}
	}
*/
func (m *FollowHandler) GetAllProfiles(rw http.ResponseWriter, h *http.Request) {
	m.logger.Print("Usao u handler: ")
	profiles, err := m.repo.GetAllProfiles()
	m.logger.Print("Vratio iz repository: ")
	if err != nil {
		m.logger.Print("Database exception: ", err)
	}

	if profiles == nil {
		return
	}

	err = profiles.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		m.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (m *FollowHandler) MiddlewareContentTypeSet(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		m.logger.Println("Method [", h.Method, "] - Hit path :", h.URL.Path)

		rw.Header().Add("Content-Type", "application/json")

		next.ServeHTTP(rw, h)
	})
}

func (m *FollowHandler) MiddlewareFollowDeserialization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		follow := &model.Follow{}
		err := follow.FromJSON(h.Body)
		if err != nil {
			http.Error(rw, "Unable to decode json", http.StatusBadRequest)
			m.logger.Fatal(err)
			return
		}
		ctx := context.WithValue(h.Context(), KeyProduct{}, follow)
		h = h.WithContext(ctx)
		next.ServeHTTP(rw, h)
	})
}
