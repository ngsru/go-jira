package jira

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Error struct {
	StatusCode int
	Status     string
	Message    string
}

func (e Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Status, e.Message)
}

type Issue struct {
	Id      string
	Key     string
	Summary string
	Project string
	Data    map[string]interface{}
}

type Client struct {
	baseUrl *url.URL
	user    string
	pass    string
	res     *http.Client
}

func NewClient(jiraUrl string, user string, pass string, timeout time.Duration) (
	*Client, error) {
	baseUrl, err := url.Parse(jiraUrl)
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{Transport: &http.Transport{
		Dial: func(proto, addr string) (net.Conn, error) {
			return net.DialTimeout(proto, addr, timeout)
		},
	}}

	client := &Client{
		baseUrl: baseUrl,
		user:    user,
		pass:    pass,
		res:     httpClient,
	}

	return client, nil
}

func (client *Client) GetIssue(key string, fields []string) (
	issue *Issue, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	response, err := client.Request("GET",
		"issue/"+key+"/?fields="+strings.Join(fields, ","),
		[]byte{})
	if err != nil {
		return nil, err
	}

	rawData := map[string]interface{}{}

	if err := json.Unmarshal(response, &rawData); err != nil {
		return nil, err
	}

	issue = &Issue{
		Id:  rawData["id"].(string),
		Key: rawData["key"].(string),
	}
	tmp := strings.Split(issue.Key, "-")
	issue.Project = strings.ToLower(tmp[0])
	issue.Data = rawData["fields"].(map[string]interface{})

	if summary, ok := issue.Data["summary"].(string); ok {
		issue.Summary = summary
	}

	return issue, nil
}

func (client *Client) GetProjectTitle(key string) (title string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	body, err := client.Request("GET", "project/"+key, []byte{})
	if err != nil {
		return "", err
	}
	var rawData interface{}
	if err := json.Unmarshal(body, &rawData); err != nil {
		return "", err
	}
	return rawData.(map[string]interface{})["name"].(string), nil
}

func (client *Client) Comment(issue string, msg string) error {
	type comment struct {
		Data string `json:"body"`
	}

	body, err := json.Marshal(comment{Data: msg})
	if err != nil {
		return err
	}
	_, err = client.Request("POST", "issue/"+issue+"/comment", body)
	if err != nil {
		return err
	}

	return nil
}

func (client *Client) Request(method string, path string, body []byte) (
	[]byte, error) {
	buffer := bytes.NewBuffer(body)

	req, err := http.NewRequest(method, client.baseUrl.String()+path, buffer)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	req.SetBasicAuth(client.user, client.pass)

	resp, err := client.res.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 404 {
		return nil, Error{
			StatusCode: resp.StatusCode, Status: resp.Status,
			Message: "Not Found"}
	}

	if (resp.StatusCode == 400) || (resp.StatusCode >= 500) {
		return nil, Error{StatusCode: resp.StatusCode,
			Status: resp.Status, Message: string(data)}
	}

	return data, nil
}
