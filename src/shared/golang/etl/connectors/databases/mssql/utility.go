package mssql

import (
	"encoding/base64"
)

func SidToString(sid []byte) string {
	return base64.RawStdEncoding.EncodeToString(sid)
}

type principalTreeNode struct {
	principal mssqlPrincipal

	// This principal is a part of these principals
	parentNodes []*principalTreeNode
}

type PrincipalTree struct {
	pidToNode map[int32]*principalTreeNode
}

func ConstructPrincipalTree(principals []mssqlPrincipal, members []mssqlRoleMember) *PrincipalTree {
	pidToNode := map[int32]*principalTreeNode{}
	for _, p := range principals {
		pidToNode[p.PrincipalId] = &principalTreeNode{
			principal: p,
		}
	}

	for _, m := range members {
		roleNode := pidToNode[m.RolePrincipalId]
		memberNode := pidToNode[m.MemberPrincipalId]
		memberNode.parentNodes = append(memberNode.parentNodes, roleNode)
	}

	return &PrincipalTree{
		pidToNode: pidToNode,
	}
}

func (t *PrincipalTree) FindParentPrincipalsFromId(pid int32) []mssqlPrincipal {
	node, ok := t.pidToNode[pid]
	if !ok {
		return []mssqlPrincipal{}
	}

	nodeQueue := node.parentNodes
	principals := []mssqlPrincipal{}

	for len(nodeQueue) > 0 {
		currentNode := nodeQueue[0]
		principals = append(principals, currentNode.principal)
		nodeQueue = append(nodeQueue[1:], currentNode.parentNodes...)
	}

	return principals
}
