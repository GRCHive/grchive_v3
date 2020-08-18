package github

import (
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
	"gitlab.com/grchive/grchive-v3/shared/utility/graphql"
	"time"
)

type EtlGithubConnectorUser struct {
	opts *EtlGithubOptions
}

func createGithubConnectorUser(opts *EtlGithubOptions) (*EtlGithubConnectorUser, error) {
	return &EtlGithubConnectorUser{
		opts: opts,
	}, nil
}

func (c *EtlGithubConnectorUser) GetUserListing() ([]*types.EtlUser, *connectors.EtlSourceInfo, error) {
	retUsers := []*types.EtlUser{}
	source := connectors.CreateSourceInfo()

	uniqueUsers := map[string]bool{}

	var afterCursor interface{}
	afterCursor = nil

	gqlQuery := `
	query($org_id:String!, $after_cursor:String) {
			organization(login: $org_id) {
				name
				membersWithRole(first: 100, after: $after_cursor) {
					edges {
						node {
							name
							login
							createdAt
						}
						role
					}
					pageInfo {
						endCursor
						hasNextPage
					}
				}
			}
		}
	`
	type ResponseBody struct {
		Data struct {
			Organization struct {
				Name            string `json:"name"`
				MembersWithRole struct {
					Edges []struct {
						Node struct {
							Name      string    `json:"name"`
							Login     string    `json:"login"`
							CreatedAt time.Time `json:"createdAt"`
						}
						Role string
					} `json:"edges"`
					PageInfo struct {
						EndCursor   string `json:"endCursor"`
						HasNextPage bool   `json:"hasNextPage"`
					} `json:"pageInfo"`
				} `json:"membersWithRole"`
			} `json:"organization"`
		} `json:"data"`
	}

	for {
		respData := ResponseBody{}

		gqlRequest := graphql_utility.GraphQLRequestBody{
			Query: gqlQuery,
			Variables: map[string]interface{}{
				"org_id":       c.opts.OrgId,
				"after_cursor": afterCursor,
			},
		}

		rawResponse, err := graphql_utility.SendGraphQLRequest(
			graphqlEndpoint,
			c.opts.Client,
			gqlRequest,
			&respData)

		if err != nil {
			return nil, nil, err
		}

		added := 0
		for _, u := range respData.Data.Organization.MembersWithRole.Edges {
			if _, ok := uniqueUsers[u.Node.Login]; ok {
				continue
			}

			tm := u.Node.CreatedAt
			etlUser := types.EtlUser{
				Username:    u.Node.Login,
				FullName:    u.Node.Name,
				CreatedTime: &tm,
				Roles: map[string]*types.EtlRole{
					u.Role: &types.EtlRole{
						Name: u.Role,
					},
				},
			}
			retUsers = append(retUsers, &etlUser)
			uniqueUsers[u.Node.Login] = true
			added = added + 1
		}

		// This is just a fallback for if something terrible goes wrong and we want to avoid getting into an infinite loop.
		if added == 0 {
			break
		}

		cmd := connectors.EtlCommandInfo{
			Command:    gqlRequest.Query,
			Parameters: gqlRequest.Variables,
			RawData:    rawResponse,
		}
		source.AddCommand(&cmd)

		if !respData.Data.Organization.MembersWithRole.PageInfo.HasNextPage {
			break
		}

		afterCursor = respData.Data.Organization.MembersWithRole.PageInfo.EndCursor
	}
	return retUsers, source, nil
}
