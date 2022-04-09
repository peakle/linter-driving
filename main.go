package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type apiResult struct {
	Items []item `json:"items"`
}

type item struct {
	CloneURL string `json:"clone_url"`
}

type project struct {
	Dir      string
	Name     string
	CloneURL string
}

func main() {
	startTime := time.Now()
	defer fmt.Println("finish: executed in:", time.Since(startTime))

	conf, err := initConfig()
	if err != nil {
		fmt.Println(err)
		return
	}

	projects, err := getProjects(conf)
	if err != nil {
		fmt.Println("on getProjects:", err)
		return
	}

	if err = buildLinter(conf); err != nil {
		fmt.Println("on buildLinter:", err)
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(projects))
	for _, p := range projects {
		go func(p project) {
			defer wg.Done()
			if err := gitClone(conf, p); err != nil {
				fmt.Printf("on gitClone: %s: %s \n", p.CloneURL, err)
				return
			}
		}(p)
	}
	wg.Wait()

	dirs, err := os.ReadDir(conf.ProjectsDir)
	if err != nil {
		fmt.Println("on os.ReadDir:", err)
		return
	}

	wg.Add(len(dirs))
	for _, projectDir := range dirs {
		p, ok := projects[projectDir.Name()]
		if !ok {
			continue
		}

		go func(p project) {
			defer wg.Done()
			if err := runLinter(conf, p); err != nil {
				fmt.Println("on runLinter:", p, err)
			}
		}(p)
	}
	wg.Wait()
}

func runLinter(conf *config, p project) error {
	args := conf.LinterArgs

	linterCmd := exec.Command(conf.BinaryName, args...)
	linterCmd.Dir = p.Dir

	out, err := linterCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, out)
	}

	return nil
}

func buildLinter(conf *config) error {
	// nolint:gocritic // TODO
	//tmpDir := os.TempDir() + string(os.PathSeparator) + filepath.Base(conf.LinterCloneURL)
	//defer os.RemoveAll(tmpDir)
	//
	//out, err := exec.Command("git", "clone", conf.LinterCloneURL, tmpDir).CombinedOutput()
	//if err != nil {
	//	return fmt.Errorf("on git clone linter: %s: %s", err, out)
	//}
	//
	//args := []string{"build", "-o", filepath.Join(tmpDir, binaryName), tmpDir + conf.PathToMain}
	//out, err = exec.Command("go", args...).CombinedOutput()
	//if err != nil {
	//	return fmt.Errorf("%s: %s", err, out)
	//}
	//
	return nil
}

func gitClone(config *config, p project) error {
	for _, excludedProject := range config.ExcludedProjects {
		if strings.Contains(p.Name, excludedProject) {
			return nil
		}
	}

	var (
		out []byte
		err error
	)
	if os.IsExist(os.Mkdir(p.Dir, 0755)) {
		cmd := exec.Command("git", "fetch")
		cmd.Dir = p.Dir
		out, err = cmd.CombinedOutput()
	} else {
		out, err = exec.Command("git", "clone", p.CloneURL, p.Dir).CombinedOutput()
	}
	if err != nil {
		return fmt.Errorf("%s: %s", err, out)
	}

	return nil
}

func getProjects(config *config) (map[string]project, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	var client http.Client

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/search/repositories?q=language:go&stars:>500", http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("on create request: %s", err)
	}
	req.Header.Add("Authorization", "token "+config.Token)

	resp, err := client.Do(req)
	defer func() {
		if resp.Body != nil {
			resp.Body.Close()
		}
	}()
	if err != nil {
		return nil, fmt.Errorf("on github search: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("on github search: non 200 status code")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("on ReadAll: %s", err)
	}

	var res apiResult
	if err = json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("on Unmarshal: %s", err)
	}

	projects := make(map[string]project, len(res.Items))
	for _, r := range res.Items {
		name := strings.TrimSuffix(filepath.Base(r.CloneURL), ".git")
		dir := config.ProjectsDir + string(os.PathSeparator) + name

		projects[name] = project{
			Dir:      dir,
			Name:     name,
			CloneURL: r.CloneURL,
		}
	}

	return projects, nil
}
