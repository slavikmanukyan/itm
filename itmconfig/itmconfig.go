package itmconfig

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type ITMConfig struct {
	DESTINATION        string
	SOURCE             string
	USE_SSH            bool
	SSH_USER           string
	SSH_PASSWORD       string
	SSH_HOST           string
	SSH_PORT           int
	SSH_PRIVATE_KEY    string
	SSH_KEY_PASSPHRASE string
	SSH_AUTH_SOCK      bool
	IGNORE             map[string]bool
	IS_TERMINAL        bool
}

func Parse(dst string) ITMConfig {
	dst = filepath.Clean(dst)

	raw, err := ioutil.ReadFile(dst)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var c ITMConfig
	json.Unmarshal(raw, &c)
	return c
}
