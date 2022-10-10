package actions

import (
	forestvpn_api "github.com/forestvpn/api-client-go"
)

type LocationWrapper struct {
	Location forestvpn_api.Location
	Premium  bool
}

func IsPremiumLocation(billingFeature forestvpn_api.BillingFeature, location forestvpn_api.Location) bool {
	constraint := billingFeature.GetConstraints()[0]
	subject := constraint.GetSubject()
	bid := billingFeature.GetBundleId()
	premium := false

	for _, s := range subject {
		if s == location.GetId() {
			if location.GetId() == "7fc5b17c-eddf-413f-8b37-9d36eb5e33ec" || location.GetId() == "b134d679-8697-4dc6-b629-c4c189392fca" {
				premium = false
			} else if bid == "com.forestvpn.premium" {
				premium = true
			}
			break
		}
	}

	return premium
}

func GetLocationWrappers(billingFeature forestvpn_api.BillingFeature, locations []forestvpn_api.Location) []LocationWrapper {
	var wrappers []LocationWrapper

	for _, l := range locations {
		wrapper := LocationWrapper{Location: l, Premium: IsPremiumLocation(billingFeature, l)}
		wrappers = append(wrappers, wrapper)
	}

	return wrappers
}
