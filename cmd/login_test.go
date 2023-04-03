package cmd

import (
	"reflect"
	"testing"
)

type args struct {
	configFile  string
	noPrompt    bool
	envFile     string
	authContext string
}

var loginTests = []struct {
	name       string `json:"name,omitempty"`
	args       args   `json:"args"`
	shouldPass bool   `json:"shouldPass,omitempty"`
	expected   string `json:"expected,omitempty"`
}{
	{
		name: "Test from ENV no config file - PASS",
		args: args{
			envFile:  ".env_login",
			noPrompt: true,
		},
		shouldPass: true,
	},
	{
		name: "Test username w/ domain - PASS",
		args: args{
			envFile:  ".env_login",
			noPrompt: true,
		},
		shouldPass: true,
	},
	{
		name: "Test username w/o domain - PASS",
		args: args{
			envFile:  ".env_login_no_domain",
			noPrompt: true,
		},
		shouldPass: true,
	},
	{
		name: "Test username w/ domain and domain var set to same value - PASS",
		args: args{
			envFile:  ".env_login",
			noPrompt: true,
		},
		shouldPass: true,
	},
	{
		name: "Test username w/ domain and domain var set to different value - FAIL",
		args: args{
			envFile:  ".env_login_diff_domain",
			noPrompt: true,
		},
		shouldPass: false,
	},
	{
		name: "Test username w/o domain and empty domain - FAIL",
		args: args{
			envFile:  ".env_login_double_no_domain",
			noPrompt: true,
		},
		shouldPass: true,
	},
	{
		name: "Test from config file no ENV - PASS",
		args: args{
			envFile:    "",
			noPrompt:   true,
			configFile: "command_config.json",
		},
		shouldPass: true,
	},
}

func Test_authConfigFile(t *testing.T) {

	for _, tt := range loginTests {
		t.Run(tt.name, func(t *testing.T) {
			if got := authConfigFile(tt.args.configFile, tt.args.noPrompt); got != tt.shouldPass {
				t.Errorf("authConfigFile() = %v, shouldPass %v", got, tt.shouldPass)
			}
		})
	}
}

func Test_getPassword(t *testing.T) {
	for _, tt := range loginTests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getPassword(""); got != tt.expected {
				t.Errorf("getPassword() = %v, shouldPass %v", got, tt.shouldPass)
			}
		})
	}
}

func Test_loadConfigFile(t *testing.T) {

	for _, tt := range loginTests {
		t.Run(tt.name, func(t *testing.T) {
			if got := loadConfigFile(tt.args.configFile, nil); !reflect.DeepEqual(got, tt.shouldPass) {
				t.Errorf("loadConfigFile() = %v, shouldPass %v", got, tt.shouldPass)
			}
		})
	}
}
