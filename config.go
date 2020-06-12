package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
)

const (
	ConfigFileName       = "cbot.json"
	ConfigDirNameWindows = "cbot"
	ConfigDirNamePosix   = ".cbot"
)

var (
	ConfigFileNotFoundError = errors.New("config file not found")
)

type Config struct {
	AccessToken     string `json:"AccessToken"`
	SecretKey       string `json:"SecretKey"`
	ApiPath         string `json:"ApiPath"`
	ContentLanguage string `json:"ContentLanguage"`
}

func getConfigDir() string {
	if runtime.GOOS == "windows" {
		dir := os.Getenv("APPDATA")
		return filepath.Join(dir, ConfigDirNameWindows)
	} else {
		dir := os.Getenv("HOME")
		return filepath.Join(dir, ConfigDirNamePosix)
	}
}

func getConfigPath() string {
	return filepath.Join(getConfigDir(), ConfigFileName)
}

func isExist(path string) (bool, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func getConfig() (*Config, error) {
	path := getConfigPath()

	if ok, err := isExist(path); err != nil {
		return nil, fmt.Errorf("config file load failed. %v: %v", path, err)
	} else if !ok {
		return nil, ConfigFileNotFoundError
	}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config file load failed. %v: %v", path, err)
	}

	var config Config
	err = json.Unmarshal(b, &config)
	if err != nil {
		return nil, fmt.Errorf("config file load failed. %v: %v", path, err)
	}
	// TODO
	config.ContentLanguage = "ja"
	return &config, nil
}

func createConfigFile() (*Config, error) {
	path := getConfigPath()

	if err := os.MkdirAll(getConfigDir(), 0700); err != nil {
		return nil, err
	}

	config, err := showConfigSetupPrompt()
	if err != nil {
		return nil, err
	}

	b, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(path, b, 0700)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func showConfigSetupPrompt() (*Config, error) {
	var config Config
	stdin := bufio.NewScanner(os.Stdin)

	fmt.Printf("Input your Access Token:")
	if !stdin.Scan() {
		return nil, fmt.Errorf("setup canceled")
	}
	config.AccessToken = stdin.Text()

	fmt.Printf("Input your Secret Key:")
	if !stdin.Scan() {
		return nil, fmt.Errorf("setup canceled")
	}
	config.SecretKey = stdin.Text()

	fmt.Printf("Input your API public path:")
	if !stdin.Scan() {
		return nil, fmt.Errorf("setup canceled")
	}
	config.ApiPath = stdin.Text()

	return &config, nil
}

func updateConfigFile() (*Config, error) {
	return createConfigFile()
}

func displayCurrentConfig() {
	fmt.Printf("Access Token : %s\n", UserConfig.AccessToken)
	fmt.Printf("Secret Key   : %s\n", UserConfig.SecretKey)
	fmt.Printf("API Path     : %s\n", UserConfig.ApiPath)
}
