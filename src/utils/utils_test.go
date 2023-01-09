package utils_test

import (
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/c-robinson/iplib"
	"github.com/forestvpn/cli/utils"
)

func TestExcludeDisallowedIpsResult(t *testing.T) {
	allowed := []string{"192.168.0.0/16"}
	disallowed := "192.168.1.0/24"
	expectedResult := strings.Split("192.168.0.0/24, 192.168.2.0/23, 192.168.4.0/22, 192.168.8.0/21, 192.168.16.0/20, 192.168.32.0/19, 192.168.64.0/18, 192.168.128.0/17", ", ")
	actualResult, err := utils.ExcludeDisallowedIps(allowed, disallowed)

	if err != nil {
		t.Error(err)
	}

	sort.Strings(actualResult)
	sort.Strings(expectedResult)

	actual := strings.Join(actualResult, ",")
	expected := strings.Join(expectedResult, ",")

	if actual != expected {
		t.Errorf("%s != %s; want ==", actual, expected)
	}
}

func TestExcludeDisallowedIpsExclude(t *testing.T) {
	allowed := []string{"192.168.0.0/16"}
	disallowed := "192.168.1.0/24"
	result, err := utils.ExcludeDisallowedIps(allowed, disallowed)

	if err != nil {
		t.Error(err)
	}
	for _, n := range result {
		resultingv4net := iplib.Net4FromStr(n)
		disallowedv4net := iplib.Net4FromStr(disallowed)

		if resultingv4net.ContainsNet(disallowedv4net) {
			t.Errorf("%s contains %s", resultingv4net.String(), disallowedv4net.String())
		}

	}
}

func TestHumanizeDuration(t *testing.T) {
	d, err := time.ParseDuration("2h45m")

	if err != nil {
		t.Error(err)
	}

	h := utils.HumanizeDuration(d)

	if h != "2 hours 45 minutes 0 seconds" {
		t.Error(h)
	}
}
