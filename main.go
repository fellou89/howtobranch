package main

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/oauth2"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/google/go-github/github"
)

// Github struct
type Github struct {
	Owner   string
	Repo    string
	Context context.Context
	Client  *github.Client
	Service *github.GitService
	HeadRef *github.Reference
}

var g *Github

func init() {
	ctx := context.Background()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("CMS_GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	g = &Github{Owner: "fellou89", Repo: "howtobranch", Context: ctx, Client: client, Service: client.Git}
	ref, _, err := g.Service.GetRef(g.Context, g.Owner, g.Repo, "refs/heads/master")
	if err != nil {
		log.Fatal(err)
	}
	g.HeadRef = ref
	fmt.Printf("%+v\n\n", g)
}

// NewBranch function
func (g *Github) NewBranch(branchName string) {
	reference := &github.Reference{
		Ref: github.String("refs/heads/" + branchName),
		Object: &github.GitObject{
			Type: github.String("commit"),
			SHA:  g.HeadRef.Object.SHA,
		},
	}
	ref, _, err := g.Service.CreateRef(g.Context, g.Owner, g.Repo, reference)
	if err != nil {
		log.Fatal(err)
	}
	g.HeadRef = ref
	fmt.Printf("%+v\n\n", g)
}

// MakeCommit function
func (g *Github) MakeCommit(filesChanged []string, commitMessage string) {

	// Create changed file structure
	entries := []github.TreeEntry{}

	for _, fileArg := range filesChanged {
		file, content, err := getFileContent(fileArg)
		if err != nil {
			log.Fatal(err)
		}
		entries = append(entries, github.TreeEntry{Path: github.String(file), Type: github.String("blob"), Content: github.String(string(content)), Mode: github.String("100644")})
	}

	fmt.Printf("%+v\n\n", g)
	tree, _, err := g.Service.CreateTree(g.Context, g.Owner, g.Repo, *g.HeadRef.Object.SHA, entries)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Tree created")

	// Create and push Commit
	parent, _, err := g.Client.Repositories.GetCommit(g.Context, g.Owner, g.Repo, *g.HeadRef.Object.SHA)
	if err != nil {
		log.Fatal(err)
	}
	parent.Commit.SHA = parent.SHA

	commit := &github.Commit{Message: &commitMessage, Tree: tree, Parents: []github.Commit{*parent.Commit}}
	newCommit, _, err := g.Service.CreateCommit(g.Context, g.Owner, g.Repo, commit)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Commit created")

	// Attach the commit to the master branch.
	g.HeadRef.Object.SHA = newCommit.SHA
	ref, _, err := g.Client.Git.UpdateRef(g.Context, g.Owner, g.Repo, g.HeadRef, false)
	if err != nil {
		log.Fatal(err)
	}
	g.HeadRef = ref
	fmt.Println("Ref updated")
}

func getFileContent(fileArg string) (targetName string, b []byte, err error) {
	var localFile string
	files := strings.Split(fileArg, ":")
	switch {
	case len(files) < 1:
		return "", nil, errors.New("empty `-files` parameter")
	case len(files) == 1:
		localFile = files[0]
		targetName = files[0]
	default:
		localFile = files[0]
		targetName = files[1]
	}

	b, err = ioutil.ReadFile(localFile)
	return targetName, b, err
}

// GetFile function
func (g *Github) GetFile(fileSHA string) {
	// *tree.Entries[0].SHA
	blob, _, err := g.Service.GetBlobRaw(g.Context, g.Owner, g.Repo, fileSHA)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(blob))
}

func main() {
	// getTree has to be run before this to have access to entries file shas
	//g.GetFile()

	g.NewBranch("test")
	g.MakeCommit([]string{"main.go"}, "new commit")
}
