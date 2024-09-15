package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/urfave/cli/v2"
)

func NumGroupsValidate(searchRegexp *regexp.Regexp, replace string) error {
	numGroups := searchRegexp.NumSubexp()
	numPlaces := strings.Count(replace, "%s")

	if numGroups != numPlaces {
		return fmt.Errorf("number of search regexp groups must be equal to the number of placeholder in replace string")
	}

	return nil
}

func Replace(path string, searchRegexp *regexp.Regexp, replace string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsPermission(err) {
			return 0, fmt.Errorf("permission denied")
		}
		return 0, err
	}

	numReplaces := 0

	replaceFunc := func(subString []byte) []byte {
		numReplaces++

		groups := searchRegexp.FindStringSubmatch(string(subString))

		var args []any
		for i, group := range groups {
			if i != 0 {
				args = append(args, group)
			}
		}

		res := fmt.Sprintf(replace, args...)

		return []byte(res)
	}

	replacedData := searchRegexp.ReplaceAllFunc(data, replaceFunc)

	err = os.WriteFile(path, replacedData, 0666)
	if err != nil {
		if os.IsPermission(err) {
			return 0, fmt.Errorf("permission denied")
		}
		return 0, err
	}

	return numReplaces, nil
}

func Execute(cctx *cli.Context) error {
	if cctx.Args().Len() < 2 {
		return fmt.Errorf("search regex and replace string must be specified")
	}

	path := cctx.String("path")
	pathRegexp, err := regexp.Compile(path)
	if err != nil {
		return fmt.Errorf("invalid regular expression")
	}

	search := cctx.Args().Get(0)
	searchRegexp, err := regexp.Compile(search)
	if err != nil {
		return fmt.Errorf("invalid regular expression")
	}

	replace := cctx.Args().Get(1)

	err = NumGroupsValidate(searchRegexp, replace)
	if err != nil {
		return err
	}

	paths := []string{}

	if cctx.Bool("r") {
		err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				paths = append(paths, path)
			}
			return nil
		})
		if err != nil {
			return err
		}
	} else {
		entries, err := os.ReadDir("./")
		if err != nil {
			return err
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				paths = append(paths, entry.Name())
			}
		}
	}

	matchingPaths := []string{}
	wdPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory")
	}

	for _, path := range paths {
		fullPath := filepath.Join(wdPath, path)
		if match := pathRegexp.Find([]byte(fullPath)); match != nil {
			matchingPaths = append(matchingPaths, path)
		}
	}

	var goodResults []string
	var badResults []string

	replacesCount := 0
	affectedCount := 0

	for _, path := range matchingPaths {
		numReplaces, err := Replace(path, searchRegexp, replace)
		if err != nil {
			badResults = append(badResults, fmt.Sprintf("%s - error (%s)", path, err))
		} else {
			goodResults = append(goodResults, fmt.Sprintf("%s - %d replaces", path, numReplaces))
		}

		replacesCount += numReplaces
		if numReplaces != 0 {
			affectedCount++
		}
	}

	if !cctx.Bool("s") {
		fmt.Printf("\n\033[1mSummary\033[0m:\nMatching files: %d\nFiles affected: %d\nTotal replaces: %d\nErrors: %d\n\n", len(matchingPaths), affectedCount, replacesCount, len(badResults))

		if cctx.Bool("d") && len(matchingPaths) != 0 {
			fmt.Println("\033[1mChanges:\033[0m")

			for _, res := range goodResults {
				fmt.Println(res)
			}

			for _, res := range badResults {
				fmt.Println(res)
			}

			fmt.Println("")
		}
	}

	return nil
}

func main() {
	app := &cli.App{
		Name:                   "mfrep",
		Usage:                  "Tool for quick automated editing of file contents.",
		Action:                 Execute,
		Args:                   true,
		ArgsUsage:              "<search_regexp> <replace_string>",
		UseShortOptionHandling: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "path",
				Aliases:     []string{"p"},
				Value:       ".*",
				DefaultText: ".*",
				Usage:       "Regular expression to specify in which files the replace should be made.",
			},
			&cli.BoolFlag{
				Name:        "details",
				Aliases:     []string{"d"},
				Value:       false,
				DefaultText: "false",
				Usage:       "Whenever the output should provide additional summary. (list affected files)",
			},
			&cli.BoolFlag{
				Name:        "silent",
				Aliases:     []string{"s"},
				Value:       false,
				DefaultText: "false",
				Usage:       "Whenever the output should provide no summary.",
			},
			&cli.BoolFlag{
				Name:        "recursive",
				Aliases:     []string{"r"},
				Value:       false,
				DefaultText: "false",
				Usage:       "Whenever the files in subdirectories should be affected.",
			},
		},
		CustomAppHelpTemplate: `NAME:
	mfrep - Tool for quick automated editing of file contents.

USAGE:
	mfrep [options] <search_regexp> <replace_string>

OPTIONS:
	--path <path_regexp>, -p <path_regexp>  Regular expression to specify which files should be affected.

	--details, -d           Whenever the output should list affected files. (default: false)
	--silent, -s            Whenever the output should provide no summary. (default: false)
	--recursive, -r         Whenever the files in subdirectories should be affected too. (default: false)

	--help, -h              Show help
`,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(fmt.Errorf("mfrp: %w", err))
		return
	}
}
