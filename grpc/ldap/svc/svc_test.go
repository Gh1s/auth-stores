package svc

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"csb.nc/auth/stores/tools"
	"csb.nc/auth/stores/users"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	cfgName = "config"
	cfgType = "yaml"
	cfgPath = "../"

	viperKeyTestsUsersLdapAuthenticate = "tests.users.ldap.authenticate"
	viperKeyTestsUsersLdapFindClaims   = "tests.users.ldap.findClaims"
	viperKeyTestsUsersLdapSearchClaims = "tests.users.ldap.searchClaims"
)

func TestMain(m *testing.M) {
	os.Setenv(strings.ToUpper(tools.ViperKeyEnvironment), "test")
	tools.InitConfig(cfgName, cfgType, cfgPath)
	if viper.Get(viperKeyLdapUsername) == "" || viper.Get(viperKeyLdapPassword) == "" {
		zap.L().Fatal(missingLdapCredentialsError)
	} else {
		os.Exit(m.Run())
	}
}

type authenticateTestCases struct {
	Cases []authenticateTestCase `mapstructure:"cases"`
}

type authenticateTestCase struct {
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
	Subject   string `mapstructure:"subject"`
	Succeeded bool   `mapstructure:"succeeded"`
}

func TestAuthenticate(t *testing.T) {
	testCases := &authenticateTestCases{}
	viper.Sub(viperKeyTestsUsersLdapAuthenticate).Unmarshal(testCases)
	for i, tc := range testCases.Cases {
		t.Run(fmt.Sprintf("Case=%d;Principal=%s", i, tc.Username), func(t *testing.T) {
			req := &users.AuthRequest{
				Username: tc.Username,
				Password: tc.Password,
			}
			if resp := Authenticate(req); resp.Succeeded != tc.Succeeded {
				t.Errorf("Authentication failed with error %d.", resp.Error)
			} else if resp.Subject != tc.Subject {
				t.Errorf("The subject %s is different from %s.", resp.Subject, tc.Subject)
			}
		})
	}
}

type findClaimsTestCases struct {
	Cases []findClaimsTestCase `mapstructure:"cases"`
}

type findClaimsTestCase struct {
	Identifier     string               `mapstructure:"identifier"`
	IdentifierType users.IdentifierType `mapstructure:"identifierType"`
	Claims         []string             `mapstructure:"claims"`
	Values         map[string]string    `mapstructure:"values"`
	Succeeded      bool                 `mapstructure:"succeeded"`
}

func TestFindClaims(t *testing.T) {
	testCases := &findClaimsTestCases{}
	viper.Sub(viperKeyTestsUsersLdapFindClaims).Unmarshal(testCases)
	for i, tc := range testCases.Cases {
		t.Run(fmt.Sprintf("Case=%d;Identifier=%s;IdentifierType=%d", i, tc.Identifier, tc.IdentifierType), func(t *testing.T) {
			req := &users.ClaimsRequest{
				Identifier:     tc.Identifier,
				IdentifierType: tc.IdentifierType,
				Claims:         tc.Claims,
			}
			if resp := FindClaims(req); resp.Succeeded != tc.Succeeded {
				t.Errorf("FindClaims failed with error %d.", resp.Error)
			} else {
				for k, v := range tc.Values {
					if fc := resp.Claims[k]; fc != v {
						t.Errorf("Claim '%s' does not have the expected value of '%s', value: %s", k, v, fc)
					}
				}
			}
		})
	}
}

type searchClaimsTestCases struct {
	Cases []searchClaimsTestCase `mapstructure:"cases"`
}

type searchClaimsTestCase struct {
	Search    string              `mapstructure:"search"`
	Claims    []string            `mapstructure:"claims"`
	Values    []map[string]string `mapstructure:"values"`
	Succeeded bool                `mapstructure:"succeeded"`
}

func TestSearchClaims(t *testing.T) {
	testCases := &searchClaimsTestCases{}
	viper.Sub(viperKeyTestsUsersLdapSearchClaims).Unmarshal(testCases)
	for i, tc := range testCases.Cases {
		t.Run(fmt.Sprintf("Case=%d;Searcg=%s", i, tc.Search), func(t *testing.T) {
			req := &users.SearchRequest{
				Search: tc.Search,
				Claims: tc.Claims,
			}
			if resp := SearchClaims(req); resp.Succeeded != tc.Succeeded {
				t.Errorf("SearchClaims failed with error %d.", resp.Error)
			} else {
				for i, m := range tc.Values {
					for k, v := range m {
						if fc := resp.Results[i].Properties[k]; fc != v {
							t.Errorf("Claim '%s' does not have the expected value of '%s', value: %s", k, v, fc)
						}
					}
				}
			}
		})
	}
}
