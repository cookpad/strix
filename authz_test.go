package main_test

import (
	"testing"

	main "github.com/m-mizutani/strix"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthzService(t *testing.T) {
	raw := `{
		"users": [
			{"user_id": "alpha@example.com", "role":"blue"},
			{"user_id": "bravo@example.com", "role":"orange"},
			{"user_id": "charlie@example.com", "role":"orange"}
		],
		"roles": [
			{"name":"blue", "allowed_tags":[]},
			{"name":"orange", "allowed_tags":["spell.1"]}
		],
		"rules": [
			{"user_regex":"^delta@", "role":"blue"},
			{"user_regex":"@example.com$", "role":"orange"}
		]
	}`

	authz, err := main.NewAuthzService([]byte(raw))
	require.NoError(t, err)

	// Test default users and roles
	userA := main.AuthzServiceLookup(authz, "alpha@example.com")
	assert.NotNil(t, userA)
	assert.NotContains(t, main.AuthzUserAllowed(userA), "spell.1")

	userB := main.AuthzServiceLookup(authz, "bravo@example.com")
	assert.NotNil(t, userB)
	assert.Contains(t, main.AuthzUserAllowed(userB), "spell.1")

	userC := main.AuthzServiceLookup(authz, "charlie@example.com")
	assert.NotNil(t, userC)
	assert.Contains(t, main.AuthzUserAllowed(userC), "spell.1")

	// Test rules
	userD1 := main.AuthzServiceLookup(authz, "delta@example.com")
	assert.NotNil(t, userD1)
	assert.Equal(t, 0, len(main.AuthzUserAllowed(userD1)))

	userD2 := main.AuthzServiceLookup(authz, "delta@example.org")
	assert.NotNil(t, userD2)
	assert.Equal(t, 0, len(main.AuthzUserAllowed(userD2)))

	userD3 := main.AuthzServiceLookup(authz, "xxxxdelta@example.org")
	assert.Nil(t, userD3)

	userE1 := main.AuthzServiceLookup(authz, "echo@example.com")
	assert.NotNil(t, userE1)
	assert.Contains(t, main.AuthzUserAllowed(userE1), "spell.1")

	userF1 := main.AuthzServiceLookup(authz, "foxtrot@example.org")
	assert.Nil(t, userF1)
	userF2 := main.AuthzServiceLookup(authz, "foxtrot@example.comcom")
	assert.Nil(t, userF2)
}

func TestAuthzServiceInvalidJSON(t *testing.T) {
	raw := `{
		"users": [
			{"user_id": "alpha@example.com", "role":"blue"}
		],
		"roles": [
			{"name":"blue", "allowed_tags":[]}
		],
		"rules": [
			{"user_regex":"^delta@", "role":"blue"}
		],
	}` // ^^^ invalid comma

	_, err := main.NewAuthzService([]byte(raw))
	require.Error(t, err)
}

func TestAuthzServiceDuplicatedRole(t *testing.T) {
	raw := `{
		"users": [
			{"user_id": "alpha@example.com", "role":"blue"}
		],
		"roles": [
			{"name":"blue", "allowed_tags":[]},
			{"name":"blue", "allowed_tags":["spell.1"]}
		]
	}`

	_, err := main.NewAuthzService([]byte(raw))
	assert.EqualError(t, err, "Role 'blue' is duplicated")
}

func TestAuthzServiceDuplicatedUser(t *testing.T) {
	raw := `{
		"users": [
			{"user_id": "alpha@example.com", "role":"blue"},
			{"user_id": "alpha@example.com", "role":"orange"}
		],
		"roles": [
			{"name":"blue", "allowed_tags":[]},
			{"name":"orange", "allowed_tags":["spell.1"]}
		]
	}`

	_, err := main.NewAuthzService([]byte(raw))
	assert.EqualError(t, err, "User 'alpha@example.com' is duplicated")
}

func TestAuthzServiceUserRoleNotFound(t *testing.T) {
	raw := `{
		"users": [
			{"user_id": "alpha@example.com", "role":"blue"},
			{"user_id": "bravo@example.com", "role":"orange"}
		],
		"roles": [
			{"name":"blue", "allowed_tags":[]}
		]
	}`

	_, err := main.NewAuthzService([]byte(raw))
	assert.EqualError(t, err, "Role 'orange' of User 'bravo@example.com' is not found")
}

func TestAuthzServiceRuleRoleNotFound(t *testing.T) {
	raw := `{
		"roles": [
			{"name":"blue", "allowed_tags":[]}
		],
		"rules": [
			{"user_regex":"^delta@", "role":"orange"}
		]
	}`

	_, err := main.NewAuthzService([]byte(raw))
	assert.EqualError(t, err, "Role 'orange' of Rule '^delta@' is not found")
}

func TestAuthzServiceRuleInavlidRegex(t *testing.T) {
	raw := `{
		"roles": [
			{"name":"orange", "allowed_tags":[]}
		],
		"rules": [
			{"user_regex":"^[delta@", "role":"orange"}
		]
	}`

	_, err := main.NewAuthzService([]byte(raw))
	assert.EqualError(t, err, "Fail to compile regex of a rule: ^[delta@")
}
