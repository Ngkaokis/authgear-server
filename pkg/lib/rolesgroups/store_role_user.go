package rolesgroups

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/authgear/authgear-server/pkg/util/slice"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

func (s *Store) ListRolesByUserID(userID string) ([]*Role, error) {
	q := s.SQLBuilder.Select(
		"r.id",
		"r.created_at",
		"r.updated_at",
		"r.key",
		"r.name",
		"r.description",
	).
		From(s.SQLBuilder.TableName("_auth_user_role"), "ur").
		Join(s.SQLBuilder.TableName("_auth_role"), "r", "ur.role_id = r.id").
		Where("ur.user_id = ?", userID)

	return s.queryRoles(q)
}

func (s *Store) ListUserIDsByRoleID(roleID string, pageArgs graphqlutil.PageArgs) ([]string, uint64, error) {
	q := s.SQLBuilder.Select(
		"u.id",
	).
		From(s.SQLBuilder.TableName("_auth_user_role"), "ur").
		Join(s.SQLBuilder.TableName("_auth_user"), "u", "ur.user_id = u.id").
		Where("ur.role_id = ?", roleID)

	q, offset, err := db.ApplyPageArgs(q, pageArgs)
	if err != nil {
		return nil, 0, err
	}

	userIDs, err := s.queryUsers(q)
	if err != nil {
		return nil, 0, err
	}

	return userIDs, offset, nil
}

type AddRoleToUsersOptions struct {
	RoleKey string
	UserIDs []string
}

func (s *Store) AddRoleToUsers(options *AddRoleToUsersOptions) (*Role, error) {
	r, err := s.GetRoleByKey(options.RoleKey)
	if err != nil {
		return nil, err
	}

	userIds, err := s.GetManyUsersByIds(options.UserIDs)
	if err != nil {
		return nil, err
	}

	var seenKeys []string
	now := s.Clock.NowUTC()
	for _, u := range userIds {
		id := uuid.New()
		q := s.SQLBuilder.
			Insert(s.SQLBuilder.TableName("_auth_user_role")).
			Columns(
				"id",
				"created_at",
				"updated_at",
				"user_id",
				"role_id",
			).
			Values(
				id,
				now,
				now,
				u,
				r.ID,
			).Suffix("ON CONFLICT DO NOTHING")

		_, err := s.SQLExecutor.ExecWith(q)
		if err != nil {
			return nil, err
		}

		seenKeys = append(seenKeys, u)
	}

	missingKeys := slice.ExceptStrings(options.UserIDs, seenKeys)
	if len(missingKeys) > 0 {
		err := UserUnknownKeys.NewWithInfo("unknown user ids", apierrors.Details{"keys": missingKeys})
		return nil, err
	}

	return r, nil
}

type RemoveRoleFromUsersOptions struct {
	RoleKey string
	UserIDs []string
}

func (s *Store) RemoveRoleFromUsers(options *RemoveRoleFromUsersOptions) (*Role, error) {
	r, err := s.GetRoleByKey(options.RoleKey)
	if err != nil {
		return nil, err
	}

	userIds, err := s.GetManyUsersByIds(options.UserIDs)
	if err != nil {
		return nil, err
	}

	var seenKeys []string
	for _, u := range userIds {
		q := s.SQLBuilder.
			Delete(s.SQLBuilder.TableName("_auth_user_role")).
			Where("role_id = ? AND user_id = ?", r.ID, u)

		_, err := s.SQLExecutor.ExecWith(q)
		if err != nil {
			return nil, err
		}

		seenKeys = append(seenKeys, u)
	}

	missingKeys := slice.ExceptStrings(options.UserIDs, seenKeys)
	if len(missingKeys) > 0 {
		err := UserUnknownKeys.NewWithInfo("unknown user ids", apierrors.Details{"keys": missingKeys})
		return nil, err
	}

	return r, nil
}

type AddUserToRolesOptions struct {
	UserID   string
	RoleKeys []string
}

func (s *Store) AddUserToRoles(options *AddUserToRolesOptions) error {
	u, err := s.GetUserByID(options.UserID)
	if err != nil {
		return err
	}

	rs, err := s.GetManyRolesByKeys(options.RoleKeys)
	if err != nil {
		return err
	}

	var seenKeys []string
	now := s.Clock.NowUTC()
	for _, r := range rs {
		id := uuid.New()
		q := s.SQLBuilder.
			Insert(s.SQLBuilder.TableName("_auth_user_role")).
			Columns(
				"id",
				"created_at",
				"updated_at",
				"user_id",
				"role_id",
			).
			Values(
				id,
				now,
				now,
				u,
				r.ID,
			).Suffix("ON CONFLICT DO NOTHING")

		_, err := s.SQLExecutor.ExecWith(q)
		if err != nil {
			return err
		}

		seenKeys = append(seenKeys, r.Key)
	}

	missingKeys := slice.ExceptStrings(options.RoleKeys, seenKeys)
	if len(missingKeys) > 0 {
		err := RoleUnknownKeys.NewWithInfo("unknown role keys", apierrors.Details{"keys": missingKeys})
		return err
	}

	return nil
}

type RemoveUserFromRolesOptions struct {
	UserID   string
	RoleKeys []string
}

func (s *Store) RemoveUserFromRoles(options *RemoveUserFromRolesOptions) error {
	u, err := s.GetUserByID(options.UserID)
	if err != nil {
		return err
	}

	rs, err := s.GetManyRolesByKeys(options.RoleKeys)
	if err != nil {
		return err
	}

	var seenKeys []string
	for _, r := range rs {
		q := s.SQLBuilder.
			Delete(s.SQLBuilder.TableName("_auth_user_role")).
			Where("role_id = ? AND user_id = ?", r.ID, u)

		_, err := s.SQLExecutor.ExecWith(q)
		if err != nil {
			return err
		}

		seenKeys = append(seenKeys, r.Key)
	}

	missingKeys := slice.ExceptStrings(options.RoleKeys, seenKeys)
	if len(missingKeys) > 0 {
		err := RoleUnknownKeys.NewWithInfo("unknown role keys", apierrors.Details{"keys": missingKeys})
		return err
	}

	return nil
}
