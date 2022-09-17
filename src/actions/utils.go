package actions

import (
	forestvpn_api "github.com/forestvpn/api-client-go"
)

type LocationWrapper struct {
	Location forestvpn_api.Location
	Premium  bool
}

func IsPremiumUser(billingFeature forestvpn_api.BillingFeature) bool {
	return billingFeature.GetBundleId() == "com.forestvpn.premium"

}

func GetWrappedLocations(billingFeature forestvpn_api.BillingFeature, locations []forestvpn_api.Location) []LocationWrapper {
	var wrappedLocations []LocationWrapper
	constraint := billingFeature.GetConstraints()[0]
	subject := constraint.GetSubject()

	if IsPremiumUser(billingFeature) {
		for _, location := range locations {
			var premium bool

			if location.GetId() == "7fc5b17c-eddf-413f-8b37-9d36eb5e33ec" || location.GetId() == "b134d679-8697-4dc6-b629-c4c189392fca" {
				premium = false
			} else {
				premium = true
			}

			wrappedLocation := LocationWrapper{Location: location, Premium: premium}
			wrappedLocations = append(wrappedLocations, wrappedLocation)
		}
	} else {
		for _, location := range locations {
			wrappedLocation := LocationWrapper{Location: location, Premium: true}

			for _, uuid := range subject {
				if uuid == location.GetId() {
					wrappedLocation.Premium = false
				}
			}

			wrappedLocations = append(wrappedLocations, wrappedLocation)
		}
	}

	return wrappedLocations
}
