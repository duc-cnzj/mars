package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

// initCmd 是 init 子命令：在缺少 config.yaml 时生成一份默认配置文件。
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "create default config file.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(configExampleFile) > 0 {
			if fileExists("config.yaml") {
				log.Println("config.yaml 文件已存在！")
				return
			}
			if err := os.WriteFile("config.yaml", configExampleFile, 0600); err != nil {
				log.Println("写入 config.yaml 文件失败")
				return
			}
			log.Println("创建成功！")
			return
		}
		log.Println("config example file is empty!")
	},
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
