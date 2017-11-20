package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type Repo struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Owner    struct {
		Login             string `json:"login"`
		ID                int    `json:"id"`
		AvatarURL         string `json:"avatar_url"`
		GravatarID        string `json:"gravatar_id"`
		URL               string `json:"url"`
		HTMLURL           string `json:"html_url"`
		FollowersURL      string `json:"followers_url"`
		FollowingURL      string `json:"following_url"`
		GistsURL          string `json:"gists_url"`
		StarredURL        string `json:"starred_url"`
		SubscriptionsURL  string `json:"subscriptions_url"`
		OrganizationsURL  string `json:"organizations_url"`
		ReposURL          string `json:"repos_url"`
		EventsURL         string `json:"events_url"`
		ReceivedEventsURL string `json:"received_events_url"`
		Type              string `json:"type"`
		SiteAdmin         bool   `json:"site_admin"`
	} `json:"owner"`
	Private          bool        `json:"private"`
	HTMLURL          string      `json:"html_url"`
	Description      interface{} `json:"description"`
	Fork             bool        `json:"fork"`
	URL              string      `json:"url"`
	ForksURL         string      `json:"forks_url"`
	KeysURL          string      `json:"keys_url"`
	CollaboratorsURL string      `json:"collaborators_url"`
	TeamsURL         string      `json:"teams_url"`
	HooksURL         string      `json:"hooks_url"`
	IssueEventsURL   string      `json:"issue_events_url"`
	EventsURL        string      `json:"events_url"`
	AssigneesURL     string      `json:"assignees_url"`
	BranchesURL      string      `json:"branches_url"`
	TagsURL          string      `json:"tags_url"`
	BlobsURL         string      `json:"blobs_url"`
	GitTagsURL       string      `json:"git_tags_url"`
	GitRefsURL       string      `json:"git_refs_url"`
	TreesURL         string      `json:"trees_url"`
	StatusesURL      string      `json:"statuses_url"`
	LanguagesURL     string      `json:"languages_url"`
	StargazersURL    string      `json:"stargazers_url"`
	ContributorsURL  string      `json:"contributors_url"`
	SubscribersURL   string      `json:"subscribers_url"`
	SubscriptionURL  string      `json:"subscription_url"`
	CommitsURL       string      `json:"commits_url"`
	GitCommitsURL    string      `json:"git_commits_url"`
	CommentsURL      string      `json:"comments_url"`
	IssueCommentURL  string      `json:"issue_comment_url"`
	ContentsURL      string      `json:"contents_url"`
	CompareURL       string      `json:"compare_url"`
	MergesURL        string      `json:"merges_url"`
	ArchiveURL       string      `json:"archive_url"`
	DownloadsURL     string      `json:"downloads_url"`
	IssuesURL        string      `json:"issues_url"`
	PullsURL         string      `json:"pulls_url"`
	MilestonesURL    string      `json:"milestones_url"`
	NotificationsURL string      `json:"notifications_url"`
	LabelsURL        string      `json:"labels_url"`
	ReleasesURL      string      `json:"releases_url"`
	DeploymentsURL   string      `json:"deployments_url"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
	PushedAt         time.Time   `json:"pushed_at"`
	GitURL           string      `json:"git_url"`
	SSHURL           string      `json:"ssh_url"`
	CloneURL         string      `json:"clone_url"`
	SvnURL           string      `json:"svn_url"`
	Homepage         interface{} `json:"homepage"`
	Size             int         `json:"size"`
	StargazersCount  int         `json:"stargazers_count"`
	WatchersCount    int         `json:"watchers_count"`
	Language         string      `json:"language"`
	HasIssues        bool        `json:"has_issues"`
	HasProjects      bool        `json:"has_projects"`
	HasDownloads     bool        `json:"has_downloads"`
	HasWiki          bool        `json:"has_wiki"`
	HasPages         bool        `json:"has_pages"`
	ForksCount       int         `json:"forks_count"`
	MirrorURL        interface{} `json:"mirror_url"`
	Archived         bool        `json:"archived"`
	OpenIssuesCount  int         `json:"open_issues_count"`
	Forks            int         `json:"forks"`
	OpenIssues       int         `json:"open_issues"`
	Watchers         int         `json:"watchers"`
	DefaultBranch    string      `json:"default_branch"`
	Permissions      struct {
		Admin bool `json:"admin"`
		Push  bool `json:"push"`
		Pull  bool `json:"pull"`
	} `json:"permissions"`
}

var logger = log.New(os.Stdout, "", 0)

const usage = `Usage:
	gh <org/user> <name> [username:password/token]

Example:
	gh user heppu heppu:my-secret-token

`

func init() {
	if err := exec.Command("git", "--version").Run(); err != nil {
		logger.Fatal(err)
	}
}

func main() {
	if len(os.Args) < 3 || len(os.Args) > 4 {
		logger.Fatal(usage)
	}
	if len(os.Args) == 4 && !strings.ContainsRune(os.Args[3], ':') {
		logger.Fatal(usage)
	}

	var category string
	switch os.Args[1] {
	case "org":
		category = "orgs"
	case "user":
		category = "users"
	default:
		logger.Fatal(usage)
	}

	if err := clone(category, os.Args[2]); err != nil {
		logger.Fatal(err)
	}
}

func clone(category, name string) (err error) {
	var repos []Repo
	if repos, err = GetRepos(category, name); err != nil {
		return
	}

	// GetRepos function already makes sure there isn't any special characters
	// in organization's / user's name except '-' which is ok.
	if err = os.Mkdir(name, os.ModePerm); err != nil {
		return
	}
	if err = os.Chdir(name); err != nil {
		return
	}

	errors := make([]error, len(repos))
	wg := sync.WaitGroup{}
	wg.Add(len(repos))

	for i, repo := range repos {
		go func(i int, repo Repo) {
			errors[i] = exec.Command("git", "clone", repo.CloneURL).Run()
			wg.Done()
		}(i, repo)
	}
	wg.Wait()
	for _, err := range errors {
		if err != nil {
			fmt.Println("err")
		}
	}

	if err = os.Chdir(".."); err != nil {
		return
	}
	return
}

func GetRepos(category, name string) (repos []Repo, err error) {
	var req *http.Request
	url := fmt.Sprintf("https://api.github.com/%s/%s/repos", category, name)
	if req, err = http.NewRequest(http.MethodGet, url, nil); err != nil {
		return
	}

	if len(os.Args) == 4 {
		auth := strings.SplitN(os.Args[3], ":", 2)
		req.SetBasicAuth(auth[0], auth[1])
	}

	var res *http.Response
	client := http.Client{}
	if res, err = client.Do(req); err != nil {
		return
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		err = fmt.Errorf("%s", res.Status)
		return
	}

	err = json.NewDecoder(res.Body).Decode(&repos)
	return
}
