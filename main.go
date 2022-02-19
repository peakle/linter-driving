package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "io/ioutil"
    "net/http"
    "net/url"
    "os"
    "os/exec"
    "path/filepath"
    "sync"
    "time"
)

const binaryName = "linter.exe"

type ApiResult struct {
    Items []Item `json:"items"`
}

type Item struct {
    CloneURL string `json:"clone_url"`
}

func main() {
    startTime := time.Now()
    defer fmt.Println("finish: executed in:", time.Since(startTime))

    conf, err := InitConfig()
    if err != nil {
        fmt.Println(err)
        return
    }

    cloneURLs, err := getProjects(conf)
    if err != nil {
        fmt.Println(err)
        return
    }

    var wg *sync.WaitGroup

    if err = buildLinter(conf); err != nil {
        fmt.Println("on buildLinter:", err)
        return
    }

    wg.Add(len(cloneURLs))
    for _, p := range cloneURLs {
        go func(project string) {
            defer wg.Done()
            if err := gitClone(conf, project); err != nil {
                fmt.Printf("on gitClone: %s: %s \n", project, err)
                return
            }
        }(p)
    }
    wg.Wait()

    projects, err := os.ReadDir(conf.ProjectsDir)
    if err != nil {
        fmt.Println("on os.ReadDir:", err)
        return
    }

    wg.Add(len(projects))
    for _, p := range projects {
        go func(project string) {
            defer wg.Done()
            if err := runLinter(conf, project); err != nil {
                fmt.Println("on runLinter:", project, err)
            }
        }(p.Name())
    }
    wg.Wait()
}

func runLinter(conf *Config, project string) error {
    tmpDir := os.TempDir()

    args := conf.LinterArgs
    args = append(args, project)

    out, err := exec.Command(filepath.Join(tmpDir, binaryName), args...).CombinedOutput()
    if err != nil {
        return fmt.Errorf("%s: %s", out, err)
    }

    return nil
}

func buildLinter(conf *Config) error {
    tmpDir := os.TempDir()
    args := []string{"build", "-o", filepath.Join(tmpDir, binaryName), conf.LinterCloneURL}
    out, err := exec.Command("go", args...).CombinedOutput()
    if err != nil {
        return fmt.Errorf("%s: %s", err, out)
    }

    return nil
}

func gitClone(config *Config, project string) error {
    out, err := exec.Command("git", "clone", project, config.ProjectsDir).CombinedOutput()
    if err != nil {
        return fmt.Errorf("%s: %s", out, err)
    }

    return nil
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
