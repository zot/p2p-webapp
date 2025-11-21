// CRC: crc-CommandRouter.md, Spec: main.md
package commands

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zot/p2p-webapp/internal/bundle"
)

// CatCmd represents the cat command
var CatCmd = &cobra.Command{
	Use:   "cat FILE",
	Short: "Display contents of a bundled file",
	Long: `Display the contents of a file from the bundled site.
Useful for viewing configuration files and other text files without extracting.

Example:
  p2p-webapp cat p2p-webapp.toml    # view configuration file`,
	Args: cobra.ExactArgs(1),
	RunE: runCat,
}

func runCat(cmd *cobra.Command, args []string) error {
	fileName := args[0]

	// Get bundle reader
	zipReader, err := bundle.GetBundleReader()
	if err != nil {
		return fmt.Errorf("failed to read bundle: %w", err)
	}
	if zipReader == nil {
		return fmt.Errorf("binary is not bundled")
	}

	// Find the file in the bundle (check config/ and html/ directories)
	possiblePaths := []string{
		"config/" + fileName, // config/ directory (for config files)
		"html/" + fileName,   // html/ directory (for site files)
	}

	for _, targetPath := range possiblePaths {
		for _, f := range zipReader.File {
			if f.Name == targetPath && !f.FileInfo().IsDir() {
				// Found the file - read and output contents
				rc, err := f.Open()
				if err != nil {
					return fmt.Errorf("failed to open file: %w", err)
				}
				defer rc.Close()

				// Copy file contents to stdout
				if _, err := io.Copy(cmd.OutOrStdout(), rc); err != nil {
					return fmt.Errorf("failed to read file: %w", err)
				}

				return nil
			}
		}
	}

	// File not found - try to provide helpful error message
	available := []string{}
	for _, f := range zipReader.File {
		if strings.HasPrefix(f.Name, "html/") && !f.FileInfo().IsDir() {
			relPath := strings.TrimPrefix(f.Name, "html/")
			if strings.Contains(relPath, fileName) {
				available = append(available, relPath)
			}
		}
	}

	if len(available) > 0 {
		return fmt.Errorf("file not found: %s\nDid you mean one of these?\n  %s",
			fileName, strings.Join(available, "\n  "))
	}

	return fmt.Errorf("file not found: %s", fileName)
}
