package handler

import (
	"context"
	"follower-service/mapper"
	"follower-service/model"
	"follower-service/proto/follower"
	"follower-service/repository"
	"log"
	"net/http"
)

type KeyProduct struct{}

type FollowHandler struct {
	follower.UnimplementedFollowServiceServer
	logger *log.Logger
	repo   *repository.FollowRepo
}

// Injecting the logger makes this code much more testable.
func NewFollowHandler(l *log.Logger, r *repository.FollowRepo) *FollowHandler {
	return &FollowHandler{logger: l, repo: r}
}

/*func (m *FollowHandler) AddFollow(rw http.ResponseWriter, h *http.Request) {
	follow := h.Context().Value(KeyProduct{}).(*model.Follow)
	err := m.repo.WriteFollow(follow)
	if err != nil {
		m.logger.Print("Database exception: ", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusCreated)
}*/

func (h *FollowHandler) AddFollow(ctx context.Context, request *follower.AddFollowRequest) (*follower.AddFollowResponse, error) {

	follow := mapper.MapToFollower(request.Follow)

	h.repo.WriteFollow(follow)

	return &follower.AddFollowResponse{}, nil
}

/*func (m *FollowHandler) GetAllFollowersForUser(rw http.ResponseWriter, h *http.Request) {
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
}*/

func (h *FollowHandler) GetAllFollowersForUser(ctx context.Context, request *follower.GetFollowersRequest) (*follower.GetFollowersResponse, error) {
	userId := request.UserId

	id := uint32(userId)

	modelFollows, err := h.repo.GetAllFollowersForUser(id)
	if err != nil {
		h.logger.Print("Database exception: ", err)
	}

	if modelFollows == nil {
		return nil, nil
	}

	var followers []model.Profile

	for _, follow := range modelFollows {
		followers = append(followers, *follow)
	}

	protoFollows := mapper.MapSliceToProtoFollow(followers)
	response := &follower.GetFollowersResponse{
		Followers: protoFollows,
	}
	return response, nil
}

/*
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
}*/

func (h *FollowHandler) GetAllFollowersOfMyFollowers(ctx context.Context, request *follower.GetFollowersOfMyFollowersRequest) (*follower.GetFollowersResponse, error) {
	userId := request.UserId

	id := uint32(userId)

	profiles, err := h.repo.GetAllFollowersOfMyFollowers(id)
	if err != nil {
		println("Database exception: ")
	}

	if profiles == nil {
		return nil, nil
	}

	var profiless []model.Profile

	for _, h := range profiles {
		profiless = append(profiless, *h)
	}

	protoProfiles := mapper.MapSliceToProtoProfiles(profiless)
	response := &follower.GetFollowersResponse{
		Followers: protoProfiles,
	}
	return response, nil

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
/*func (m *FollowHandler) GetAllProfiles(rw http.ResponseWriter, h *http.Request) {
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
}*/

func (h *FollowHandler) GetAllProfiles(ctx context.Context, request *follower.GetAllProfilesRequest) (*follower.GetAllProfilesResponse, error) {
	profiles, err := h.repo.GetAllProfiles()
	if err != nil {
		println("Database exception: ")
	}

	if profiles == nil {
		return nil, nil
	}

	protoProfiles := mapper.MapSliceToProtoProfilesPointer(profiles)
	response := &follower.GetAllProfilesResponse{
		Profiles: protoProfiles,
	}
	return response, nil

}

/*
func (m *FollowHandler) CheckIfUserFollows(rw http.ResponseWriter, h *http.Request) {
	m.logger.Print("Usao u handler: ")
	vars := mux.Vars(h)
	followerID, err := strconv.Atoi(vars["followerID"])
	if err != nil {
		m.logger.Printf("Expected integer, got: %s", vars["userId"])
		http.Error(rw, "Unable to convert user ID to integer", http.StatusBadRequest)
		return
	}
	userID, err := strconv.Atoi(vars["userID"])
	if err != nil {
		m.logger.Printf("Expected integer, got: %s", vars["userId"])
		http.Error(rw, "Unable to convert user ID to integer", http.StatusBadRequest)
		return
	}
	isFollowing, _ := m.repo.IsFollowing(followerID, userID)

	booleanDto := dto.BooleanDto{BoolField: isFollowing}

	err = booleanDto.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		m.logger.Fatal("Unable to convert to json :", err)
		return
	}
}*/

func (m *FollowHandler) CheckIfUserFollows(ctx context.Context, request *follower.CheckIfUserFollowsRequest) (*follower.CheckIfUserFollowsResponse, error) {
	followerID := request.FollowerId
	userID := request.UserId

	isFollowing, _ := m.repo.IsFollowing(int(followerID), int(userID))

	response := &follower.CheckIfUserFollowsResponse{
		IsFollowing: isFollowing,
	}
	return response, nil
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
