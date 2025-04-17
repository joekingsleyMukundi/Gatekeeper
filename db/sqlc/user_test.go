package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/joekingsleyMukundi/Gatekeeper/utils"
	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T) User {
	hashedPassword, err := utils.HashPassword(utils.RandomString(6))
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword)
	arg := CreateUserParams{
		Username:       utils.RandomUsername(),
		Email:          utils.RandomEmail(),
		HashedPassword: hashedPassword,
	}
	user, err := testQueries.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)
	require.Equal(t, user.Username, arg.Username)
	require.Equal(t, user.Email, arg.Email)
	require.Equal(t, user.HashedPassword, arg.HashedPassword)
	require.NotEmpty(t, user.CreatedAt)
	return user
}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUser(t *testing.T) {
	usercreated := createRandomUser(t)
	usergot, err := testQueries.GetUser(context.Background(), usercreated.Username)
	require.NoError(t, err)
	require.NotEmpty(t, usergot)
	require.Equal(t, usercreated.Email, usergot.Email)
	require.Equal(t, usercreated.Username, usergot.Username)
	require.Equal(t, usercreated.HashedPassword, usergot.HashedPassword)
	require.WithinDuration(t, usercreated.CreatedAt, usergot.CreatedAt, time.Second)
}

func TestUpdateUser(t *testing.T) {
	usercreated := createRandomUser(t)
	newPassword := utils.RandomString(6)
	hashedNewPass, err := utils.HashPassword(newPassword)
	require.NoError(t, err)
	require.NotEmpty(t, hashedNewPass)
	arg := UpdateUserParams{
		HashedPassword: sql.NullString{
			String: hashedNewPass,
			Valid:  true,
		},
		PasswordChangedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		Username: usercreated.Username,
	}
	updatedUser, err := testQueries.UpdateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, updatedUser)
	require.Equal(t, usercreated.Email, updatedUser.Email)
	require.Equal(t, usercreated.Username, updatedUser.Username)
	require.Equal(t, arg.HashedPassword.String, updatedUser.HashedPassword)
	require.WithinDuration(t, usercreated.CreatedAt, updatedUser.CreatedAt, time.Second)
	require.WithinDuration(t, arg.PasswordChangedAt.Time, updatedUser.PasswordChangedAt.UTC(), time.Second)
}

func TestListUsers(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomUser(t)
	}
	args := ListUsersParams{
		Limit:  5,
		Offset: 0,
	}
	users, err := testQueries.ListUsers(context.Background(), args)
	require.NoError(t, err)
	require.NotEmpty(t, users)
	for _, user := range users {
		require.NotEmpty(t, user)
	}
}
