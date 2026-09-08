package main

import (
	"fmt"
	"os"

	"github.com/nanoteck137/dusk"
	"github.com/nanoteck137/dusk/dev"
	"github.com/nanoteck137/dusk/service"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     fmt.Sprintf("%s [path]", dusk.AppName),
	Version: dusk.Version,
	Short:   "Show how much disk space a directory uses",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		size, err := service.DiskUsage(path)
		if err != nil {
			return err
		}

		cmd.Printf("%s\t%s\n", service.HumanSize(size), path)
		return nil
	},
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func init() {
	rootCmd.SetVersionTemplate(dusk.VersionTemplate(dusk.AppName))
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		dev.Println(err)
		os.Exit(1)
	}
}
