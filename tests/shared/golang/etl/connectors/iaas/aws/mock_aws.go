package aws_utility

import (
	"net/http"
	"reflect"
)

type MockAWSFn func() (*http.Response, error)

type MockAWSIamClient struct {
	ListUsers                 MockAWSFn
	GetPolicy                 map[string]MockAWSFn
	GetPolicyVersion          map[string]map[string]MockAWSFn
	GetGroupPolicy            map[string]map[string]MockAWSFn
	GetUserPolicy             map[string]map[string]MockAWSFn
	ListGroupsForUser         map[string]MockAWSFn
	ListAttachedGroupPolicies map[string]MockAWSFn
	ListGroupPolicies         map[string]MockAWSFn
	ListAttachedUserPolicies  map[string]MockAWSFn
	ListUserPolicies          map[string]MockAWSFn
}

type MockAWSClient struct {
	Iam MockAWSIamClient
}

func (c *MockAWSClient) Do(req *http.Request) (*http.Response, error) {
	cVal := reflect.ValueOf(c).Elem()

	var ret []reflect.Value
	if req.URL.Host == "iam.amazonaws.com" {
		query := req.URL.Query()
		action := query["Action"][0]
		switch action {
		case "GetPolicy":
			policy := query["PolicyArn"][0]
			ret = cVal.
				FieldByName("Iam").
				FieldByName(action).
				MapIndex(reflect.ValueOf(policy)).
				Call([]reflect.Value{})
		case "GetPolicyVersion":
			policy := query["PolicyArn"][0]
			version := query["VersionId"][0]
			ret = cVal.
				FieldByName("Iam").
				FieldByName(action).
				MapIndex(reflect.ValueOf(policy)).
				MapIndex(reflect.ValueOf(version)).
				Call([]reflect.Value{})
		case "GetGroupPolicy":
			group := query["GroupName"][0]
			policy := query["PolicyName"][0]
			ret = cVal.
				FieldByName("Iam").
				FieldByName(action).
				MapIndex(reflect.ValueOf(group)).
				MapIndex(reflect.ValueOf(policy)).
				Call([]reflect.Value{})
		case "GetUserPolicy":
			user := query["UserName"][0]
			policy := query["PolicyName"][0]
			ret = cVal.
				FieldByName("Iam").
				FieldByName(action).
				MapIndex(reflect.ValueOf(user)).
				MapIndex(reflect.ValueOf(policy)).
				Call([]reflect.Value{})
		case "ListGroupPolicies":
			fallthrough
		case "ListAttachedGroupPolicies":
			group := query["GroupName"][0]
			ret = cVal.FieldByName("Iam").FieldByName(action).MapIndex(reflect.ValueOf(group)).Call([]reflect.Value{})
		case "ListGroupsForUser":
			fallthrough
		case "ListUserPolicies":
			fallthrough
		case "ListAttachedUserPolicies":
			user := query["UserName"][0]
			ret = cVal.FieldByName("Iam").FieldByName(action).MapIndex(reflect.ValueOf(user)).Call([]reflect.Value{})
		default:
			ret = cVal.FieldByName("Iam").FieldByName(action).Call([]reflect.Value{})
		}
	}

	var retResponse *http.Response
	var retError error

	if !ret[0].IsNil() {
		retResponse = ret[0].Interface().(*http.Response)
	}

	if !ret[1].IsNil() {
		retError = ret[1].Interface().(error)
	}

	return retResponse, retError
}
