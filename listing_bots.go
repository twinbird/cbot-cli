package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
)

type listingBotsResponse struct {
	Code int                      `json:"code"`
	Bots []listingBotsResponseBot `json:"bots"`
}

type listingBotsResponseBot struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Created      string `json:"created"`
	LastModified string `json:"last_modified"`
	Creator      string `json:"creator"`
}

func listingBotsPortal(format string) {
	err := execListingBots(format)
	if err == UnauthorizedError {
		fmt.Fprintf(os.Stderr, "unauthorized error returned. Check your access token and key.")
		os.Exit(1)
	} else if err == ForbiddenError {
		fmt.Fprintf(os.Stderr, "forbidden error returned. Do you have a reference authorize?")
		os.Exit(1)
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}

func buildListingBotsURL() (string, error) {
	u, err := url.Parse(UserConfig.ApiPath)
	if err != nil {
		return "", err
	}

	u.Path = path.Join(u.Path, "bots")

	q := u.Query()
	q.Set("properties", "created,last_modified,creator")
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func buildListingBotsRequest() (*http.Request, error) {
	url, err := buildListingBotsURL()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("content-type", "application/json")
	req.Header.Add("content-language", UserConfig.ContentLanguage)
	req.Header.Add("access-token", UserConfig.AccessToken)
	req.Header.Add("secret-key", UserConfig.SecretKey)

	return req, nil
}

func processListingBotsResponse(resp *http.Response, format string) error {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var ret listingBotsResponse
	err = json.Unmarshal(body, &ret)
	if err != nil {
		return err
	}

	switch ret.Code {
	case 200:
		// nothing todo
	case 401:
		// unauthorized
		return UnauthorizedError
	case 403:
		// forbidden
		return ForbiddenError
	default:
		// error
		return fmt.Errorf("response code '%d' returned.", ret.Code)
	}

	if format != "text" {
		fmt.Println(string(body))
		return nil
	}

	fmt.Println("id\tname\tdescription\tcreated\tlast_modified\tcreator")
	for _, r := range ret.Bots {
		fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\n", r.Id, r.Name, r.Description, r.Created, r.LastModified, r.Creator)
	}

	return nil
}

func execListingBots(format string) error {
	client := http.DefaultClient

	req, err := buildListingBotsRequest()
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := processListingBotsResponse(resp, format); err != nil {
		return err
	}

	return nil
}
