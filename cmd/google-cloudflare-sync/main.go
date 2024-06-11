package main

import (
	"flag"
	"fmt"
	"google-cloudflare-sync/internal/cloudflare"
	"google-cloudflare-sync/internal/google"
	"google-cloudflare-sync/internal/log"
	"os"
	"regexp"

	"go.uber.org/zap"
)

// AppVersion -release version
const AppVersion = "0.0.2"

func main() {
	version := flag.Bool("version", false, "current version")
	debug := flag.Bool("debug", false, "debug mode")
	delete_stale_users := flag.Bool("delete_stale", false, "delete stale users")
	google_groups := flag.String("google_groups", "", "google groups comma separated")
	google_impesonate_email := flag.String("google_impersonate", "", "google impersonate email")
	google_groups_regex := flag.String("google_groups_regex", "", "google groups regex, example: name:contact*, link https://developers.google.com/admin-sdk/directory/v1/guides/search-groups")
	cf_list_name := flag.String("cf_list_name", "vpn-users", "cloudflare list name, use only with google_groups, google_groups_regex will create name from founded groups")
	flag.Parse()
	if *version {
		fmt.Println(AppVersion)
		os.Exit(0)
	}
	if *google_groups == "" && *google_groups_regex == "" {
		fmt.Println("ERROR: one of google_groups or google_groups_regex is required")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *google_groups != "" && *google_groups_regex != "" {
		fmt.Println("ERROR: ONLY one of google_groups or google_groups_regex is required")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *debug {
		log.InitLogger(zap.DebugLevel)
	} else {
		log.InitLogger(zap.InfoLevel)
	}

	if *google_impesonate_email == "" {
		fmt.Println("ERROR: impersonate email is required")
		flag.PrintDefaults()
		os.Exit(1)
	}

	defer log.SharedLogger.Sync()
	cfCli := cloudflare.NewCloudflareClient()
	googleCli := google.NewGoogleClient(*google_impesonate_email)

	if *delete_stale_users && *google_groups != "" {
		log.SharedLogger.Infof("Deleting stale users")
		google_users := googleCli.GetUsers(*google_groups)
		cfUsers := cfCli.GetAllUsers()
		// find cloudflare users not in google
		cfUsersToDelete := make(map[string]string)
		for email := range cfUsers {
			//if cfUser is not in easybraim domain, skip
			regexp := regexp.MustCompile(`.*@easybrain.com$`)
			if !regexp.MatchString(email) {
				continue
			}
			if _, ok := google_users[email]; !ok {
				log.SharedLogger.Infof("User %s not found in google", email)
				cfUsersToDelete[email] = cfUsers[email]
			}
		}

		cfCli.DeleteUsers(cfUsersToDelete)
		log.SharedLogger.Infof("Done")
		os.Exit(0)
	}

	if *google_groups_regex != "" {
		google_groups_list := googleCli.GetGroupsByRegex(*google_groups_regex)
		log.SharedLogger.Infof("Found groups: %v", google_groups_list)
		for _, group := range google_groups_list {
			log.SharedLogger.Infof("Processing group: %s", group)
			cfCli.CreateTeamList(group)
			google_users := googleCli.GetUsers(group)
			log.SharedLogger.Debugf("Found GOOGLE users: %v", google_users)
			cfUsers := cfCli.GetUsers(group)
			log.SharedLogger.Debugf("Found CF %s users: %v", group, cfUsers)
			// find cloudflare users not in google
			cfUsersToDelete := make(map[string]string)
			for email := range cfUsers {
				log.SharedLogger.Debugf("Checking CF user %s", email)
				if _, ok := google_users[email]; !ok {
					log.SharedLogger.Infof("User %s not found in google", email)
					cfUsersToDelete[email] = cfUsers[email]
				}
			}
			// find users in google not in cloudflare
			googleUsersToAdd := make(map[string]string)
			for email := range google_users {
				if _, ok := cfUsers[email]; !ok {
					log.SharedLogger.Infof("User %s not found in cloudflare", email)
					googleUsersToAdd[email] = google_users[email]
				}
			}
			cfCli.PatchUsers(googleUsersToAdd, cfUsersToDelete, group)
		}
	} else if *google_groups != "" {
		google_users := googleCli.GetUsers(*google_groups)
		cfUsers := cfCli.GetUsers(*cf_list_name)
		// find cloudflare users not in google
		cfUsersToDelete := make(map[string]string)
		for email := range cfUsers {
			if _, ok := google_users[email]; !ok {
				log.SharedLogger.Infof("User %s not found in google", email)
				cfUsersToDelete[email] = cfUsers[email]
			}
		}
		// find users in google not in cloudflare
		googleUsersToAdd := make(map[string]string)
		for email := range google_users {
			if _, ok := cfUsers[email]; !ok {
				log.SharedLogger.Infof("User %s not found in cloudflare", email)
				googleUsersToAdd[email] = google_users[email]
			}
		}
		cfCli.PatchUsers(googleUsersToAdd, cfUsersToDelete, *cf_list_name)
	}
	log.SharedLogger.Infof("Done")

}
