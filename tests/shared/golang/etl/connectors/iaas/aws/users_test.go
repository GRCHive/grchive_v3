package aws

import (
	"fmt"
	"github.com/onsi/gomega"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
	"gitlab.com/grchive/grchive-v3/shared/test_utility"
	"gitlab.com/grchive/grchive-v3/tests/shared/etl/connectors/iaas/aws_utility"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func createStandardClient() *aws_utility.MockAWSClient {
	return &aws_utility.MockAWSClient{
		Iam: aws_utility.MockAWSIamClient{
			ListUsers: func() (*http.Response, error) {
				return test_utility.WrapHttpResponse(`
<ListUsersResponse xmlns="https://iam.amazonaws.com/doc/2010-05-08/">
  <ListUsersResult>
    <IsTruncated>false</IsTruncated>
    <Users>
      <member>
        <Path>/</Path>
        <Arn>arn:aws:iam::455499087251:user/grchive-api</Arn>
        <UserName>grchive-api</UserName>
        <UserId>AIDAWUDO4PWJT6ICINTCV</UserId>
        <CreateDate>2020-08-19T22:06:55Z</CreateDate>
      </member>
      <member>
        <Path>/</Path>
        <PasswordLastUsed>2020-08-19T22:09:03Z</PasswordLastUsed>
        <Arn>arn:aws:iam::455499087251:user/mikebao-iam</Arn>
        <UserName>mikebao-iam</UserName>
        <UserId>AIDAWUDO4PWJ7W25KEKF7</UserId>
        <CreateDate>2020-08-19T22:07:51Z</CreateDate>
      </member>
    </Users>
  </ListUsersResult>
  <ResponseMetadata>
    <RequestId>58c111fc-5e64-4ddd-9d6c-1a29faa73e03</RequestId>
  </ResponseMetadata>
</ListUsersResponse>
`), nil
			},
			ListUserPolicies: map[string]aws_utility.MockAWSFn{
				"mikebao-iam": func() (*http.Response, error) {
					data := fmt.Sprintf(`<ListUserPoliciesResponse>
	<ListUserPoliciesResult>
		<IsTruncated>false</IsTruncated>
		<PolicyNames>
		  <member>TestInlinePolicy</member>
		</PolicyNames>
	  </ListUserPoliciesResult>
	  <ResponseMetadata>
		<RequestId>1c6a71f2-bcdb-4849-9c81-e4e7b1028798</RequestId>
	  </ResponseMetadata>
	</ListUserPoliciesResponse>
	`)
					body := ioutil.NopCloser(strings.NewReader(data))
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       body,
					}, nil
				},
				"grchive-api": func() (*http.Response, error) {
					data := fmt.Sprintf(`<ListUserPoliciesResponse xmlns="https://iam.amazonaws.com/doc/2010-05-08/">
  <ListUserPoliciesResult>
    <IsTruncated>false</IsTruncated>
    <PolicyNames/>
  </ListUserPoliciesResult>
  <ResponseMetadata>
    <RequestId>2f45cd69-7ff5-4f55-bf31-0183298ddb71</RequestId>
  </ResponseMetadata>
</ListUserPoliciesResponse>`)
					body := ioutil.NopCloser(strings.NewReader(data))
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       body,
					}, nil
				},
			},
			ListAttachedUserPolicies: map[string]aws_utility.MockAWSFn{
				"grchive-api": func() (*http.Response, error) {
					return test_utility.WrapHttpResponse(`
<ListAttachedUserPoliciesResponse xmlns="https://iam.amazonaws.com/doc/2010-05-08/">
  <ListAttachedUserPoliciesResult>
    <IsTruncated>false</IsTruncated>
    <AttachedPolicies>
      <member>
        <PolicyArn>arn:aws:iam::455499087251:policy/GRCHiveAPI</PolicyArn>
        <PolicyName>GRCHiveAPI</PolicyName>
      </member>
    </AttachedPolicies>
  </ListAttachedUserPoliciesResult>
  <ResponseMetadata>
    <RequestId>8051cad5-2087-49e2-9130-534054061ccf</RequestId>
  </ResponseMetadata>
</ListAttachedUserPoliciesResponse>
					`), nil
				},
				"mikebao-iam": func() (*http.Response, error) {
					return test_utility.WrapHttpResponse(`
<ListAttachedUserPoliciesResponse xmlns="https://iam.amazonaws.com/doc/2010-05-08/">
  <ListAttachedUserPoliciesResult>
    <IsTruncated>false</IsTruncated>
    <AttachedPolicies>
      <member>
        <PolicyArn>arn:aws:iam::aws:policy/IAMUserChangePassword</PolicyArn>
        <PolicyName>IAMUserChangePassword</PolicyName>
      </member>
    </AttachedPolicies>
  </ListAttachedUserPoliciesResult>
  <ResponseMetadata>
    <RequestId>2826323e-fa86-49c9-b4fb-95e9162cc239</RequestId>
  </ResponseMetadata>
</ListAttachedUserPoliciesResponse>
					`), nil
				},
			},
			ListGroupPolicies: map[string]aws_utility.MockAWSFn{
				"GRCHive": func() (*http.Response, error) {
					return test_utility.WrapHttpResponse(`
<ListGroupPoliciesResponse xmlns="https://iam.amazonaws.com/doc/2010-05-08/">
  <ListGroupPoliciesResult>
    <IsTruncated>false</IsTruncated>
    <PolicyNames>
      <member>policygen-GRCHive-202008201657</member>
    </PolicyNames>
  </ListGroupPoliciesResult>
  <ResponseMetadata>
    <RequestId>644de3c9-cf36-4015-82f5-6e8c6524c969</RequestId>
  </ResponseMetadata>
</ListGroupPoliciesResponse>
					`), nil
				},
			},
			ListAttachedGroupPolicies: map[string]aws_utility.MockAWSFn{
				"GRCHive": func() (*http.Response, error) {
					return test_utility.WrapHttpResponse(`
<ListAttachedGroupPoliciesResponse xmlns="https://iam.amazonaws.com/doc/2010-05-08/">
  <ListAttachedGroupPoliciesResult>
    <IsTruncated>false</IsTruncated>
    <AttachedPolicies>
      <member>
        <PolicyArn>arn:aws:iam::455499087251:policy/GRCHiveAPI</PolicyArn>
        <PolicyName>GRCHiveAPI</PolicyName>
      </member>
    </AttachedPolicies>
  </ListAttachedGroupPoliciesResult>
  <ResponseMetadata>
    <RequestId>43e2be5c-9e16-49cb-b620-0f3f94bdec83</RequestId>
  </ResponseMetadata>
</ListAttachedGroupPoliciesResponse>
					`), nil
				},
			},
			ListGroupsForUser: map[string]aws_utility.MockAWSFn{
				"grchive-api": func() (*http.Response, error) {
					return test_utility.WrapHttpResponse(`
<ListGroupsForUserResponse xmlns="https://iam.amazonaws.com/doc/2010-05-08/">
  <ListGroupsForUserResult>
    <IsTruncated>false</IsTruncated>
    <Groups>
      <member>
        <Path>/</Path>
        <GroupName>GRCHive</GroupName>
        <GroupId>AGPAWUDO4PWJ3VBSHRMBX</GroupId>
        <Arn>arn:aws:iam::455499087251:group/GRCHive</Arn>
        <CreateDate>2020-08-20T20:57:10Z</CreateDate>
      </member>
    </Groups>
  </ListGroupsForUserResult>
  <ResponseMetadata>
    <RequestId>0c9f91e6-0bd4-471f-a45b-469ae5dc2280</RequestId>
  </ResponseMetadata>
</ListGroupsForUserResponse>
					`), nil
				},
				"mikebao-iam": func() (*http.Response, error) {
					return test_utility.WrapHttpResponse(`
<ListGroupsForUserResponse xmlns="https://iam.amazonaws.com/doc/2010-05-08/">
  <ListGroupsForUserResult>
    <IsTruncated>false</IsTruncated>
    <Groups/>
  </ListGroupsForUserResult>
  <ResponseMetadata>
    <RequestId>d41965c1-02a6-4e17-b555-6c7668da8906</RequestId>
  </ResponseMetadata>
</ListGroupsForUserResponse>
					`), nil
				},
			},
			GetUserPolicy: map[string]map[string]aws_utility.MockAWSFn{
				"mikebao-iam": map[string]aws_utility.MockAWSFn{
					"TestInlinePolicy": func() (*http.Response, error) {
						return test_utility.WrapHttpResponse(fmt.Sprintf(`
<GetUserPolicyResponse xmlns="https://iam.amazonaws.com/doc/2010-05-08/">
  <GetUserPolicyResult>
    <PolicyDocument>%s</PolicyDocument>
    <PolicyName>TestInlinePolicy</PolicyName>
    <UserName>mikebao-iam</UserName>
  </GetUserPolicyResult>
  <ResponseMetadata>
    <RequestId>f6cd9598-76f1-46d2-9229-65e57967ef57</RequestId>
  </ResponseMetadata>
</GetUserPolicyResponse>`, url.PathEscape(`
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VisualEditor0",
            "Effect": "Allow",
            "Action": "ec2:CreateImage",
            "Resource": "*"
        }
    ]
}
`))), nil
					},
				},
			},
			GetGroupPolicy: map[string]map[string]aws_utility.MockAWSFn{
				"GRCHive": map[string]aws_utility.MockAWSFn{
					"policygen-GRCHive-202008201657": func() (*http.Response, error) {
						return test_utility.WrapHttpResponse(fmt.Sprintf(`
<GetGroupPolicyResponse xmlns="https://iam.amazonaws.com/doc/2010-05-08/">
  <GetGroupPolicyResult>
    <PolicyDocument>%s</PolicyDocument>
    <GroupName>GRCHive</GroupName>
    <PolicyName>policygen-GRCHive-202008201657</PolicyName>
  </GetGroupPolicyResult>
  <ResponseMetadata>
    <RequestId>c6113564-ab35-490f-bcf7-d07d6f502d29</RequestId>
  </ResponseMetadata>
</GetGroupPolicyResponse>
`, url.PathEscape(`
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "Stmt1597957067000",
            "Effect": "Allow",
            "Action": [
                "waf:GetRule"
            ],
            "Resource": [
                "*"
            ]
        }
    ]
}
`))), nil
					},
				},
			},

			GetPolicy: map[string]aws_utility.MockAWSFn{
				"arn:aws:iam::455499087251:policy/GRCHiveAPI": func() (*http.Response, error) {
					return test_utility.WrapHttpResponse(`
<GetPolicyResponse xmlns="https://iam.amazonaws.com/doc/2010-05-08/">
  <GetPolicyResult>
    <Policy>
      <PermissionsBoundaryUsageCount>0</PermissionsBoundaryUsageCount>
      <Path>/</Path>
      <UpdateDate>2020-08-20T18:22:40Z</UpdateDate>
      <DefaultVersionId>v3</DefaultVersionId>
      <PolicyId>ANPAWUDO4PWJQTL23I4CR</PolicyId>
      <IsAttachable>true</IsAttachable>
      <PolicyName>GRCHiveAPI</PolicyName>
      <AttachmentCount>2</AttachmentCount>
      <Arn>arn:aws:iam::455499087251:policy/GRCHiveAPI</Arn>
      <CreateDate>2020-08-20T14:31:08Z</CreateDate>
    </Policy>
  </GetPolicyResult>
  <ResponseMetadata>
    <RequestId>e6301037-a5fd-47da-838b-ce106d1edfe5</RequestId>
  </ResponseMetadata>
</GetPolicyResponse>
`), nil
				},
				"arn:aws:iam::aws:policy/IAMUserChangePassword": func() (*http.Response, error) {
					return test_utility.WrapHttpResponse(`
<GetPolicyResponse xmlns="https://iam.amazonaws.com/doc/2010-05-08/">
  <GetPolicyResult>
    <Policy>
      <PermissionsBoundaryUsageCount>0</PermissionsBoundaryUsageCount>
      <Path>/</Path>
      <UpdateDate>2016-11-15T23:18:55Z</UpdateDate>
      <DefaultVersionId>v2</DefaultVersionId>
      <PolicyId>ANPAJ4L4MM2A7QIEB56MS</PolicyId>
      <IsAttachable>true</IsAttachable>
      <PolicyName>IAMUserChangePassword</PolicyName>
      <Description>Provides the ability for an IAM user to change their own password.</Description>
      <AttachmentCount>1</AttachmentCount>
      <Arn>arn:aws:iam::aws:policy/IAMUserChangePassword</Arn>
      <CreateDate>2016-11-15T00:25:16Z</CreateDate>
    </Policy>
  </GetPolicyResult>
  <ResponseMetadata>
    <RequestId>2cfe4dc5-4d5b-4247-bbb6-4d6e8df42be6</RequestId>
  </ResponseMetadata>
</GetPolicyResponse>
`), nil
				},
			},

			GetPolicyVersion: map[string]map[string]aws_utility.MockAWSFn{
				"arn:aws:iam::455499087251:policy/GRCHiveAPI": map[string]aws_utility.MockAWSFn{
					"v3": func() (*http.Response, error) {
						return test_utility.WrapHttpResponse(fmt.Sprintf(`
<GetPolicyVersionResponse xmlns="https://iam.amazonaws.com/doc/2010-05-08/">
  <GetPolicyVersionResult>
    <PolicyVersion>
      <VersionId>v3</VersionId>
      <IsDefaultVersion>true</IsDefaultVersion>
      <Document>%s</Document>
      <CreateDate>2020-08-20T18:22:40Z</CreateDate>
    </PolicyVersion>
  </GetPolicyVersionResult>
  <ResponseMetadata>
    <RequestId>986e7fac-7480-44b7-bfc6-ff4be61202e3</RequestId>
  </ResponseMetadata>
</GetPolicyVersionResponse>
	`, url.PathEscape(`
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VisualEditor0",
            "Effect": "Allow",
            "Action": [
                "iam:GetPolicyVersion",
                "iam:GetPolicy",
                "iam:GetUserPolicy",
                "iam:ListGroupsForUser",
                "iam:ListGroupPolicies",
                "iam:ListAttachedUserPolicies",
                "iam:ListUsers",
                "iam:ListAttachedGroupPolicies",
                "iam:ListUserPolicies",
                "iam:GetGroupPolicy"
            ],
            "Resource": "*"
        }
    ]
}
`))), nil
					},
				},
				"arn:aws:iam::aws:policy/IAMUserChangePassword": map[string]aws_utility.MockAWSFn{
					"v2": func() (*http.Response, error) {
						return test_utility.WrapHttpResponse(fmt.Sprintf(`
<GetPolicyVersionResponse xmlns="https://iam.amazonaws.com/doc/2010-05-08/">
  <GetPolicyVersionResult>
    <PolicyVersion>
      <VersionId>v2</VersionId>
      <IsDefaultVersion>true</IsDefaultVersion>
      <Document>%s</Document>
      <CreateDate>2016-11-15T23:18:55Z</CreateDate>
    </PolicyVersion>
  </GetPolicyVersionResult>
  <ResponseMetadata>
    <RequestId>9665acf0-c5c5-48e8-9915-4c1e228eb17c</RequestId>
  </ResponseMetadata>
</GetPolicyVersionResponse>

	`, url.PathEscape(`
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "iam:ChangePassword"
            ],
            "Resource": [
                "arn:aws:iam::*:user/${aws:username}"
            ]
        },
        {
            "Effect": "Allow",
            "Action": [
                "iam:GetAccountPasswordPolicy"
            ],
            "Resource": "*"
        }
    ]
}
	`))), nil
					},
				},
			},
		},
	}
}

func TestCreateEtlRoleFromAwsPolicy(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	for _, test := range []struct {
		Policy   awsIamPolicy
		Document awsIamPolicyDocument
		Role     types.EtlRole
	}{
		{
			Policy: awsIamPolicy{
				PolicyName: "NAME",
			},
			Document: awsIamPolicyDocument{
				Statement: []awsIamPolicyStatement{
					awsIamPolicyStatement{
						Effect:   "Allow",
						Action:   "test2",
						Resource: "test1",
					},
				},
			},
			Role: types.EtlRole{
				Name: "NAME",
				Permissions: map[string][]string{
					"test1": []string{"test2"},
				},
			},
		},
		{
			Policy: awsIamPolicy{
				PolicyName: "Another Name",
			},
			Document: awsIamPolicyDocument{
				Statement: []awsIamPolicyStatement{
					awsIamPolicyStatement{
						Effect:   "Allow",
						Action:   []interface{}{"a1", "a2"},
						Resource: []interface{}{"r1", "r2"},
					},
				},
			},
			Role: types.EtlRole{
				Name: "Another Name",
				Permissions: map[string][]string{
					"r1": []string{"a1", "a2"},
					"r2": []string{"a1", "a2"},
				},
			},
		},
		{
			Policy: awsIamPolicy{
				PolicyName: "Another Name",
			},
			Document: awsIamPolicyDocument{
				Statement: []awsIamPolicyStatement{
					awsIamPolicyStatement{
						Effect:   "Allow",
						Action:   []interface{}{"a1", "a2"},
						Resource: []interface{}{"r1", "r2"},
					},
					awsIamPolicyStatement{
						Effect:   "Allow",
						Action:   []interface{}{"a3"},
						Resource: []interface{}{"r1"},
					},
				},
			},
			Role: types.EtlRole{
				Name: "Another Name",
				Permissions: map[string][]string{
					"r1": []string{"a1", "a2", "a3"},
					"r2": []string{"a1", "a2"},
				},
			},
		},
	} {
		cmp := createEtlRoleFromAwsPolicy(&test.Policy, &test.Document)
		g.Expect(*cmp).To(gomega.Equal(test.Role))
	}
}

func TestAwsIamUserToEtlUser(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	for _, test := range []struct {
		AwsUser awsIamUser
	}{
		{
			AwsUser: awsIamUser{
				UserName:   "mikebao",
				CreateDate: time.Date(2020, 10, 12, 12, 35, 11, 10, time.UTC),
			},
		},
	} {
		cmp := test.AwsUser.toEtlUser()
		g.Expect(cmp.Username).To(gomega.Equal(test.AwsUser.UserName))
		g.Expect(*cmp.CreatedTime).To(gomega.BeTemporally("==", test.AwsUser.CreateDate))
		g.Expect(len(cmp.Roles)).To(gomega.Equal(0))
	}
}

func TestGetInlineUserPolicies(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	conn, err := CreateAWSConnector(&EtlAWSOptions{
		Client: createStandardClient(),
	})
	g.Expect(err).To(gomega.BeNil())

	itf, err := conn.GetUserInterface()
	g.Expect(err).To(gomega.BeNil())

	for _, test := range []struct {
		User     string
		Policies []*awsIamPolicy
	}{
		{
			User: "mikebao-iam",
			Policies: []*awsIamPolicy{
				&awsIamPolicy{
					PolicyName: "TestInlinePolicy",
					Type:       "User",
					Inline:     true,
					InlineId:   "mikebao-iam",
				},
			},
		},
		{
			User:     "grchive-api",
			Policies: []*awsIamPolicy{},
		},
	} {
		policies, source, err := itf.(*EtlAWSConnectorUser).getInlineUserPolicies(test.User)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(source).NotTo(gomega.BeNil())
		g.Expect(len(source.Commands)).To(gomega.Equal(1))
		g.Expect(policies).To(gomega.Equal(test.Policies))
	}
}

func TestGetAttachedUserPolicies(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	conn, err := CreateAWSConnector(&EtlAWSOptions{
		Client: createStandardClient(),
	})
	g.Expect(err).To(gomega.BeNil())

	itf, err := conn.GetUserInterface()
	g.Expect(err).To(gomega.BeNil())

	for _, test := range []struct {
		User     string
		Policies []*awsIamPolicy
	}{
		{
			User: "mikebao-iam",
			Policies: []*awsIamPolicy{
				&awsIamPolicy{
					PolicyName: "IAMUserChangePassword",
					PolicyArn:  "arn:aws:iam::aws:policy/IAMUserChangePassword",
					Type:       "User",
					Inline:     false,
				},
			},
		},
		{
			User: "grchive-api",
			Policies: []*awsIamPolicy{
				&awsIamPolicy{
					PolicyName: "GRCHiveAPI",
					PolicyArn:  "arn:aws:iam::455499087251:policy/GRCHiveAPI",
					Type:       "User",
					Inline:     false,
				},
			},
		},
	} {
		policies, source, err := itf.(*EtlAWSConnectorUser).getAttachedUserPolicies(test.User)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(source).NotTo(gomega.BeNil())
		g.Expect(len(source.Commands)).To(gomega.Equal(1))
		g.Expect(policies).To(gomega.Equal(test.Policies))
	}
}

func TestGetInlineGroupPolicies(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	conn, err := CreateAWSConnector(&EtlAWSOptions{
		Client: createStandardClient(),
	})
	g.Expect(err).To(gomega.BeNil())

	itf, err := conn.GetUserInterface()
	g.Expect(err).To(gomega.BeNil())

	for _, test := range []struct {
		Group    string
		Policies []*awsIamPolicy
	}{
		{
			Group: "GRCHive",
			Policies: []*awsIamPolicy{
				&awsIamPolicy{
					PolicyName: "policygen-GRCHive-202008201657",
					Type:       "Group",
					Inline:     true,
					InlineId:   "GRCHive",
				},
			},
		},
	} {
		policies, source, err := itf.(*EtlAWSConnectorUser).getInlineGroupPolicies(test.Group)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(source).NotTo(gomega.BeNil())
		g.Expect(len(source.Commands)).To(gomega.Equal(1))
		g.Expect(policies).To(gomega.Equal(test.Policies))
	}
}

func TestGetAttachedGroupPolicies(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	conn, err := CreateAWSConnector(&EtlAWSOptions{
		Client: createStandardClient(),
	})
	g.Expect(err).To(gomega.BeNil())

	itf, err := conn.GetUserInterface()
	g.Expect(err).To(gomega.BeNil())

	for _, test := range []struct {
		Group    string
		Policies []*awsIamPolicy
	}{
		{
			Group: "GRCHive",
			Policies: []*awsIamPolicy{
				&awsIamPolicy{
					PolicyName: "GRCHiveAPI",
					PolicyArn:  "arn:aws:iam::455499087251:policy/GRCHiveAPI",
					Type:       "Group",
					Inline:     false,
				},
			},
		},
	} {
		policies, source, err := itf.(*EtlAWSConnectorUser).getAttachedGroupPolicies(test.Group)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(source).NotTo(gomega.BeNil())
		g.Expect(len(source.Commands)).To(gomega.Equal(1))
		g.Expect(policies).To(gomega.Equal(test.Policies))
	}
}

func TestGetUserGroups(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	conn, err := CreateAWSConnector(&EtlAWSOptions{
		Client: createStandardClient(),
	})
	g.Expect(err).To(gomega.BeNil())

	itf, err := conn.GetUserInterface()
	g.Expect(err).To(gomega.BeNil())

	for _, test := range []struct {
		User   string
		Groups []*awsIamGroup
	}{
		{
			User:   "mikebao-iam",
			Groups: []*awsIamGroup{},
		},
		{
			User: "grchive-api",
			Groups: []*awsIamGroup{
				&awsIamGroup{
					Path:      "/",
					GroupName: "GRCHive",
					GroupId:   "AGPAWUDO4PWJ3VBSHRMBX",
					Arn:       "arn:aws:iam::455499087251:group/GRCHive",
				},
			},
		},
	} {
		groups, source, err := itf.(*EtlAWSConnectorUser).getUserGroups(test.User)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(source).NotTo(gomega.BeNil())
		g.Expect(len(source.Commands)).To(gomega.Equal(1))
		g.Expect(groups).To(gomega.Equal(test.Groups))
	}
}

func TestGetGroupPolicies(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	conn, err := CreateAWSConnector(&EtlAWSOptions{
		Client: createStandardClient(),
	})
	g.Expect(err).To(gomega.BeNil())

	itf, err := conn.GetUserInterface()
	g.Expect(err).To(gomega.BeNil())

	for _, test := range []struct {
		Group    string
		Policies []*awsIamPolicy
	}{
		{
			Group: "GRCHive",
			Policies: []*awsIamPolicy{
				&awsIamPolicy{
					PolicyName: "policygen-GRCHive-202008201657",
					Type:       "Group",
					Inline:     true,
					InlineId:   "GRCHive",
				},
				&awsIamPolicy{
					PolicyName: "GRCHiveAPI",
					PolicyArn:  "arn:aws:iam::455499087251:policy/GRCHiveAPI",
					Type:       "Group",
					Inline:     false,
				},
			},
		},
	} {
		policies, source, err := itf.(*EtlAWSConnectorUser).getGroupPolicies(test.Group)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(source).NotTo(gomega.BeNil())
		g.Expect(len(source.Commands)).To(gomega.Equal(2))
		g.Expect(policies).To(gomega.Equal(test.Policies))
	}
}

func TestGetUserPolicies(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	conn, err := CreateAWSConnector(&EtlAWSOptions{
		Client: createStandardClient(),
	})
	g.Expect(err).To(gomega.BeNil())

	itf, err := conn.GetUserInterface()
	g.Expect(err).To(gomega.BeNil())

	for _, test := range []struct {
		User        string
		Policies    []*awsIamPolicy
		NumCommands int
	}{
		{
			User: "mikebao-iam",
			Policies: []*awsIamPolicy{
				&awsIamPolicy{
					PolicyName: "TestInlinePolicy",
					Type:       "User",
					Inline:     true,
					InlineId:   "mikebao-iam",
				},
				&awsIamPolicy{
					PolicyName: "IAMUserChangePassword",
					PolicyArn:  "arn:aws:iam::aws:policy/IAMUserChangePassword",
					Type:       "User",
					Inline:     false,
				},
			},
			NumCommands: 3,
		},
		{
			User: "grchive-api",
			Policies: []*awsIamPolicy{
				&awsIamPolicy{
					PolicyName: "GRCHiveAPI",
					PolicyArn:  "arn:aws:iam::455499087251:policy/GRCHiveAPI",
					Type:       "User",
					Inline:     false,
				},
				&awsIamPolicy{
					PolicyName: "policygen-GRCHive-202008201657",
					Type:       "Group",
					Inline:     true,
					InlineId:   "GRCHive",
				},
				&awsIamPolicy{
					PolicyName: "GRCHiveAPI",
					PolicyArn:  "arn:aws:iam::455499087251:policy/GRCHiveAPI",
					Type:       "Group",
					Inline:     false,
				},
			},
			NumCommands: 5,
		},
	} {
		policies, source, err := itf.(*EtlAWSConnectorUser).getUserPolicies(test.User)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(source).NotTo(gomega.BeNil())
		g.Expect(len(source.Commands)).To(gomega.Equal(test.NumCommands))
		g.Expect(policies).To(gomega.Equal(test.Policies))
	}
}

func TestGetInlineUserPolicyDocument(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	conn, err := CreateAWSConnector(&EtlAWSOptions{
		Client: createStandardClient(),
	})
	g.Expect(err).To(gomega.BeNil())

	itf, err := conn.GetUserInterface()
	g.Expect(err).To(gomega.BeNil())

	for _, test := range []struct {
		Policy   awsIamPolicy
		Document awsIamPolicyDocument
	}{
		{
			Policy: awsIamPolicy{
				PolicyName: "TestInlinePolicy",
				InlineId:   "mikebao-iam",
			},
			Document: awsIamPolicyDocument{
				Version: "2012-10-17",
				Statement: []awsIamPolicyStatement{
					awsIamPolicyStatement{
						Effect:   "Allow",
						Action:   "ec2:CreateImage",
						Resource: "*",
					},
				},
			},
		},
	} {
		doc, source, err := itf.(*EtlAWSConnectorUser).getAwsInlineUserPolicyDocument(&test.Policy)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(source).NotTo(gomega.BeNil())
		g.Expect(len(source.Commands)).To(gomega.Equal(1))
		g.Expect(*doc).To(gomega.Equal(test.Document))
	}
}

func TestGetInlineGroupPolicyDocument(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	conn, err := CreateAWSConnector(&EtlAWSOptions{
		Client: createStandardClient(),
	})
	g.Expect(err).To(gomega.BeNil())

	itf, err := conn.GetUserInterface()
	g.Expect(err).To(gomega.BeNil())

	for _, test := range []struct {
		Policy   awsIamPolicy
		Document awsIamPolicyDocument
	}{
		{
			Policy: awsIamPolicy{
				PolicyName: "policygen-GRCHive-202008201657",
				InlineId:   "GRCHive",
			},
			Document: awsIamPolicyDocument{
				Version: "2012-10-17",
				Statement: []awsIamPolicyStatement{
					awsIamPolicyStatement{
						Effect:   "Allow",
						Action:   []interface{}{"waf:GetRule"},
						Resource: []interface{}{"*"},
					},
				},
			},
		},
	} {
		doc, source, err := itf.(*EtlAWSConnectorUser).getAwsInlineGroupPolicyDocument(&test.Policy)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(source).NotTo(gomega.BeNil())
		g.Expect(len(source.Commands)).To(gomega.Equal(1))
		g.Expect(*doc).To(gomega.Equal(test.Document))
	}
}

func TestGetAttachedPolicyDocument(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	conn, err := CreateAWSConnector(&EtlAWSOptions{
		Client: createStandardClient(),
	})
	g.Expect(err).To(gomega.BeNil())

	itf, err := conn.GetUserInterface()
	g.Expect(err).To(gomega.BeNil())

	for _, test := range []struct {
		Policy   awsIamPolicy
		Document awsIamPolicyDocument
	}{
		{
			Policy: awsIamPolicy{
				PolicyArn: "arn:aws:iam::455499087251:policy/GRCHiveAPI",
			},
			Document: awsIamPolicyDocument{
				Version: "2012-10-17",
				Statement: []awsIamPolicyStatement{
					awsIamPolicyStatement{
						Effect: "Allow",
						Action: []interface{}{
							"iam:GetPolicyVersion",
							"iam:GetPolicy",
							"iam:GetUserPolicy",
							"iam:ListGroupsForUser",
							"iam:ListGroupPolicies",
							"iam:ListAttachedUserPolicies",
							"iam:ListUsers",
							"iam:ListAttachedGroupPolicies",
							"iam:ListUserPolicies",
							"iam:GetGroupPolicy",
						},
						Resource: "*",
					},
				},
			},
		},
		{
			Policy: awsIamPolicy{
				PolicyArn: "arn:aws:iam::aws:policy/IAMUserChangePassword",
			},
			Document: awsIamPolicyDocument{
				Version: "2012-10-17",
				Statement: []awsIamPolicyStatement{
					awsIamPolicyStatement{
						Effect: "Allow",
						Action: []interface{}{
							"iam:ChangePassword",
						},
						Resource: []interface{}{
							"arn:aws:iam::*:user/${aws:username}",
						},
					},
					awsIamPolicyStatement{
						Effect: "Allow",
						Action: []interface{}{
							"iam:GetAccountPasswordPolicy",
						},
						Resource: "*",
					},
				},
			},
		},
	} {
		doc, source, err := itf.(*EtlAWSConnectorUser).getAwsAttachedPolicyDocument(&test.Policy)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(source).NotTo(gomega.BeNil())
		g.Expect(len(source.Commands)).To(gomega.Equal(2))
		g.Expect(*doc).To(gomega.Equal(test.Document))
	}
}

func TestGetAwsPolicyDocument(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	conn, err := CreateAWSConnector(&EtlAWSOptions{
		Client: createStandardClient(),
	})
	g.Expect(err).To(gomega.BeNil())

	itf, err := conn.GetUserInterface()
	g.Expect(err).To(gomega.BeNil())

	for _, test := range []struct {
		Policy   awsIamPolicy
		Document awsIamPolicyDocument
	}{
		{
			Policy: awsIamPolicy{
				PolicyArn: "arn:aws:iam::455499087251:policy/GRCHiveAPI",
				Type:      "Group",
			},
			Document: awsIamPolicyDocument{
				Version: "2012-10-17",
				Statement: []awsIamPolicyStatement{
					awsIamPolicyStatement{
						Effect: "Allow",
						Action: []interface{}{
							"iam:GetPolicyVersion",
							"iam:GetPolicy",
							"iam:GetUserPolicy",
							"iam:ListGroupsForUser",
							"iam:ListGroupPolicies",
							"iam:ListAttachedUserPolicies",
							"iam:ListUsers",
							"iam:ListAttachedGroupPolicies",
							"iam:ListUserPolicies",
							"iam:GetGroupPolicy",
						},
						Resource: "*",
					},
				},
			},
		},
		{
			Policy: awsIamPolicy{
				PolicyArn: "arn:aws:iam::455499087251:policy/GRCHiveAPI",
				Type:      "User",
			},
			Document: awsIamPolicyDocument{
				Version: "2012-10-17",
				Statement: []awsIamPolicyStatement{
					awsIamPolicyStatement{
						Effect: "Allow",
						Action: []interface{}{
							"iam:GetPolicyVersion",
							"iam:GetPolicy",
							"iam:GetUserPolicy",
							"iam:ListGroupsForUser",
							"iam:ListGroupPolicies",
							"iam:ListAttachedUserPolicies",
							"iam:ListUsers",
							"iam:ListAttachedGroupPolicies",
							"iam:ListUserPolicies",
							"iam:GetGroupPolicy",
						},
						Resource: "*",
					},
				},
			},
		},
		{
			Policy: awsIamPolicy{
				PolicyName: "policygen-GRCHive-202008201657",
				Type:       "Group",
				Inline:     true,
				InlineId:   "GRCHive",
			},
			Document: awsIamPolicyDocument{
				Version: "2012-10-17",
				Statement: []awsIamPolicyStatement{
					awsIamPolicyStatement{
						Effect:   "Allow",
						Action:   []interface{}{"waf:GetRule"},
						Resource: []interface{}{"*"},
					},
				},
			},
		},
		{
			Policy: awsIamPolicy{
				PolicyName: "TestInlinePolicy",
				Type:       "User",
				Inline:     true,
				InlineId:   "mikebao-iam",
			},
			Document: awsIamPolicyDocument{
				Version: "2012-10-17",
				Statement: []awsIamPolicyStatement{
					awsIamPolicyStatement{
						Effect:   "Allow",
						Action:   "ec2:CreateImage",
						Resource: "*",
					},
				},
			},
		},
	} {
		doc, source, err := itf.(*EtlAWSConnectorUser).getAwsPolicyDocument(&test.Policy)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(source).NotTo(gomega.BeNil())
		if test.Policy.Inline {
			g.Expect(len(source.Commands)).To(gomega.Equal(1))
		} else {
			g.Expect(len(source.Commands)).To(gomega.Equal(2))
		}
		g.Expect(*doc).To(gomega.Equal(test.Document))
	}
}

func TestGetUserListing(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	conn, err := CreateAWSConnector(&EtlAWSOptions{
		Client: createStandardClient(),
	})
	g.Expect(err).To(gomega.BeNil())

	itf, err := conn.GetUserInterface()
	g.Expect(err).To(gomega.BeNil())

	users, source, err := itf.GetUserListing()
	g.Expect(err).To(gomega.BeNil())

	g.Expect(source).NotTo(gomega.BeNil())
	// 1. ListUsers (+1)
	// 2. For Each User, ListUserPolicies (+2)
	// 3. For Each User, ListAttachedUserPolicies (+2)
	// 4. For Each User, ListGroupsForUser (+2)
	// 4. For Each Group, ListGroupPolicies (+1)
	// 4. For Each Group, ListAttachedGroupPolicies (+1)
	// 5. For Each Inline Policy, Get*Policy (+2)
	// 5. For Each Attached Policy, GetPolicy (+2)
	// 5. For Each Attached Policy, GetPolicyVersion (+2)
	g.Expect(len(source.Commands)).To(gomega.Equal(15))

	u1Time := time.Date(2020, 8, 19, 22, 6, 55, 0, time.UTC)
	u2Time := time.Date(2020, 8, 19, 22, 7, 51, 0, time.UTC)

	refUsers := map[string]*types.EtlUser{
		"grchive-api": &types.EtlUser{
			Username:    "grchive-api",
			CreatedTime: &u1Time,
			Roles: map[string]*types.EtlRole{
				"GRCHiveAPI": &types.EtlRole{
					Name: "GRCHiveAPI",
					Permissions: map[string][]string{
						"*": []string{
							"iam:GetPolicyVersion",
							"iam:GetPolicy",
							"iam:GetUserPolicy",
							"iam:ListGroupsForUser",
							"iam:ListGroupPolicies",
							"iam:ListAttachedUserPolicies",
							"iam:ListUsers",
							"iam:ListAttachedGroupPolicies",
							"iam:ListUserPolicies",
							"iam:GetGroupPolicy",
						},
					},
				},
				"policygen-GRCHive-202008201657": &types.EtlRole{
					Name: "policygen-GRCHive-202008201657",
					Permissions: map[string][]string{
						"*": []string{"waf:GetRule"},
					},
				},
			},
		},
		"mikebao-iam": &types.EtlUser{
			Username:    "mikebao-iam",
			CreatedTime: &u2Time,
			Roles: map[string]*types.EtlRole{
				"TestInlinePolicy": &types.EtlRole{
					Name: "TestInlinePolicy",
					Permissions: map[string][]string{
						"*": []string{"ec2:CreateImage"},
					},
				},
				"IAMUserChangePassword": &types.EtlRole{
					Name: "IAMUserChangePassword",
					Permissions: map[string][]string{
						"arn:aws:iam::*:user/${aws:username}": []string{"iam:ChangePassword"},
						"*":                                   []string{"iam:GetAccountPasswordPolicy"},
					},
				},
			},
		},
	}
	test_utility.CompareUserListing(g, users, refUsers, test_utility.CompareUserListingOptions{})
}
