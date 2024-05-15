package mapper

import (
	"follower-service/model"
	"soa/grpc/proto/follower"
	p "soa/grpc/proto/follower"
)

func MapToPFollow(follow *model.Follow) *p.Follow {
	followP := &p.Follow{
		ProfileId:  follow.ProfileID,
		FollowerId: follow.FollowerID,
	}
	return followP
}

func MapToFollower(followP *p.Follow) *model.Follow {
	follow := &model.Follow{
		ProfileID:  followP.ProfileId,
		FollowerID: followP.FollowerId,
	}
	return follow
}

func MapToPProfile(profile *model.Profile) *p.Profile {
	profileP := &p.Profile{
		Id:             profile.ID,
		FirstName:      profile.FirstName,
		LastName:       profile.LastName,
		ProfilePicture: profile.ProfilePicture,
		UserId:         profile.UserID,
	}
	for _, follower := range profile.Followers {
		profileP.Followers = append(profileP.Followers, MapToPProfile(&follower))
	}
	return profileP
}

func MapToProfile(profileP *p.Profile) *model.Profile {
	profile := &model.Profile{
		ID:             profileP.Id,
		FirstName:      profileP.FirstName,
		LastName:       profileP.LastName,
		ProfilePicture: profileP.ProfilePicture,
		UserID:         profileP.UserId,
	}
	for _, followerP := range profileP.Followers {
		profile.Followers = append(profile.Followers, *MapToProfile(followerP))
	}
	return profile
}

func MapSliceToProtoProfiles(modelProfiles []model.Profile) []*p.Profile {
	var protoProfiles []*p.Profile

	for _, modelProfile := range modelProfiles {
		protoProfile := MapToPProfile(&modelProfile)
		protoProfiles = append(protoProfiles, protoProfile)
	}

	return protoProfiles
}

func MapSliceToProtoProfilesPointer(modelProfiles []*model.Profile) []*p.Profile {
	var protoProfiles []*p.Profile

	for _, modelProfile := range modelProfiles {
		protoProfile := MapToPProfile(modelProfile)
		protoProfiles = append(protoProfiles, protoProfile)
	}

	return protoProfiles
}
func MapSliceToModelProfiles(protoProfiles []*p.Profile) []model.Profile {
	var modelProfiles []model.Profile

	for _, protoProfile := range protoProfiles {
		modelProfile := MapToProfile(protoProfile)
		modelProfiles = append(modelProfiles, *modelProfile)
	}

	return modelProfiles
}

func MapSliceToProtoFollow(modelFollows []model.Profile) []*follower.Profile {
	var protoFollows []*follower.Profile

	for _, modelFollow := range modelFollows {
		protoFollow := MapToPProfile(&modelFollow)
		protoFollows = append(protoFollows, protoFollow)
	}

	return protoFollows
}
