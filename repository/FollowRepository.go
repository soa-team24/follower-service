package repository

import (
	"context"
	"log"
	"os"

	// NoSQL: module containing Neo4J api client
	"follower-service/model"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// NoSQL: MovieRepo struct encapsulating Neo4J api client
type FollowRepo struct {
	// Thread-safe instance which maintains a database connection pool
	driver neo4j.DriverWithContext
	logger *log.Logger
}

// NoSQL: Constructor which reads db configuration from environment and creates a keyspace
func New(logger *log.Logger) (*FollowRepo, error) {
	// Local instance
	uri := os.Getenv("NEO4J_DB")
	user := os.Getenv("NEO4J_USERNAME")
	pass := os.Getenv("NEO4J_PASS")
	auth := neo4j.BasicAuth(user, pass, "")

	driver, err := neo4j.NewDriverWithContext(uri, auth)
	if err != nil {
		logger.Panic(err)
		return nil, err
	}

	// Return repository with logger and DB session
	return &FollowRepo{
		driver: driver,
		logger: logger,
	}, nil
}

// Check if connection is established
func (mr *FollowRepo) CheckConnection() {
	ctx := context.Background()
	err := mr.driver.VerifyConnectivity(ctx)
	if err != nil {
		mr.logger.Panic(err)
		return
	}
	// Print Neo4J server address
	mr.logger.Printf(`Neo4J server address: %s`, mr.driver.Target().Host)
}

// Disconnect from database
func (mr *FollowRepo) CloseDriverConnection(ctx context.Context) {
	mr.driver.Close(ctx)
}

func (mr *FollowRepo) GetAllProfiles() (model.Profiles, error) {
	mr.logger.Println("Usao u repo")
	ctx := context.Background()
	session := mr.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	// ExecuteRead for read transactions (Read and queries)
	profileResults, err := session.ExecuteRead(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx,
				`MATCH (profile:Profile)
				RETURN profile.userID as profileID, profile.firstName as firstName, profile.lastName as lastName, 
				profile.profilePicture as profilePicture, profile.userID as userID `,
				map[string]any{})
			if err != nil {
				return nil, err
			}

			// Option 1: we iterate over result while there are records
			var profiles model.Profiles
			for result.Next(ctx) {
				record := result.Record()

				ID, _ := record.Get("profileID")
				FirstName, _ := record.Get("firstName")
				LastName, _ := record.Get("lastName")
				ProfilePicture, _ := record.Get("profilePicture")
				UserID, _ := record.Get("userID")
				profiles = append(profiles, &model.Profile{
					ID:             ID.(int64),
					FirstName:      FirstName.(string),
					LastName:       LastName.(string),
					ProfilePicture: ProfilePicture.(string),
					UserID:         UserID.(int64),
				})
			}
			return profiles, nil
			// Option 2: we collect all records from result and iterate and map outside of the transaction
			// return result.Collect(ctx)
		})
	if err != nil {
		mr.logger.Println("Error querying search:", err)
		return nil, err
	}
	return profileResults.(model.Profiles), nil
}

func (mr *FollowRepo) WriteProfile(profile *model.Profile) error {
	// Neo4J Sessions are lightweight so we create one for each transaction (Cassandra sessions are not lightweight!)
	// Sessions are NOT thread safe
	ctx := context.Background()
	session := mr.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	// ExecuteWrite for write transactions (Create/Update/Delete)
	savedProfile, err := session.ExecuteWrite(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx,
				"CREATE (p:Profile) SET p.firstName = $firstName, p.lastName = $lastName, p.profilePicture = $profilePicture, p.userID = $userID  RETURN p.firstName + p.lastName + p.profilePicture + p.userID + ', from node ' + id(p)",
				map[string]any{"id": profile.ID, "firstName": profile.FirstName, "lastName": profile.LastName, "profilePicture": profile.ProfilePicture, "userID": profile.UserID})
			if err != nil {
				return nil, err
			}

			if result.Next(ctx) {
				return result.Record().Values[0], nil
			}

			return nil, result.Err()
		})
	if err != nil {
		mr.logger.Println("Error inserting Person:", err)
		return err
	}
	mr.logger.Println(savedProfile.(string))
	return nil
}

func (mr *FollowRepo) EmptyBase() error {
	// Neo4J Sessions are lightweight so we create one for each transaction (Cassandra sessions are not lightweight!)
	// Sessions are NOT thread safe
	ctx := context.Background()
	session := mr.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	// ExecuteWrite for write transactions (Create/Update/Delete)
	_, err := session.ExecuteWrite(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx,
				"MATCH (n) DETACH DELETE n",
				map[string]any{})
			if err != nil {
				return nil, err
			}

			return nil, result.Err()
		})
	if err != nil {
		mr.logger.Println("Error inserting Person:", err)
		return err
	}
	mr.logger.Println("Succesfully deleted base")
	return nil
}

// ///////
func (mr *FollowRepo) WriteFollow(follow *model.Follow) error {
	// Neo4J Sessions are lightweight so we create one for each transaction (Cassandra sessions are not lightweight!)
	// Sessions are NOT thread safe
	ctx := context.Background()
	session := mr.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	// ExecuteWrite for write transactions (Create/Update/Delete)
	_, err := session.ExecuteWrite(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			_, err := transaction.Run(ctx,
				"MATCH (p:Profile), (f:Profile) WHERE p.userID = $profileID AND f.userID = $followerID CREATE (f)-[:FOLLOWS]->(p)",
				map[string]interface{}{
					"profileID":  follow.ProfileID,
					"followerID": follow.FollowerID,
				})
			if err != nil {
				return nil, err
			}

			return nil, nil
		})
	if err != nil {
		mr.logger.Println("Error inserting Person:", err)
		return err
	}
	mr.logger.Println("Added follow sucesfuly")
	return nil
}

func (mr *FollowRepo) GetAllFollowersForUser(id uint32) (model.Profiles, error) {
	ctx := context.Background()
	session := mr.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	// ExecuteRead for read transactions (Read and queries)
	movieResults, err := session.ExecuteRead(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx,
				`MATCH (p1:Profile)<-[:FOLLOWS]-(p2:Profile)
				WHERE p1.userID = $profileID
				RETURN collect({ id: p2.userID, firstName: p2.firstName, lastName: p2.lastName, profilePicture: p2.profilePicture, userID: p2.userID }) as followers`,

				map[string]any{"profileID": id})
			if err != nil {
				return nil, err
			}
			var followersRetVal model.Profiles

			for result.Next(ctx) {
				//mr.logger.Println("Found one profile:")
				record := result.Record()
				followers, _ := record.Get("followers")
				followersRetVal = mr.convertDataToProfileSlice(followers)

			}

			if followersRetVal == nil {
				followersRetVal = make(model.Profiles, 0)
				mr.logger.Println("Profile NOT FOUND:")
				return followersRetVal, nil
			}

			return followersRetVal, nil
		})
	if err != nil {
		mr.logger.Println("Error querying search:", err)
		return nil, err
	}
	return movieResults.(model.Profiles), nil
}

func (mr *FollowRepo) GetAllFollowersOfMyFollowers(id uint32) (model.Profiles, error) {
	ctx := context.Background()
	session := mr.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	// ExecuteRead for read transactions (Read and queries)
	movieResults, err := session.ExecuteRead(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx,
				`MATCH (p1:Profile)-[:FOLLOWS]->(p2:Profile)-[:FOLLOWS]->(p3:Profile)
				WHERE p1.userID = $profileID
				RETURN collect({ id: p3.userID, firstName: p3.firstName, lastName: p3.lastName, profilePicture: p3.profilePicture, userID: p3.userID }) as followers`,

				map[string]any{"profileID": id})
			if err != nil {
				return nil, err
			}
			var followersRetVal model.Profiles

			for result.Next(ctx) {
				//mr.logger.Println("Found one profile:")
				record := result.Record()
				followers, _ := record.Get("followers")
				followersRetVal = mr.convertDataToProfileSlice(followers)

			}

			if followersRetVal == nil {
				followersRetVal = make(model.Profiles, 0)
				mr.logger.Println("Profile NOT FOUND:")
				return followersRetVal, nil
			}

			return followersRetVal, nil
		})
	if err != nil {
		mr.logger.Println("Error querying search:", err)
		return nil, err
	}
	return movieResults.(model.Profiles), nil
}

func (fr *FollowRepo) IsFollowing(followerID int, userID int) (bool, error) {
	ctx := context.Background()
	session := fr.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	// Izvršite upit za provjeru postoji li veza zapraćivanja između korisnika i autora
	boolResult, err := session.ExecuteRead(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx,
				"MATCH (f:Profile)-[:FOLLOWS]->(u:Profile) WHERE f.userID = $followerID AND u.userID = $userID RETURN COUNT(*) as count",
				map[string]any{"followerID": followerID, "userID": userID})
			if err != nil {
				return nil, err
			}
			if result.Next(ctx) {
				return result.Record().Values[0].(int64) > 0, nil
			}

			return false, nil
		})

	if err != nil {
		return false, err
	}
	// Ako veza ne postoji, korisnik ne prati autora
	return boolResult.(bool), nil
}

func (fr *FollowRepo) convertDataToProfileSlice(data any) model.Profiles {
	var profiles model.Profiles
	if data == nil {
		return profiles
	}
	list := data.([]interface{})
	if len(list) == 0 {
		return profiles
	}
	for _, prof := range list {
		if prof == nil {
			continue
		}
		profileProps := prof.(map[string]interface{})

		profiles = append(profiles, &model.Profile{
			ID:             profileProps["id"].(int64),
			FirstName:      profileProps["firstName"].(string),
			LastName:       profileProps["lastName"].(string),
			ProfilePicture: profileProps["profilePicture"].(string),
			UserID:         profileProps["userID"].(int64),
		})
	}
	return profiles
}
