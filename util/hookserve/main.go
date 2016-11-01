package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/reyoung/hookserve/hookserve"
	"os"
	"os/exec"
	"github.com/bmatsuo/go-jsontree"
)

func main() {

	app := cli.NewApp()
	app.Name = "hookserve"
	app.Usage = "A small little application that listens for commit / push webhook events from github and runs a specified command\n\n"
	app.Usage += "EXAMPLE:\n"
	app.Usage += "   hookserve --secret=whiskey --port=8888 echo  #Echo back the information provided\n"
	app.Usage += "   hookserve logger -t PushEvent #log the push event to the system log (/var/log/message)"
	app.Version = "1.0"
	app.Author = "Patrick Hayes"
	app.Email = "patrick.d.hayes@gmail.com"

	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:  "port, p",
			Value: 80,
			Usage: "port on which to listen for github webhooks",
		},
		cli.StringFlag{
			Name:  "secret, s",
			Value: "",
			Usage: "Secret for HMAC verification. If not provided no HMAC verification will be done and all valid requests will be processed",
		},
		cli.BoolFlag{
			Name:  "tags, t",
			Usage: "Also execute the command when a tag is pushed",
		},
	}

	app.Action = func(c *cli.Context) {
		server := hookserve.NewServer()
		server.Port = c.Int("port")
		server.Secret = c.String("secret")
		server.IgnoreTags = !c.Bool("tags")
		server.CustomEventHandler["ping"] = func(*jsontree.JsonTree)(interface{}, error) {
			return "pong", nil
		}
		server.GoListenAndServe()

		for event := range server.Events {
			switch event.(type) {
			case hookserve.Event:
				commit := event.(hookserve.Event)
				if args := c.Args(); len(args) != 0 {
				root := args[0]
				rest := append(args[1:], commit.Owner, commit.Repo, commit.Branch, commit.Commit)
				cmd := exec.Command(root, rest...)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Run()
				} else {
					fmt.Println(commit.Owner + " " + commit.Repo + " " + commit.Branch + " " + commit.Commit)
				}
			case string:
				fmt.Println(event.(string))
			}


		}
	}

	app.Run(os.Args)
}
