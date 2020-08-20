package aws

import (
	"encoding/json"
	"errors"
	"fmt"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
	"gitlab.com/grchive/grchive-v3/shared/utility/mt"
	"net/url"
	"sync"
	"time"
)

type awsIamUser struct {
	UserName   string
	CreateDate time.Time
}

type awsIamPolicy struct {
	PolicyName string
	PolicyArn  string

	Type     string
	Inline   bool
	InlineId string
}

type awsIamPolicyStatement struct {
	Effect string
	// Allow and Resource can either be a string or an array of strings
	Action   interface{}
	Resource interface{}
}

type awsIamPolicyDocument struct {
	Version   string
	Statement []awsIamPolicyStatement
}

type awsIamGroup struct {
	Path      string
	GroupName string
	GroupId   string
	Arn       string
}

func createEtlRoleFromAwsPolicy(policy *awsIamPolicy, document *awsIamPolicyDocument) *types.EtlRole {
	role := types.EtlRole{
		Name:        policy.PolicyName,
		Permissions: map[string][]string{},
	}

	for _, st := range document.Statement {
		// TODO: Figure out how to handle the deny elegantly?
		if st.Effect != "Allow" {
			continue
		}
		addPermissionToResource := func(res string, allow interface{}) {
			arr, ok := role.Permissions[res]
			if !ok {
				arr = []string{}
			}

			switch perm := allow.(type) {
			case string:
				arr = append(arr, perm)
			case []interface{}:
				for _, p := range perm {
					arr = append(arr, p.(string))
				}
			default:
				// ????
			}

			role.Permissions[res] = arr
		}

		switch res := st.Resource.(type) {
		case string:
			addPermissionToResource(res, st.Action)
		case []interface{}:
			for _, r := range res {
				addPermissionToResource(r.(string), st.Action)
			}
		default:
			// ????
		}
	}

	return &role
}

type userPolicy struct {
	User     *types.EtlUser
	Policies []*awsIamPolicy
}

func (u *awsIamUser) toEtlUser() *types.EtlUser {
	cloneTime := u.CreateDate
	return &types.EtlUser{
		Username:    u.UserName,
		CreatedTime: &cloneTime,
		Roles:       map[string]*types.EtlRole{},
	}
}

type EtlAWSConnectorUser struct {
	opts *EtlAWSOptions
}

func createAWSConnectorUser(opts *EtlAWSOptions) (*EtlAWSConnectorUser, error) {
	return &EtlAWSConnectorUser{
		opts: opts,
	}, nil
}

func (c *EtlAWSConnectorUser) getInlineUserPolicies(username string) ([]*awsIamPolicy, *connectors.EtlSourceInfo, error) {
	type ResponseBody struct {
		ListUserPoliciesResult struct {
			IsTruncated bool
			Marker      string
			PolicyNames struct {
				Member []string `xml:"member"`
			}
		}
	}

	endpoint := fmt.Sprintf("%s/?Action=ListUserPolicies&Version=2010-05-08&MaxItems=1000&UserName=%s", iamBaseUrl, username)
	pages := []ResponseBody{}
	source, err := awsPaginatedGet(c.opts.Client, "ListUserPoliciesResult", endpoint, &pages)
	if err != nil {
		return nil, nil, err
	}

	policies := []*awsIamPolicy{}
	for _, p := range pages {
		for _, m := range p.ListUserPoliciesResult.PolicyNames.Member {
			policies = append(policies, &awsIamPolicy{
				PolicyName: m,
				Type:       "User",
				Inline:     true,
				InlineId:   username,
			})
		}
	}

	return policies, source, nil
}

func (c *EtlAWSConnectorUser) getAttachedUserPolicies(username string) ([]*awsIamPolicy, *connectors.EtlSourceInfo, error) {
	type ResponseBody struct {
		ListAttachedUserPoliciesResult struct {
			IsTruncated      bool
			Marker           string
			AttachedPolicies struct {
				Member []awsIamPolicy `xml:"member"`
			}
		}
	}

	endpoint := fmt.Sprintf("%s/?Action=ListAttachedUserPolicies&Version=2010-05-08&MaxItems=1000&UserName=%s", iamBaseUrl, username)
	pages := []ResponseBody{}
	source, err := awsPaginatedGet(c.opts.Client, "ListAttachedUserPoliciesResult", endpoint, &pages)
	if err != nil {
		return nil, nil, err
	}

	policies := []*awsIamPolicy{}
	for _, p := range pages {
		for _, m := range p.ListAttachedUserPoliciesResult.AttachedPolicies.Member {
			policies = append(policies, &awsIamPolicy{
				PolicyName: m.PolicyName,
				PolicyArn:  m.PolicyArn,
				Type:       "User",
				Inline:     false,
			})
		}
	}

	return policies, source, nil
}

func (c *EtlAWSConnectorUser) getInlineGroupPolicies(groupName string) ([]*awsIamPolicy, *connectors.EtlSourceInfo, error) {
	type ResponseBody struct {
		ListGroupPoliciesResult struct {
			IsTruncated bool
			Marker      string
			PolicyNames struct {
				Member []string `xml:"member"`
			}
		}
	}

	endpoint := fmt.Sprintf("%s/?Action=ListGroupPolicies&Version=2010-05-08&MaxItems=1000&GroupName=%s", iamBaseUrl, groupName)
	pages := []ResponseBody{}
	source, err := awsPaginatedGet(c.opts.Client, "ListGroupPoliciesResult", endpoint, &pages)
	if err != nil {
		return nil, nil, err
	}

	policies := []*awsIamPolicy{}
	for _, p := range pages {
		for _, m := range p.ListGroupPoliciesResult.PolicyNames.Member {
			policies = append(policies, &awsIamPolicy{
				PolicyName: m,
				Type:       "Group",
				Inline:     true,
				InlineId:   groupName,
			})
		}
	}

	return policies, source, nil

}

func (c *EtlAWSConnectorUser) getAttachedGroupPolicies(groupName string) ([]*awsIamPolicy, *connectors.EtlSourceInfo, error) {
	type ResponseBody struct {
		ListAttachedGroupPoliciesResult struct {
			IsTruncated      bool
			Marker           string
			AttachedPolicies struct {
				Member []awsIamPolicy `xml:"member"`
			}
		}
	}

	endpoint := fmt.Sprintf("%s/?Action=ListAttachedGroupPolicies&Version=2010-05-08&MaxItems=1000&GroupName=%s", iamBaseUrl, groupName)
	pages := []ResponseBody{}
	source, err := awsPaginatedGet(c.opts.Client, "ListAttachedGroupPoliciesResult", endpoint, &pages)
	if err != nil {
		return nil, nil, err
	}

	policies := []*awsIamPolicy{}
	for _, p := range pages {
		for _, m := range p.ListAttachedGroupPoliciesResult.AttachedPolicies.Member {
			policies = append(policies, &awsIamPolicy{
				PolicyName: m.PolicyName,
				PolicyArn:  m.PolicyArn,
				Type:       "Group",
				Inline:     false,
			})
		}
	}

	return policies, source, nil
}

func (c *EtlAWSConnectorUser) getUserGroups(username string) ([]*awsIamGroup, *connectors.EtlSourceInfo, error) {
	type ResponseBody struct {
		ListGroupsForUserResult struct {
			IsTruncated bool
			Marker      string
			Groups      struct {
				Member []awsIamGroup `xml:"member"`
			}
		}
	}

	endpoint := fmt.Sprintf("%s/?Action=ListGroupsForUser&Version=2010-05-08&MaxItems=1000&UserName=%s", iamBaseUrl, username)
	groupPages := []ResponseBody{}
	source, err := awsPaginatedGet(c.opts.Client, "ListGroupsForUserResult", endpoint, &groupPages)
	if err != nil {
		return nil, nil, err
	}

	retGroups := []*awsIamGroup{}
	for _, page := range groupPages {
		for _, m := range page.ListGroupsForUserResult.Groups.Member {
			retGroups = append(retGroups, &m)
		}
	}

	return retGroups, source, nil
}

func (c *EtlAWSConnectorUser) getGroupPolicies(groupName string) ([]*awsIamPolicy, *connectors.EtlSourceInfo, error) {
	finalPolicies := []*awsIamPolicy{}
	finalSource := connectors.CreateSourceInfo()

	inlinePolicies, inlineSource, err := c.getInlineGroupPolicies(groupName)
	if err != nil {
		return nil, nil, err
	}
	finalPolicies = append(finalPolicies, inlinePolicies...)
	finalSource.MergeWith(inlineSource)

	attachedPolicies, attachedSource, err := c.getAttachedGroupPolicies(groupName)
	if err != nil {
		return nil, nil, err
	}
	finalPolicies = append(finalPolicies, attachedPolicies...)
	finalSource.MergeWith(attachedSource)

	return finalPolicies, finalSource, nil
}

func (c *EtlAWSConnectorUser) getUserPolicies(username string) ([]*awsIamPolicy, *connectors.EtlSourceInfo, error) {
	finalPolicies := []*awsIamPolicy{}
	finalSource := connectors.CreateSourceInfo()

	inlinePolicies, inlineSource, err := c.getInlineUserPolicies(username)
	if err != nil {
		return nil, nil, err
	}
	finalPolicies = append(finalPolicies, inlinePolicies...)
	finalSource.MergeWith(inlineSource)

	attachedPolicies, attachedSource, err := c.getAttachedUserPolicies(username)
	if err != nil {
		return nil, nil, err
	}
	finalPolicies = append(finalPolicies, attachedPolicies...)
	finalSource.MergeWith(attachedSource)

	groups, groupSource, err := c.getUserGroups(username)
	if err != nil {
		return nil, nil, err
	}
	finalSource.MergeWith(groupSource)

	for _, g := range groups {
		groupPolicies, groupPolSource, err := c.getGroupPolicies(g.GroupName)
		if err != nil {
			return nil, nil, err
		}

		finalPolicies = append(finalPolicies, groupPolicies...)
		finalSource.MergeWith(groupPolSource)
	}
	return finalPolicies, finalSource, nil
}

type awsGetUserPolicyJob struct {
	// Input
	User      *types.EtlUser
	Connector *EtlAWSConnectorUser

	// Output
	AllPolicies     *sync.Map
	PerUserPolicies map[string]*userPolicy
	Commands        chan *connectors.EtlSourceInfo
}

func (j *awsGetUserPolicyJob) Do() error {
	policies, source, err := j.Connector.getUserPolicies(j.User.Username)
	if err == nil {
		up := userPolicy{
			User:     j.User,
			Policies: policies,
		}
		j.PerUserPolicies[j.User.Username] = &up

		for _, p := range policies {
			j.AllPolicies.Store(p.PolicyName, p)
		}
		j.Commands <- source
	}
	return err
}

type awsGetPolicyDocumentJob struct {
	// Input
	Policy    *awsIamPolicy
	Connector *EtlAWSConnectorUser

	// Output
	PolicyDocuments *sync.Map
	Commands        chan *connectors.EtlSourceInfo
}

func (c *EtlAWSConnectorUser) getAwsInlineUserPolicyDocument(policy *awsIamPolicy) (*awsIamPolicyDocument, *connectors.EtlSourceInfo, error) {
	type ResponseBody struct {
		GetUserPolicyResult struct {
			PolicyDocument string
		}
	}

	endpoint := fmt.Sprintf("%s/?Action=GetUserPolicy&Version=2010-05-08&UserName=%s&PolicyName=%s", iamBaseUrl, policy.InlineId, policy.PolicyName)
	body := ResponseBody{}
	source, err := awsGet(c.opts.Client, endpoint, &body)
	if err != nil {
		return nil, nil, err
	}

	rawDocument, err := url.PathUnescape(body.GetUserPolicyResult.PolicyDocument)
	if err != nil {
		return nil, nil, err
	}

	doc := awsIamPolicyDocument{}
	err = json.Unmarshal([]byte(rawDocument), &doc)
	if err != nil {
		return nil, nil, err
	}

	return &doc, source, nil
}

func (c *EtlAWSConnectorUser) getAwsInlineGroupPolicyDocument(policy *awsIamPolicy) (*awsIamPolicyDocument, *connectors.EtlSourceInfo, error) {
	type ResponseBody struct {
		GetGroupPolicyResult struct {
			PolicyDocument string
		}
	}

	endpoint := fmt.Sprintf("%s/?Action=GetGroupPolicy&Version=2010-05-08&GroupName=%s&PolicyName=%s", iamBaseUrl, policy.InlineId, policy.PolicyName)
	body := ResponseBody{}
	source, err := awsGet(c.opts.Client, endpoint, &body)
	if err != nil {
		return nil, nil, err
	}

	rawDocument, err := url.PathUnescape(body.GetGroupPolicyResult.PolicyDocument)
	if err != nil {
		return nil, nil, err
	}

	doc := awsIamPolicyDocument{}
	err = json.Unmarshal([]byte(rawDocument), &doc)
	if err != nil {
		return nil, nil, err
	}

	return &doc, source, nil
}

func (c *EtlAWSConnectorUser) getAwsAttachedPolicyDocument(policy *awsIamPolicy) (*awsIamPolicyDocument, *connectors.EtlSourceInfo, error) {
	// First get the policy to figure out the default version.
	finalSource := connectors.CreateSourceInfo()
	version := ""
	{
		type GetPolicyResponse struct {
			GetPolicyResult struct {
				Policy struct {
					DefaultVersionId string
				}
			}
		}

		endpoint := fmt.Sprintf("%s/?Action=GetPolicy&Version=2010-05-08&PolicyArn=%s", iamBaseUrl, policy.PolicyArn)
		body := GetPolicyResponse{}
		source, err := awsGet(c.opts.Client, endpoint, &body)
		if err != nil {
			return nil, nil, err
		}

		version = body.GetPolicyResult.Policy.DefaultVersionId
		finalSource.MergeWith(source)
	}

	// Next, retrieve the policy document for the default version.
	type GetPolicyVersionResponse struct {
		GetPolicyVersionResult struct {
			PolicyVersion struct {
				Document string
			}
		}
	}

	endpoint := fmt.Sprintf("%s/?Action=GetPolicyVersion&Version=2010-05-08&PolicyArn=%s&VersionId=%s", iamBaseUrl, policy.PolicyArn, version)
	body := GetPolicyVersionResponse{}
	source, err := awsGet(c.opts.Client, endpoint, &body)
	if err != nil {
		return nil, nil, err
	}

	rawDocument, err := url.PathUnescape(body.GetPolicyVersionResult.PolicyVersion.Document)
	if err != nil {
		return nil, nil, err
	}

	doc := awsIamPolicyDocument{}
	err = json.Unmarshal([]byte(rawDocument), &doc)
	if err != nil {
		return nil, nil, err
	}

	finalSource.MergeWith(source)
	return &doc, finalSource, nil
}

func (c *EtlAWSConnectorUser) getAwsPolicyDocument(policy *awsIamPolicy) (*awsIamPolicyDocument, *connectors.EtlSourceInfo, error) {
	switch policy.Type {
	case "User":
		if policy.Inline {
			return c.getAwsInlineUserPolicyDocument(policy)
		} else {
			return c.getAwsAttachedPolicyDocument(policy)
		}
	case "Group":
		if policy.Inline {
			return c.getAwsInlineGroupPolicyDocument(policy)
		} else {
			return c.getAwsAttachedPolicyDocument(policy)
		}
	default:
		return nil, nil, errors.New("Unsupported policy type.")
	}
}

func (j *awsGetPolicyDocumentJob) Do() error {
	doc, source, err := j.Connector.getAwsPolicyDocument(j.Policy)
	if err == nil {
		j.PolicyDocuments.Store(j.Policy.PolicyName, doc)
		j.Commands <- source
	}
	return err
}

func (c *EtlAWSConnectorUser) GetUserListing() ([]*types.EtlUser, *connectors.EtlSourceInfo, error) {
	type ResponseBody struct {
		ListUsersResult struct {
			IsTruncated bool
			Marker      string
			Users       struct {
				Member []awsIamUser `xml:"member"`
			}
		}
	}
	endpoint := fmt.Sprintf("%s/?Action=ListUsers&Version=2010-05-08&MaxItems=1000", iamBaseUrl)
	retPages := []ResponseBody{}
	source, err := awsPaginatedGet(c.opts.Client, "ListUsersResult", endpoint, &retPages)
	if err != nil {
		return nil, nil, err
	}

	retUsers := []*types.EtlUser{}
	for _, page := range retPages {
		for _, m := range page.ListUsersResult.Users.Member {
			retUsers = append(retUsers, m.toEtlUser())
		}
	}

	sourcesToMerge := make(chan *connectors.EtlSourceInfo)
	go func(source *connectors.EtlSourceInfo, input chan *connectors.EtlSourceInfo) {
		for s := range input {
			source.MergeWith(s)
		}
	}(source, sourcesToMerge)

	// Now that we have every user extracted from the IAM we now need to extract
	// policies and the associated permissions.
	allPolicies := sync.Map{}
	perUserPolicies := map[string]*userPolicy{}
	{
		policyTaskPool := mt.NewTaskPool(10)
		for _, u := range retUsers {
			policyTaskPool.AddJob(&awsGetUserPolicyJob{
				User:            u,
				Connector:       c,
				AllPolicies:     &allPolicies,
				PerUserPolicies: perUserPolicies,
				Commands:        sourcesToMerge,
			})
		}

		err := policyTaskPool.SyncExecute()
		if err != nil {
			return nil, nil, err
		}
	}

	// Next we need to get the associated policy document for each policy we found.
	policyDocuments := sync.Map{}
	{
		documentTaskPool := mt.NewTaskPool(10)
		allPolicies.Range(func(_ interface{}, value interface{}) bool {
			documentTaskPool.AddJob(&awsGetPolicyDocumentJob{
				Policy:          value.(*awsIamPolicy),
				Connector:       c,
				PolicyDocuments: &policyDocuments,
				Commands:        sourcesToMerge,
			})
			return true
		})

		err := documentTaskPool.SyncExecute()
		if err != nil {
			return nil, nil, err
		}
	}

	// Finally, convert the AWS policies to our abstraction of roles and permissions.
	policyToRole := map[string]*types.EtlRole{}
	allPolicies.Range(func(key interface{}, value interface{}) bool {
		documentInt, ok := policyDocuments.Load(key)
		// Still continue -- ideally in the future we want to handle this error somehow?
		if !ok {
			return true
		}
		document := documentInt.(*awsIamPolicyDocument)
		policy := value.(*awsIamPolicy)
		policyToRole[policy.PolicyName] = createEtlRoleFromAwsPolicy(policy, document)
		return true
	})

	// Now go back through the users and associate the created role per policy.
	for _, obj := range perUserPolicies {
		for _, policy := range obj.Policies {
			etlRole := policyToRole[policy.PolicyName]
			obj.User.Roles[etlRole.Name] = etlRole
		}
	}

	close(sourcesToMerge)
	return retUsers, source, nil
}
