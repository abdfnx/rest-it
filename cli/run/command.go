/*
Restofile exapmle

request {
   method "POST"
   url "https://api.spacex.land/graphql"
   contentType "application/graphql"
}

body {
   openBodyEditor "true"
   # if `openBodyEditor` prop is false or not set
   readFrom "examples/spacex.gql"
}

auth {
   type "bearer"
   token: "MY_TOKEN"
   # or basic auth
   type "basic"
   username: "USERNAME"
   password: "P@$$w0rd"
   # to use from env variable
   password "env:MY_PASSWORD"
}
*/

package run

import (
	"fmt"
	"os"
	"log"
	"io/ioutil"
	"strings"

	"github.com/abdfnx/resto/tools"
	"github.com/abdfnx/resto/core/api"
	"github.com/abdfnx/resto/core/editor"
	"github.com/abdfnx/resto/core/editor/runtime"
	"github.com/abdfnx/resto/core/options"

	tcell "github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"
	"github.com/pkg/errors"
)

func RunCMD() *cobra.Command {
	opts := options.RunCommandOptions{
		Path: "",
		ShowAll: false,
	}

	cmd := &cobra.Command{
		Use:   "run  [flags]",
		Short: "Send a request from Restofile",
		Long:  `Send a request via file "Restofile"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(&opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Path, "file", "f", "", "Path to Restofile (Default: PATH/Restofile)")
	cmd.Flags().BoolVarP(&opts.ShowAll, "all", "a", false, "Show all response headers & status")

	return cmd
}

func run(opts *options.RunCommandOptions) error {
	path := "./Restofile"

	fn := tools.CLIRequestFile("txt")
	cType := ""

	method := ""
	url := ""
	contentType := ""

	openBodyEditor := false
	readFrom := ""
	content := ""

	authType := ""
	token := ""
	username := ""
	password := ""

	if opts.Path != "" {
		path = opts.Path
	}

	data, err := ioutil.ReadFile(path)

	if err != nil {
		return errors.Errorf("Error: %v", err)
	}

	if strings.Contains(string(data), "request {") || strings.Contains(string(data), "request{") {
		if strings.Contains(string(data), "method") {
			method = strings.TrimSpace(strings.Split(string(data), "method")[1])
			method = strings.TrimSpace(strings.Split(method, "\"")[1])

			if strings.Contains(method, "env:") {
				method = strings.TrimSpace(strings.Split(method, "env:")[1])
				method = os.Getenv(method)
			}
		}

		if strings.Contains(string(data), "url") {
			url = strings.TrimSpace(strings.Split(string(data), "url")[1])
			url = strings.TrimSpace(strings.Split(url, "\"")[1])

			if strings.Contains(url, "env:") {
				url = strings.TrimSpace(strings.Split(url, "env:")[1])
				url = os.Getenv(url)
			}
		}

		if strings.Contains(string(data), "contentType") {
			contentType = strings.TrimSpace(strings.Split(string(data), "contentType")[1])
			contentType = strings.TrimSpace(strings.Split(contentType, "\"")[1])

			if strings.Contains(contentType, "env:") {
				contentType = strings.TrimSpace(strings.Split(contentType, "env:")[1])
				contentType = os.Getenv(contentType)
			}

			if contentType != "" {
				if contentType == "application/json" || contentType == "json" {
					fn = tools.CLIRequestFile("json")
					cType = "application/json"
				} else if contentType == "application/graphql" || contentType == "graphql" {
					fn = tools.CLIRequestFile("graphql")
					cType = "application/graphql"
				} else if contentType == "application/xml" || contentType == "xml" {
					fn = tools.CLIRequestFile("xml")
					cType = "application/xml"
				} else if contentType == "text/html" || contentType == "html" {
					fn = tools.CLIRequestFile("html")
					cType = "text/html"
				} else {
					fn = tools.CLIRequestFile("txt")
					cType = "text/plain"
				}
			}
		}
	}

	if strings.Contains(string(data), "body {") || strings.Contains(string(data), "body{") {
		if strings.Contains(string(data), "openBodyEditor") {
			openBodyEditorValue := strings.TrimSpace(strings.Split(string(data), "openBodyEditor")[1])
			openBodyEditorValue = strings.TrimSpace(strings.Split(openBodyEditorValue, "\"")[1])

			if strings.Contains(openBodyEditorValue, "env:") {
				openBodyEditorValue = strings.TrimSpace(strings.Split(openBodyEditorValue, "env:")[1])
				openBodyEditorValue = os.Getenv(openBodyEditorValue)
			}

			if openBodyEditorValue == "yes" || openBodyEditorValue == "true" {
				openBodyEditor = true
			} else {
				openBodyEditor = false
			}

			if openBodyEditor {
				fileContent, err := ioutil.ReadFile(fn)
				buffer := editor.NewBufferFromString(string(fileContent), fn)
				if err != nil {
					log.Fatalf("could not read %v: %v", fn, err)
				}

				var colorscheme editor.Colorscheme
				if railscast := runtime.Files.FindFile(editor.RTColorscheme, "railscast"); railscast != nil {
					if data, err := railscast.Data(); err == nil {
						colorscheme = editor.ParseColorscheme(string(data))
					}
				}

				bodyEditor := editor.NewView(buffer)
				bodyEditor.SetRuntimeFiles(runtime.Files)
				bodyEditor.SetColorscheme(colorscheme)

				app := tview.NewApplication()
				bodyEditor.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
					switch event.Key() {
						case tcell.KeyCtrlS:
							tools.SaveBuffer(buffer, fn)
							app.Stop()
							return nil

						case tcell.KeyCtrlQ:
							app.Stop()
							return nil
					}

					return event
				})

				app.SetRoot(bodyEditor, true)

				if err := app.Run(); err != nil {
					log.Fatalf("%v", err)
				}

				b, e := os.Open(fn)

				if e != nil {
					fmt.Println(e)
				}

				defer b.Close()

				body, err := ioutil.ReadAll(b)

				if err != nil {
					panic(err)
				}

				content = string(body)
			}
		}

		if strings.Contains(string(data), "readFrom") {
			readFrom = strings.TrimSpace(strings.Split(string(data), "readFrom")[1])
			readFrom = strings.TrimSpace(strings.Split(readFrom, "\"")[1])

			if strings.Contains(readFrom, "env:") {
				readFrom = strings.TrimSpace(strings.Split(readFrom, "env:")[1])
				readFrom = os.Getenv(readFrom)
			}

			body, err := ioutil.ReadFile(readFrom)
			if err != nil {
				return err
			}

			content = string(body)
		}

		if !openBodyEditor && readFrom == "" {
			if method == "POST" || method == "PUT" || method == "PATCH" || method == "DELETE" {
				return errors.Errorf("Error: body is required")
			}
		}
	}

	if strings.Contains(string(data), "auth {") || strings.Contains(string(data), "auth{") {
		if strings.Contains(string(data), "type") {
			authType = strings.TrimSpace(strings.Split(string(data), "type")[1])
			authType = strings.TrimSpace(strings.Split(authType, "\"")[1])

			if strings.Contains(authType, "env:") {
				authType = strings.TrimSpace(strings.Split(authType, "env:")[1])
				authType = os.Getenv(authType)
			}
		}

		if authType == "bearer" {
			if strings.Contains(string(data), "token") {
				token = strings.TrimSpace(strings.Split(string(data), "token")[1])
				token = strings.TrimSpace(strings.Split(token, "\"")[1])

				if strings.Contains(token, "env:") {
					token = strings.TrimSpace(strings.Split(token, "env:")[1])
					token = os.Getenv(token)
				}
			}
		} else if authType == "basic" {
			if strings.Contains(string(data), "username") {
				username = strings.TrimSpace(strings.Split(string(data), "username")[1])
				username = strings.TrimSpace(strings.Split(username, "\"")[1])

				if strings.Contains(username, "env:") {
					username = strings.TrimSpace(strings.Split(username, "env:")[1])
					username = os.Getenv(username)
				}
			}

			if strings.Contains(string(data), "password") {
				password = strings.TrimSpace(strings.Split(string(data), "password")[1])
				password = strings.TrimSpace(strings.Split(password, "\"")[1])

				if strings.Contains(password, "env:") {
					password = strings.TrimSpace(strings.Split(password, "env:")[1])
					password = os.Getenv(password)
				}
			}
		}
	}

	if method == "GET" || method == "HEAD" {
		respone, status, headers, err := api.BasicGet(
			url,
			method,
			authType,
			token,
			username,
			password,
			true,
			0,
			nil,
		)

		if err != nil {
			return err
		}

		if opts.ShowAll {
			fmt.Println(headers)
			fmt.Println("")
			fmt.Println(status)
		}
	
		fmt.Println("\n" + respone)
	} else if method == "POST" || method == "PUT" || method == "PATCH" || method == "DELETE" {
		respone, status, headers, err :=
			api.BasicRequestWithBody(
				url,
				method,
				cType,
				content,
				authType,
				token,
				username,
				password,
				true,
				0,
				nil,
			)

		if err != nil {
			return err
		}
		
		if opts.ShowAll {
			fmt.Println(headers)
			fmt.Println("")
			fmt.Println(status)
		}

		fmt.Println("\n" + respone)
	}

	return nil
}
