package cmd

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func init() {
	showCmd.AddCommand(showBootTagsCmd)
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "show app info.",
}

var showBootTagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "show app boot tags.",
	Run: func(cmd *cobra.Command, args []string) {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Name", "Tags"})

		for i, boot := range ServerBootstrappers {
			s := strings.Split(reflect.TypeOf(boot).String(), ".")
			name := s[len(s)-1]
			tags := strings.Join(boot.Tags(), ",")
			table.Append([]string{fmt.Sprintf("%d", i+1), name, tags})
		}
		table.Render()
	},
}
