package cmd

import (
	"bytes"
	"io/ioutil"
	"log"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		parse, err := godotenv.Parse(strings.NewReader(`APP_NAME=Laravel
APP_ENV=local
APP_KEY=base64:WTiu09qi7cDRl930LqKzb2xpzfZwO82snkmewcUXD/g=
APP_DEBUG=true
APP_URL=http://app.test

LOG_CHANNEL=stack
LOG_LEVEL=debug

DB_CONNECTION=mysql
DB_HOST=127.0.0.1
DB_PORT=3306
DB_DATABASE=app
DB_USERNAME=root
DB_PASSWORD=

BROADCAST_DRIVER=log
CACHE_DRIVER=file
QUEUE_CONNECTION=sync
SESSION_DRIVER=file
SESSION_LIFETIME=120

MEMCACHED_HOST=127.0.0.1

REDIS_HOST=127.0.0.1
REDIS_PASSWORD=null
REDIS_PORT=6379

MAIL_MAILER=smtp
MAIL_HOST=mailhog
MAIL_PORT=1025
MAIL_USERNAME=null
MAIL_PASSWORD=null
MAIL_ENCRYPTION=null
MAIL_FROM_ADDRESS=null
MAIL_FROM_NAME="${APP_NAME}"

AWS_ACCESS_KEY_ID=
AWS_SECRET_ACCESS_KEY=
AWS_DEFAULT_REGION=us-east-1
AWS_BUCKET=

PUSHER_APP_ID=
PUSHER_APP_KEY=
PUSHER_APP_SECRET=
PUSHER_APP_CLUSTER=mt1

MIX_PUSHER_APP_KEY="${PUSHER_APP_KEY}"
MIX_PUSHER_APP_CLUSTER="${PUSHER_APP_CLUSTER}"`))
		if err != nil {
			log.Fatal(err)
		}
		bb := &bytes.Buffer{}
		encoder := yaml.NewEncoder(bb)
		encoder.Encode(map[string]interface{}{
			"data": parse,
		})

		if err := ioutil.WriteFile("/tmp/duc.yaml", bb.Bytes(), 0644); err != nil {
			log.Fatal(err)
		}
		//log.Println(bb.String())
	},
}
