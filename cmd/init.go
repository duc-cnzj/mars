package cmd

import (
	"log"
	"os"

	"github.com/DuC-cnZj/mars/pkg/utils"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "生成配置文件",
	Run: func(cmd *cobra.Command, args []string) {
		if len(configExampleFile) > 0 {
			if utils.FileExists("config.yaml") {
				log.Println("config.yaml 文件已存在！")
				return
			}
			if err := os.WriteFile("config.yaml", configExampleFile, 0644); err != nil {
				log.Println("写入 config.yaml 文件失败")
				return
			}
			log.Println("创建成功！")
			return
		}
		log.Println("config example file is empty!")
	},
}
