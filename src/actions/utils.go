package actions

import forestvpn_api "github.com/forestvpn/api-client-go"

type LocationWrapper struct {
	Location forestvpn_api.Location
	Premium  bool
}

func GetWrappedLocations(subject []string, locations []forestvpn_api.Location) []LocationWrapper {
	var wrappedLocations []LocationWrapper

	for _, location := range locations {
		wrappedLocation := LocationWrapper{Location: location, Premium: true}

		for _, uuid := range subject {
			if uuid == location.GetId() {
				wrappedLocation.Premium = false
			}
		}
		wrappedLocations = append(wrappedLocations, wrappedLocation)
	}

	return wrappedLocations
}
