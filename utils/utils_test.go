package utils_test

import (
	"sort"
	"strings"
	"testing"

	"github.com/forestvpn/cli/utils"
)

func TestExcludeDisallowedIps(t *testing.T) {
	allowed := []string{"10.0.0.0/8"}
	disallowed := []string{"10.0.1.0/24"}
	expectedAllowedIps := strings.Split("10.0.0.0/24,10.0.2.0/23,10.0.4.0/22,10.0.8.0/21,10.0.16.0/20,10.0.32.0/19,10.0.64.0/18,10.0.128.0/17,10.1.0.0/16,10.2.0.0/15,10.4.0.0/14,10.8.0.0/13,10.16.0.0/12,10.32.0.0/11,10.64.0.0/10,10.128.0.0/9", ",")
	allowedIps, err := utils.ExcludeDisallowedIpds(allowed, disallowed)

	if err != nil {
		t.Error(err)
	}

	sort.Strings(allowedIps)
	sort.Strings(expectedAllowedIps)

	result := strings.Join(allowedIps, ",")
	expectedResult := strings.Join(expectedAllowedIps, ",")

	if result != expectedResult {
		t.Errorf("%s != %s; want ==", result, expectedResult)
	}
}
