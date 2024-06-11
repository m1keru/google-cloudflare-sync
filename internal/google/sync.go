package google

import (
	"context"
	log "google-cloudflare-sync/internal/log"
	"os"
	"strings"

	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

type Google_cli struct {
	ctx     context.Context
	service *admin.Service
}

func NewGoogleClient(impersonate_email string) *Google_cli {
	ctx := context.Background()
	if os.Getenv("GOOGLE_DOMAIN") == "" {
		log.SharedLogger.Fatal("GOOGLE_CREDENTIALS, GOOGLE_DOMAIN, env variables are required")
	}
	google_creds := []byte(os.Getenv("GOOGLE_CREDENTIALS"))
	var err error
	if len(google_creds) == 0 {
		google_creds, err = os.ReadFile("google.json")
		if err != nil {
			log.SharedLogger.Fatalf("Unable to read GOOGLE_CREDENTIALS not from ENV  neiher from  google.json file: %v", err)
		}
	}

	config, err := google.JWTConfigFromJSON(google_creds, admin.AdminDirectoryGroupMemberReadonlyScope, admin.AdminDirectoryGroupReadonlyScope)
	if err != nil {
		log.SharedLogger.Fatalf("Unable to parse service account key file to config: %v", err)
	}

	// Specify the user to impersonate
	config.Subject = impersonate_email

	client := config.Client(ctx)
	service, err := admin.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.SharedLogger.Fatalf("Unable to retrieve directory Client %v", err)
	}

	return &Google_cli{ctx: ctx, service: service}
}

func (cli *Google_cli) GetGroupsByRegex(google_groups_regex string) []string {
	// Create a regular expression pattern
	pattern := google_groups_regex

	// List all groups
	call := cli.service.Groups.List().Domain(os.Getenv("GOOGLE_DOMAIN")).Query(pattern)
	response, err := call.Do()
	if err != nil {
		log.SharedLogger.Fatalf("Unable to retrieve groups: %v", err)
	}
	// Extract group names from the response
	var groupNames []string
	for _, group := range response.Groups {
		groupNames = append(groupNames, group.Email)
	}
	return groupNames
}

func (cli *Google_cli) GetUsers(groupKey string) map[string]string {
	groups := strings.Split(groupKey, ",")
	log.SharedLogger.Debugf("Groups: %v", groups)
	// List all users in the group
	retVal := make(map[string]string)
	for _, groupKey := range groups {
		nextPageToken := ""
		for {
			call := cli.service.Members.List(groupKey).PageToken(nextPageToken)
			response, err := call.Do()
			if err != nil {
				if gerr, ok := err.(*googleapi.Error); ok && gerr.Code == 404 {
					log.SharedLogger.Errorf("Group not found: %v", groupKey)
				} else {
					log.SharedLogger.Fatalf("Unable to retrieve group members: %v", err)
				}
				break
			}

			for _, member := range response.Members {
				log.SharedLogger.Debugf("[GOOGLE]Email: %s, Role: %s\n", member.Email, member.Role)
				retVal[member.Email] = member.Role
			}

			nextPageToken = response.NextPageToken
			if nextPageToken == "" {
				break
			}
		}
	}
	return retVal
}
