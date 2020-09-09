package oracle

import (
	"fmt"
)

type UserRoleNode struct {
	ParentNodes []*UserRoleNode

	User *oracleUser
	Role *oracleRole
}

type UserRoleTree struct {
	userNodes map[string]*UserRoleNode
}

func CreateUserRoleTree(users []oracleUser, roles []oracleRole, rolePriv []oracleRolePriv) *UserRoleTree {
	userNodes := map[string]*UserRoleNode{}
	for idx, u := range users {
		userNodes[u.Username] = &UserRoleNode{
			ParentNodes: []*UserRoleNode{},
			User:        &users[idx],
		}
	}

	roleNodes := map[string]*UserRoleNode{}
	for idx, r := range roles {
		roleNodes[r.Role] = &UserRoleNode{
			ParentNodes: []*UserRoleNode{},
			Role:        &roles[idx],
		}
	}

	for _, priv := range rolePriv {
		var granteeNode *UserRoleNode
		var ok bool

		granteeNode, ok = userNodes[priv.Grantee]
		if !ok {
			granteeNode, ok = roleNodes[priv.Grantee]
			if !ok {
				fmt.Printf("Failed to find grantee: %s\n", priv.Grantee)
				continue
			}
		}

		roleNode, ok := roleNodes[priv.GrantedRole]
		if !ok {
			fmt.Printf("Failed to find role: %s\n", priv.GrantedRole)
			continue
		}

		granteeNode.ParentNodes = append(granteeNode.ParentNodes, roleNode)
	}

	return &UserRoleTree{
		userNodes: userNodes,
	}
}

func (t *UserRoleTree) FindUserParentRoleNames(username string) []string {
	userNode, ok := t.userNodes[username]
	if !ok {
		return []string{}
	}

	retRoles := []string{}
	nodeQueue := userNode.ParentNodes
	for len(nodeQueue) > 0 {
		node := nodeQueue[0]

		if node.Role != nil {
			retRoles = append(retRoles, node.Role.Role)
		}

		nodeQueue = append(nodeQueue[1:], node.ParentNodes...)
	}

	return retRoles
}
