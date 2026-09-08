package main

import (
	"fmt"
	"os"
	"sort"

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

		if showExtensions {
			return printExtensions(cmd, path)
		}

		if showLargest {
			return printLargestFiles(cmd, path)
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

var showExtensions bool
var showLargest bool

func init() {
	rootCmd.SetVersionTemplate(dusk.VersionTemplate(dusk.AppName))
	rootCmd.Flags().BoolVarP(&showExtensions, "extensions", "e", false, "show space used per file extension")
	rootCmd.Flags().BoolVarP(&showLargest, "largest", "l", false, "show some of the largest files")
}

func printLargestFiles(cmd *cobra.Command, path string) error {
	files, err := service.LargestFiles(path, 10)
	if err != nil {
		return err
	}

	for _, f := range files {
		cmd.Printf("%s\t%s\n", service.HumanSize(f.Size), f.Path)
	}

	return nil
}

func printExtensions(cmd *cobra.Command, path string) error {
	usage, err := service.ExtensionUsage(path)
	if err != nil {
		return err
	}

	exts := make([]string, 0, len(usage))
	for ext := range usage {
		exts = append(exts, ext)
	}

	sort.Slice(exts, func(i, j int) bool {
		if exts[i] == "" {
			return false
		}
		if exts[j] == "" {
			return true
		}

		if usage[exts[i]] != usage[exts[j]] {
			return usage[exts[i]] > usage[exts[j]]
		}

		return exts[i] < exts[j]
	})

	for _, ext := range exts {
		label := ext
		if label == "" {
			label = "(no extension)"
		}

		cmd.Printf("%s\t%s\n", service.HumanSize(usage[ext]), label)
	}

	var total int64
	for _, size := range usage {
		total += size
	}

	cmd.Printf("%s\tTOTAL\n", service.HumanSize(total))

	return nil
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		dev.Println(err)
		os.Exit(1)
	}
}
