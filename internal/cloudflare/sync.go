package cloudflare

import (
	"context"
	log "google-cloudflare-sync/internal/log"
	"os"

	cf "github.com/cloudflare/cloudflare-go"
)

type Cloudflare_cli struct {
	ctx        context.Context
	api        *cf.API
	TeamListID string
	accountID  string
}

func NewCloudflareClient() *Cloudflare_cli {
	if os.Getenv("CF_API_KEY") == "" || os.Getenv("CF_API_EMAIL") == "" || os.Getenv("CF_API_ACCOUNTID") == "" {
		log.SharedLogger.Fatal("CF_API_KEY, CF_API_EMAIL, CF_API_ACCOUNTID env variables are required")
	}
	api, err := cf.New(os.Getenv("CF_API_KEY"), os.Getenv("CF_API_EMAIL"))
	accountID := os.Getenv("CF_API_ACCOUNTID")
	if err != nil {
		log.SharedLogger.Fatal(err)
	}
	return &Cloudflare_cli{api: api, accountID: accountID, ctx: context.Background()}
}

func (cli *Cloudflare_cli) CreateTeamList(targetList string) string {
	teamsList, _, err := cli.api.ListTeamsLists(cli.ctx, cf.AccountIdentifier(cli.accountID), cf.ListTeamListsParams{})
	if err != nil {
		log.SharedLogger.Fatal(err)
	}
	// if list already exists, return
	for _, team := range teamsList {
		if team.Name == targetList {
			log.SharedLogger.Debugf("Team List %s already exists", targetList)
			return team.ID
		}
	}
	// create list
	params := cf.CreateTeamsListParams{
		Name: targetList,
		Type: "EMAIL",
	}
	newTeamsList, err := cli.api.CreateTeamsList(cli.ctx, cf.AccountIdentifier(cli.accountID), params)
	if err != nil {
		log.SharedLogger.Fatal(err)
	}
	log.SharedLogger.Debugf("Created Team List: %s, ID: %s\n", newTeamsList.Name, newTeamsList.ID)
	return newTeamsList.ID
}

func (cli *Cloudflare_cli) GetUsers(targetList string) map[string]string {
	userList := make(map[string]string)
	teamsList, _, err := cli.api.ListTeamsLists(cli.ctx, cf.AccountIdentifier(cli.accountID), cf.ListTeamListsParams{})
	if err != nil {
		log.SharedLogger.Fatal(err)
	}
	for _, team := range teamsList {
		log.SharedLogger.Debugf("Team: %s, ID: %s\n", team.Name, team.ID)
		if team.Name == targetList {
			log.SharedLogger.Debugf("Found Team: %s, ID: %s\n", team.Name, team.ID)
			cli.TeamListID = team.ID
			listUsers, _, err := cli.api.ListTeamsListItems(cli.ctx, cf.AccountIdentifier(cli.accountID), cf.ListTeamsListItemsParams{ListID: team.ID})
			if err != nil {
				log.SharedLogger.Fatal(err)
			}
			for _, list_user := range listUsers {
				log.SharedLogger.Debugf("[CF]Email: %s\n", list_user.Value)
				userList[list_user.Value] = team.ID
			}
			break
		}
	}
	return userList
}

func (cli *Cloudflare_cli) GetAllUsers() map[string]string {
	userList := make(map[string]string)
	cfUsers, _, err := cli.api.ListAccessUsers(cli.ctx, cf.AccountIdentifier(cli.accountID), cf.AccessUserParams{})
	if err != nil {
		log.SharedLogger.Fatal(err)
	}
	log.SharedLogger.Debugf("Users: %v", cfUsers)
	for _, user := range cfUsers {
		if *user.AccessSeat || *user.GatewaySeat {
			log.SharedLogger.Debugf("[CF]Active user found %s: Gate: %s Access: %s", user.Email, *user.GatewaySeat, *user.AccessSeat)
			userList[user.Email] = user.ID
		}
	}
	return userList
}

func (cli Cloudflare_cli) DeleteUsers(cfUsersToDelete map[string]string) {
	for email := range cfUsersToDelete {
		// Delete users with pagination
		page := 1
		for {
			accessUsers, resultInfo, err := cli.api.ListAccessUsers(cli.ctx, cf.AccountIdentifier(cli.accountID), cf.AccessUserParams{})
			if err != nil {
				log.SharedLogger.Fatal(err)
			}
			for _, user := range accessUsers {
				if user.Email == email {
					log.SharedLogger.Debugf("Found User %s id: %s in CloudFlare", email, user.ID)
					accessSeat := false
					gatewaySeat := false
					params := cf.UpdateAccessUserSeatParams{SeatUID: user.SeatUID, AccessSeat: &accessSeat, GatewaySeat: &gatewaySeat}
					seat, err := cli.api.UpdateAccessUserSeat(cli.ctx, cf.AccountIdentifier(cli.accountID), params)
					if err != nil {
						log.SharedLogger.Fatal(err)
					}
					log.SharedLogger.Infof("Deleted User %s from CloudFlare", email)
					log.SharedLogger.Debugf("Deleted User respose: %v ", seat[0])
				}
			}

			// Check if there are more pages
			if resultInfo.Page == resultInfo.TotalPages {
				break
			}
			page++
		}
	}
}

func (cli *Cloudflare_cli) PatchUsers(googleUsersToAdd map[string]string, cfUsersToDelete map[string]string, cf_list_name string) {
	//Create list of TeamListItems to add from googleUsersToAdd
	teamListItemsToAdd := []cf.TeamsListItem{}
	for email := range googleUsersToAdd {
		teamListItemsToAdd = append(teamListItemsToAdd, cf.TeamsListItem{Value: email})
		log.SharedLogger.Infof("ADD User %s to CloudFlare Team List %s", email, cf_list_name)
	}

	//create list of TeamListItems to delete from cfUsersToDelete
	teamListItemsToDelete := []string{}
	for email := range cfUsersToDelete {
		teamListItemsToDelete = append(teamListItemsToDelete, email)
		log.SharedLogger.Infof("DELETE: User %s to CloudFlare Team List %s", email, cf_list_name)
	}

	teamListParams := cf.PatchTeamsListParams{
		ID:     cli.TeamListID,
		Append: teamListItemsToAdd,
		Remove: teamListItemsToDelete,
	}

	log.SharedLogger.Debugf("PatchTeamsListParams: %v", teamListParams)
	resultingTeamList, err := cli.api.PatchTeamsList(cli.ctx, cf.AccountIdentifier(cli.accountID), teamListParams)
	if err != nil {
		log.SharedLogger.Fatal(err)
	}
	log.SharedLogger.Debugf("Resulting TeamList: %v", resultingTeamList.Items)

}
