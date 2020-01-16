/*
@Time 2019-07-26 09:49
@Author ZH

*/
package main

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/zhcppy/github-console/console"
	"github.com/zhcppy/github-console/github"
	"runtime"
)

// application's name
const name = "github-console"

var (
	// application's version string
	Version = "v0.0.1"
	// git commit hash
	Commit = "nil"
)

var rootCmd = &cobra.Command{
	Use:           name,
	Short:         "GitHub command-line console implemented by Golang.",
	Long:          "About more at https://github.com/zhcppy/githubcli.",
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		user, err := github.NewUser(cmd.Flag("token").Value.String())
		if err != nil {
			return err
		}
		ctx, cancel := context.WithCancel(context.Background())
		csl := console.New(ctx, user)
		csl.SetWordCompleter(github.WordCompleter())
		csl.Welcome("Welcome to the Github console!")
		csl.Interactive()
		cancel()
		return csl.Exit()
	},
}

var versionCmd = &cobra.Command{
	Use:           "version",
	Short:         "Print the version number of " + name,
	SilenceErrors: true,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("name: %s\n", name)
		fmt.Printf("commit: %s\n", Commit)
		fmt.Printf("version: %s\n", Version)
		fmt.Printf("golang: %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	},
}

func main() {
	rootCmd.Flags().StringP("token", "t", "", "github token")
	rootCmd.AddCommand(versionCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("\033[1;31m%s\033[0m", fmt.Sprintf("Failed to command execute: %s\n", err.Error()))
	}
}
