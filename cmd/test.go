package cmd

import (
	"encoding/json"
	"log"

	"github.com/spf13/cobra"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		type ProjectStoreInput struct {
			NamespaceId int `uri:"namespace_id" json:"namespace_id"`

			Name            string `json:"name"`
			GitlabProjectId int    `json:"gitlab_project_id"`
			GitlabBranch    string `json:"gitlab_branch"`
			GitlabCommit    string `json:"gitlab_commit"`
			Config          string `json:"config"`
		}
		//config: ""
		//gitlab_branch: "master"
		//gitlab_commit: "89f732c60f92b2514a7967ccfba8be10390921c0"
		//gitlab_project_id: 21409590
		//name: "Defer"
		p := ProjectStoreInput{
			NamespaceId:     9,
			Name:            "Defer",
			GitlabProjectId: 21409590,
			GitlabBranch:    "master",
			GitlabCommit:    "89f732c60f92b2514a7967ccfba8be10390921c0",
			Config:          "",
		}
		marshal, _ := json.Marshal(&p)
		log.Println(string(marshal))
	},
}
