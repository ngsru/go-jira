# Intro

go-jira it's package to work with Atlassian Jira written in Go

The package has a low functionality, because it was done for specific tasks.

# Usage

```
// creating a Jira instance:
jira, err := new jira.New("http://jira.local/", "username", "password")
if err != nil {
    //catch
}
[..]

//find issue by key (PROJECT-1234) with some fields (may be custom)
issue, err := jira.GetIssue("PROJECT-1234", []string{"summary"})
if err != nil {
    //catch
    panic(err)
}
fmt.Printf("%+v", issue)

[..]

//and comment 'hello world'
err := jira.Comment("PROJECT-1234", "Hello World!")
if err != nil {
    //catch
}

[..]

//go-jira also can get project title
title, err := jira.GetProjectTitle("PROJECT")
if err != nil {
    //catch
}
fmt.Printf("PROJECT title: %s", title)
```

Soon there will be more functional and will be more complete README.
