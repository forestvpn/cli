package utils_test

import (
	"reflect"
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
		return
	}

	sort.Strings(actualResult)
	sort.Strings(expectedResult)

	if !reflect.DeepEqual(actualResult, expectedResult) {
		t.Errorf("Expected %s, got %s", expectedResult, actualResult)
	}
}

func TestExcludeDisallowedIpsExclude(t *testing.T) {
	allowed := []string{"192.168.0.0/16"}
	disallowed := "192.168.1.0/24"
	result, err := utils.ExcludeDisallowedIps(allowed, disallowed)
	if err != nil {
		t.Error(err)
	}

	disallowedNet := iplib.Net4FromStr(disallowed)
	for _, n := range result {
		resultingNet := iplib.Net4FromStr(n)
		if resultingNet.ContainsNet(disallowedNet) {
			t.Errorf("%s contains %s", resultingNet.String(), disallowedNet.String())
		}
	}
}

func TestHumanizeDuration(t *testing.T) {
	duration, err := time.ParseDuration("2h45m")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected := "2 hours 45 minutes 0 seconds"
	actual := utils.HumanizeDuration(duration)
	if expected != actual {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}
