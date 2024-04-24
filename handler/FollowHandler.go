package handler

import (
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
	vars := mux.Vars(h)
	limit, err := strconv.Atoi(vars["limit"])
	if err != nil {
		m.logger.Printf("Expected integer, got: %d", limit)
		http.Error(rw, "Unable to convert limit to integer", http.StatusBadRequest)
		return
	}

	profiles, err := m.repo.GetAllProfiles()
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
