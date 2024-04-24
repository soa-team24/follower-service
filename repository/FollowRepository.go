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
	ctx := context.Background()
	session := mr.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	// ExecuteRead for read transactions (Read and queries)
	profileResults, err := session.ExecuteRead(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx,
				`MATCH (profile:Profile)
				RETURN profile.id as profileID, profile.firstName as firstName, profile.lastName as lastName, 
				profile.profilePicture as profilePicture, profile.userID as userID `,
				map[string]any{})
			if err != nil {
				return nil, err
			}

			// Option 1: we iterate over result while there are records
			var profiles model.Profiles
			for result.Next(ctx) {
				record := result.Record()

				ID, _ := record.Get("id")
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
				"CREATE (p:Profile) SET p.id = $id, p.firstName = $firstName, p.lastName = $lastName, p.profilePicture = $profilePicture, p.userID = $userID,  RETURN p.firstName + p.lastName + p.profilePicture + p.userID",
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

/*
func (mr *FollowRepo) GetAllNodesWithFollowLabel() (model.Follows, error) {
	ctx := context.Background()
	session := mr.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	// ExecuteRead for read transactions (Read and queries)
	followResults, err := session.ExecuteRead(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx,
				`MATCH (follow:Follow)
				RETURN follow.profileID as profileID, follow.followerID as followerID`,
				map[string]any{})
			if err != nil {
				return nil, err
			}

			// Option 1: we iterate over result while there are records
			var follows model.Follows
			for result.Next(ctx) {
				record := result.Record()

				profileID, _ := record.Get("profileID")
				followerID, _ := record.Get("followerID")
				follows = append(follows, &model.Follow{
					ProfileID:  profileID.(uint32),
					FollowerID: followerID.(uint32),
				})
			}
			return follows, nil
			// Option 2: we collect all records from result and iterate and map outside of the transaction
			// return result.Collect(ctx)
		})
	if err != nil {
		mr.logger.Println("Error querying search:", err)
		return nil, err
	}
	return followResults.(model.Follows), nil
}

// DOBAVLJANJE PROFILA KOJI GA PRATE
func (fr *FollowRepo) GetAllFollowers(profileID uint32, limit int) (model.Follows, error) {
	ctx := context.Background()
	session := fr.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	followResults, err := session.ExecuteRead(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx,
				`MATCH (follower:Follow {profileID: $profileID})
				RETURN follower.followerID as followerID`,
				map[string]interface{}{"profileID": profileID})
			if err != nil {
				fr.logger.Printf("Error executing query: %v\n", err)
				return nil, err
			}

			var followers model.Follows
			for result.Next(ctx) {
				record := result.Record()
				followerID, ok := record.Get("followerID")
				if !ok {
					fr.logger.Println("Error getting followerID from record")
					continue
				}
				followers = append(followers, &model.Follow{
					ProfileID:  profileID,
					FollowerID: followerID.(uint32),
				})
			}

			if err := result.Err(); err != nil {
				fr.logger.Printf("Error iterating result: %v\n", err)
				return nil, err
			}

			return followers, nil
		})
	if err != nil {
		fr.logger.Println("Error querying search:", err)
		return nil, err
	}
	return followResults.(model.Follows), nil
}
*/
/*
func (fr *FollowRepo) AddFollow(followDto Follow) error {
    // Kreiraj novi follow
    follow, err := NewFollow(followDto.ProfileId, followDto.FollowerId)
    if err != nil {
        return err
    }

    // Dohvati profil na osnovu ID-ja
    profile, err := fr.GetProfile(int(followDto.ProfileId))
    if err != nil {
        return err
    }

    // Dodaj follow u profil
    err = profile.AddFollow(follow)
    if err != nil {
        return err
    }

    // Ažuriraj profil u repozitorijumu
    err = fr.UpdateProfile(profile)
    if err != nil {
        return err
    }

    return nil
}

func (fr *FollowRepo) GetProfile(profileID int) (*model.Profile, error) {
    // Napravi HTTP zahtev ka monolitnoj aplikaciji
    resp, err := http.Get(fmt.Sprintf("http://monolith-api/profiles/%d", profileID))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    // Proveri da li je status kod uspešan
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("failed to get profile, status code: %d", resp.StatusCode)
    }

    // Dekoduj JSON odgovor u strukturu Profile
    var profile model.Profile
    err = json.NewDecoder(resp.Body).Decode(&profile)
    if err != nil {
        return nil, err
    }

    return &profile, nil
}

// UpdateProfile ažurira profil u monolitnoj aplikaciji
func (fr *FollowRepo) UpdateProfile(profile *model.Profile) error {
    // Pretvori profil u JSON
    profileJSON, err := json.Marshal(profile)
    if err != nil {
        return err
    }

    // Napravi HTTP zahtev za ažuriranje profila
    req, err := http.NewRequest(http.MethodPut, "http://monolith-api/profiles", bytes.NewBuffer(profileJSON))
    if err != nil {
        return err
    }
    req.Header.Set("Content-Type", "application/json")

    // Izvrši zahtev
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // Proveri da li je status kod uspešan
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("failed to update profile, status code: %d", resp.StatusCode)
    }

    return nil
}*/
