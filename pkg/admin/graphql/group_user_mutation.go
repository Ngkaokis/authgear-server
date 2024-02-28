package graphql

import (
	"github.com/authgear/authgear-server/pkg/lib/rolesgroups"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/graphql-go/graphql"
)

var addGroupToUsersInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "AddGroupToUsersInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"groupKey": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The key of the group.",
		},
		"userIDs": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewList(graphql.NewNonNull(graphql.ID)),
			Description: "The list of user ids.",
		},
	},
})

var addGroupToUsersPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "AddGroupToUsersPayload",
	Fields: graphql.Fields{
		"group": &graphql.Field{
			Type: graphql.NewNonNull(nodeGroup),
		},
	},
})

var _ = registerMutationField(
	"addGroupToUsers",
	&graphql.Field{
		Description: "Add the group to the users.",
		Type:        graphql.NewNonNull(addGroupToUsersPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(addGroupToUsersInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			groupKey := input["groupKey"].(string)
			userIDIfaces := input["userIDs"].([]interface{})
			userIDs := make([]string, len(userIDIfaces))
			for i, v := range userIDIfaces {
				userIDs[i] = v.(string)
			}
			gqlCtx := GQLContext(p.Context)

			options := &rolesgroups.AddGroupToUsersOptions{
				GroupKey: groupKey,
				UserIDs:  userIDs,
			}
			groupID, err := gqlCtx.RolesGroupsFacade.AddGroupToUsers(options)
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"group": gqlCtx.Groups.Load(groupID),
			}).Value, nil

		},
	},
)

var removeGroupFromUsersInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "RemoveGroupFromUsersInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"groupKey": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The key of the group.",
		},
		"userIDs": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewList(graphql.NewNonNull(graphql.ID)),
			Description: "The list of user ids.",
		},
	},
})

var removeGroupFromUsersPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "RemoveGroupToUsersPayload",
	Fields: graphql.Fields{
		"group": &graphql.Field{
			Type: graphql.NewNonNull(nodeGroup),
		},
	},
})

var _ = registerMutationField(
	"removeGroupFromUsers",
	&graphql.Field{
		Description: "Remove the group to the users.",
		Type:        graphql.NewNonNull(removeGroupFromUsersPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(removeGroupFromUsersInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			groupKey := input["groupKey"].(string)
			userIDIfaces := input["userIDs"].([]interface{})
			userIDs := make([]string, len(userIDIfaces))
			for i, v := range userIDIfaces {
				userIDs[i] = v.(string)
			}
			gqlCtx := GQLContext(p.Context)

			options := &rolesgroups.RemoveGroupFromUsersOptions{
				GroupKey: groupKey,
				UserIDs:  userIDs,
			}
			groupID, err := gqlCtx.RolesGroupsFacade.RemoveGroupFromUsers(options)
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"group": gqlCtx.Groups.Load(groupID),
			}).Value, nil

		},
	},
)

var addUserToGroupsInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "AddUserToGroupsInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"userID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "The ID of the user.",
		},
		"groupKeys": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewList(graphql.NewNonNull(graphql.String)),
			Description: "The list of group keys.",
		},
	},
})

var addUserToGroupsPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "AddUserToGroupsPayload",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type: graphql.NewNonNull(nodeUser),
		},
	},
})

var _ = registerMutationField(
	"addUserToGroups",
	&graphql.Field{
		Description: "Add the user to the groups.",
		Type:        graphql.NewNonNull(addUserToGroupsPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(addUserToGroupsInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			userID := input["userID"].(string)
			groupKeyIfaces := input["groupKeys"].([]interface{})
			groupKeys := make([]string, len(groupKeyIfaces))
			for i, v := range groupKeyIfaces {
				groupKeys[i] = v.(string)
			}
			gqlCtx := GQLContext(p.Context)

			options := &rolesgroups.AddUserToGroupsOptions{
				UserID:    userID,
				GroupKeys: groupKeys,
			}
			err := gqlCtx.RolesGroupsFacade.AddUserToGroups(options)
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"user": gqlCtx.Users.Load(userID),
			}).Value, nil

		},
	},
)

var removeUserFromGroupsInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "RemoveUserFromGroupsInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"userID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "The ID of the user.",
		},
		"groupKeys": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewList(graphql.NewNonNull(graphql.String)),
			Description: "The list of group keys.",
		},
	},
})

var removeUserFromGroupsPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "RemoveUserFromGroupsPayload",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type: graphql.NewNonNull(nodeUser),
		},
	},
})

var _ = registerMutationField(
	"removeUserFromGroups",
	&graphql.Field{
		Description: "Remove the user from the groups.",
		Type:        graphql.NewNonNull(removeUserFromGroupsPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(removeUserFromGroupsInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			userID := input["userID"].(string)
			groupKeyIfaces := input["groupKeys"].([]interface{})
			groupKeys := make([]string, len(groupKeyIfaces))
			for i, v := range groupKeyIfaces {
				groupKeys[i] = v.(string)
			}
			gqlCtx := GQLContext(p.Context)

			options := &rolesgroups.RemoveUserFromGroupsOptions{
				UserID:    userID,
				GroupKeys: groupKeys,
			}
			err := gqlCtx.RolesGroupsFacade.RemoveUserFromGroups(options)
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"user": gqlCtx.Users.Load(userID),
			}).Value, nil

		},
	},
)
