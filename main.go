package jira

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	ErrNotFound = errors.New("Not found")
)

type Error struct {
	StatusCode int
	Status     string
	Msg        string
}

func (e Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Status, e.Msg)
}

type Issue struct {
	Id      string
	Key     string
	Summary string
	Project string
	Data    map[string]interface{}
}

type Jira struct {
	baseUrl *url.URL
	user    string
	pass    string
	client  *http.Client
}

func New(jiraUrl string, user string, pass string, timeout time.Duration) (
	*Jira,
	error,
) {
	baseUrl, err := url.Parse(jiraUrl)
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{Transport: &http.Transport{
		Dial: func(proto, addr string) (net.Conn, error) {
			return net.DialTimeout(proto, addr, timeout)
		},
	}}

	j := &Jira{
		baseUrl: baseUrl,
		user:    user,
		pass:    pass,
		client:  httpClient,
	}

	return j, nil
}

func (j *Jira) GetIssue(key string, fields []string) (issue *Issue, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	response, err := j.Request("GET",
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

func (j *Jira) GetProjectTitle(key string) (title string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	b, err := j.Request("GET", "project/"+key, []byte{})
	if err != nil {
		return "", err
	}
	var f interface{}
	if err := json.Unmarshal(b, &f); err != nil {
		return "", err
	}
	return f.(map[string]interface{})["name"].(string), nil
}

func (j *Jira) Comment(issue string, msg string) (err error) {
	type comment struct {
		Data string `json:"body"`
	}

	body, err := json.Marshal(comment{Data: msg})
	if err != nil {
		return err
	}
	_, err = j.Request("POST", "issue/"+issue+"/comment", body)
	if err != nil {
		return err
	}

	return nil
}

func (j *Jira) Request(method string, path string, body []byte) ([]byte, error) {
	b := bytes.NewBuffer(body)

	req, err := http.NewRequest(method, j.baseUrl.String()+path, b)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	req.SetBasicAuth(j.user, j.pass)

	resp, err := j.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 404 {
		return nil, ErrNotFound
	}
	if resp.StatusCode >= 400 {
		return nil, Error{StatusCode: resp.StatusCode, Status: resp.Status, Msg: string(data)}
	}
	return data, nil
}
