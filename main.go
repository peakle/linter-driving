package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "io/ioutil"
    "net/http"
    "net/url"
)

type ApiResult struct {
    Items []Item `json:"items"`
}

type Item struct {
    CloneURL string `json:"clone_url"`
}

func main() {
    conf, err := InitConfig()
    if err != nil {
        fmt.Println(err)
        return
    }

    projects, err := getProjects(conf)
    if err != nil {
        fmt.Println(err)
        return
    }

    fmt.Println(projects)
}

func getProjects(config *Config) ([]string, error) {
    var client http.Client
    u, err := url.Parse("https://api.github.com/search/repositories?q=language:go&stars:>500")
    if err != nil {
        return nil, fmt.Errorf("on url.Parse: %s", err)
    }

    resp, err := client.Do(&http.Request{
        Method: http.MethodGet,
        URL:    u,
        Header: map[string][]string{
            "Authorization": {fmt.Sprintf("token %s", config.Token)},
        },
    })
    if err != nil {
        return nil, fmt.Errorf("on github search: %s", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, errors.New("on github search: non 200 status code")
    }

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("on ReadAll: %s", err)
    }

    var res ApiResult
    if err = json.Unmarshal(body, &res); err != nil {
        return nil, fmt.Errorf("on Unmarshal: %s", err)
    }

    projects := make([]string, 0, len(res.Items))
    for _, r := range res.Items {
        projects = append(projects, r.CloneURL)
    }

    return projects, nil
}
