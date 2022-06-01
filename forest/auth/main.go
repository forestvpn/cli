package auth

import "forest/utils"

func IsAuthenticated() (bool, string) {
	idToken := utils.LoadIDToken()
	return len(idToken) > 0, idToken
}
