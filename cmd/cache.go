package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/lesomnus/clade/cmd/config"
	"github.com/lesomnus/clade/registry"
	"github.com/lesomnus/xli"
	"github.com/lesomnus/xli/arg"
	"github.com/lesomnus/xli/flg"
	"github.com/lesomnus/z"
)

func NewCmdCache() *xli.Command {
	return &xli.Command{
		Name:  "cache",
		Brief: "inspect and manage the registry metadata cache",

		Commands: []*xli.Command{
			newCmdCacheList(),
			newCmdCacheRemove(),
		},

		Handler: xli.RequireSubcommand(),
	}
}

func newCmdCacheList() *xli.Command {
	return &xli.Command{
		Name:  "ls",
		Brief: "list cached repositories, or the cached tags of one repository",

		Args: arg.Args{
			&arg.String{Name: "repo", Optional: true, Brief: "repository whose cached tags to print"},
		},

		Handler: xli.OnRun(func(ctx context.Context, cmd *xli.Command, next xli.Next) error {
			c := use_config.Must(ctx)
			fc, err := openFileCache(c)
			if err != nil {
				return err
			}

			entries, err := fc.Entries()
			if err != nil {
				return z.Err(err, "read cache")
			}

			if repo, ok := arg.Get[string](cmd, "repo"); ok {
				return printRepoTags(cmd, entries, repo)
			}
			return printRepoList(cmd, fc.Dir(), entries)
		}),
	}
}

func newCmdCacheRemove() *xli.Command {
	return &xli.Command{
		Name:  "rm",
		Brief: "remove cached entries for repositories, or the entire cache",

		Args: arg.Args{
			&arg.RestStrings{Name: "repo", Brief: "repositories whose cached entries to remove"},
		},
		Flags: flg.Flags{
			&flg.Switch{Name: "all", Brief: "remove every cached entry"},
		},

		Handler: xli.OnRun(func(ctx context.Context, cmd *xli.Command, next xli.Next) error {
			c := use_config.Must(ctx)
			fc, err := openFileCache(c)
			if err != nil {
				return err
			}

			all := false
			flg.VisitP(cmd, "all", &all)
			repos, _ := arg.Get[[]string](cmd, "repo")

			if all {
				if len(repos) > 0 {
					return fmt.Errorf("--all takes no repository arguments")
				}
				n, err := fc.Clear()
				if err != nil {
					return z.Err(err, "clear cache")
				}
				cmd.Printf("removed %d cache %s\n", n, plural(n, "entry", "entries"))
				return nil
			}
			if len(repos) == 0 {
				return fmt.Errorf("specify one or more repositories to remove, or --all")
			}

			entries, err := fc.Entries()
			if err != nil {
				return z.Err(err, "read cache")
			}

			removed := 0
			for _, repo := range repos {
				for _, key := range repoKeys(entries, repo) {
					ok, err := fc.Remove(key)
					if err != nil {
						return err
					}
					if ok {
						removed++
					}
				}
			}
			cmd.Printf("removed %d cache %s\n", removed, plural(removed, "entry", "entries"))
			return nil
		}),
	}
}

// openFileCache opens the on-disk metadata cache for management. It errors when
// no cache directory is resolvable; an in-memory cache has nothing to manage.
func openFileCache(c *config.Config) (*registry.FileCache, error) {
	dir := cacheDir(c)
	if dir == "" {
		return nil, fmt.Errorf("no cache directory configured")
	}
	return registry.NewFileCache(dir)
}

// printRepoList prints one row per repository that has a cached tag listing.
func printRepoList(cmd *xli.Command, dir string, entries []registry.CacheEntry) error {
	type row struct {
		repo    string
		tags    int
		expires string
	}

	rows := make([]row, 0, len(entries))
	width := len("REPOSITORY")
	for _, e := range entries {
		repo, ok := strings.CutPrefix(e.Key, registry.KeyTags)
		if !ok {
			continue
		}

		var tags []string
		_ = json.Unmarshal(e.Val, &tags)
		rows = append(rows, row{repo: repo, tags: len(tags), expires: expiryText(e)})
		if len(repo) > width {
			width = len(repo)
		}
	}

	if len(rows) == 0 {
		cmd.Printf("no cached repositories in %s\n", dir)
		return nil
	}

	sort.Slice(rows, func(i, j int) bool { return rows[i].repo < rows[j].repo })

	cmd.Printf("%-*s  %5s  %s\n", width, "REPOSITORY", "TAGS", "EXPIRES")
	for _, r := range rows {
		cmd.Printf("%-*s  %5d  %s\n", width, r.repo, r.tags, r.expires)
	}
	return nil
}

// printRepoTags prints the cached tags of repo, one per line.
func printRepoTags(cmd *xli.Command, entries []registry.CacheEntry, repo string) error {
	for _, e := range entries {
		if e.Key != registry.KeyTags+repo {
			continue
		}

		var tags []string
		if err := json.Unmarshal(e.Val, &tags); err != nil {
			return z.Err(err, "decode cached tags")
		}
		sort.Strings(tags)
		for _, t := range tags {
			cmd.Println(t)
		}
		return nil
	}

	return fmt.Errorf("no cached tags for %q", repo)
}

// repoKeys returns the cache keys that belong to repo: its tag listing and any
// per-tag image metadata (keyed KeyStat+"repo:tag"). The trailing ":" keeps the
// metadata prefix from matching a different repository that shares a name stem.
func repoKeys(entries []registry.CacheEntry, repo string) []string {
	stat_prefix := registry.KeyStat + repo + ":"

	keys := []string{registry.KeyTags + repo}
	for _, e := range entries {
		if strings.HasPrefix(e.Key, stat_prefix) {
			keys = append(keys, e.Key)
		}
	}
	return keys
}

// expiryText renders an entry's expiry relative to now for display.
func expiryText(e registry.CacheEntry) string {
	switch {
	case e.ExpiresAt.IsZero():
		return "never"
	case e.Expired:
		return "expired"
	default:
		return "in " + time.Until(e.ExpiresAt).Round(time.Second).String()
	}
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}
