package main

type AuthzUser authzUser

var NewAuthzService = newAuthzService

func AuthzServiceLookup(x *authzService, userID string) *AuthzUser {
	return (*AuthzUser)(x.lookup(userID))
}
func AuthzUserAllowed(x *AuthzUser) []string {
	authz := (*authzUser)(x)
	return authz.rolePtr.AllowedTags
}
