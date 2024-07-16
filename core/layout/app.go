package layout

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/abdfnx/resto/core/api"
	"github.com/abdfnx/resto/core/editor"
	"github.com/abdfnx/resto/core/editor/runtime"
	"github.com/abdfnx/resto/tools"

	tcell "github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"
)

var (
	// request
	method  string
	httpURL string
	cType   string
	fn      string = tools.RequestFile()

	// respone
	body    string
	respone string
	status  string

	// auth
	authType       string
	requestHeaders string

	// headers
	headersCount int = 0
)

func Layout(version string) {
	app := tview.NewApplication()
	flex := tview.NewFlex()

	helpPage := tview.NewPages()
	updatePage := tview.NewPages()

	helpText := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetChangedFunc(func() {
			app.Draw()
		})
	updateText := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetChangedFunc(func() {
			app.Draw()
		})

	help := `
		Welcome to Resto!

		resto is a cli app can send pretty HTTP & API requests from your terminal.

		Shortcuts:
			- Ctrl+P: Open Resto Panel
			- Ctrl+W: Open Help Guide
			- Ctrl+E: Open Settings
			- Ctrl+S: Save Request Body
			- Ctrl+U: Update Your Resto
			- Ctrl+Q: Quit
	`

	update := `
		How to update Resto?

		first quit from Resto, then:

		1. if you install it from script, then run the script again to update
		2. if you install it from homebrew, then run 'brew upgrade resto' to update
		3. do you install resto from go install, run 'go get -u github.com/abdfnx/resto' to update
		4. do you get it from github cli, run 'gh extension upgrade abdfnx/resto' to upgrade
	`

	fmt.Fprintf(helpText, "%s ", help)
	fmt.Fprintf(updateText, "%s ", update)

	helpPage.AddAndSwitchToPage("help", tview.NewGrid().
		SetColumns(30, 0, 30).
		SetRows(3, 0, 3).
		AddItem(helpText, 1, 1, 1, 1, 0, 0, true), true).
		ShowPage("main")

	updatePage.AddAndSwitchToPage("update", tview.NewGrid().
		SetRows(3, 0, 3).
		SetColumns(30, 0, 30).
		SetBorders(true).
		AddItem(updateText, 1, 1, 1, 1, 0, 0, false), true).
		ShowPage("main")

	// forms
	authForm := tview.NewForm()
	headersForm := tview.NewForm()
	requestForm := tview.NewForm()

	// request inputs
	urlField := tview.NewInputField().
		SetLabel("URL").
		SetFieldWidth(32).
		SetPlaceholder("URL")

	requestMethods := tview.NewDropDown().
		SetLabel("Request Method").
		SetOptions([]string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"HEAD",
		}, func(option string, optionIndex int) {
			method = option
		}).SetCurrentOption(0)

	contentType := tview.NewDropDown().
		SetLabel("Content Type").
		SetOptions([]string{
			"none",
			"application/json",
			"application/graphql",
			"application/xml",
			"text/html",
			"text/plain",
		}, func(option string, optionIndex int) {
			cType = option
		}).SetCurrentOption(0)

	// request body
	content, err := ioutil.ReadFile(fn)
	buffer := editor.NewBufferFromString(string(content), fn)
	if err != nil {
		log.Fatalf("could not read %v: %v", fn, err)
	}

	settingsContent, err := ioutil.ReadFile(tools.SettingsFile())
	bufferSettings := editor.NewBufferFromString(string(settingsContent), tools.SettingsFile())
	if err != nil {
		log.Fatalf("could not read %v: %v", tools.SettingsFile(), err)
	}

	var colorscheme editor.Colorscheme

	vs := gjson.Get(tools.SettingsContent(), "rs_settings.request_body.theme")
	tm := ""

	if vs.Exists() {
		tm = vs.String()
	} else {
		tm = "railscast"
	}

	if theme := runtime.Files.FindFile(editor.RTColorscheme, tm); theme != nil {
		if data, err := theme.Data(); err == nil {
			colorscheme = editor.ParseColorscheme(string(data))
		}
	}

	bodyEditor := editor.NewView(buffer)
	bodyEditor.SetRuntimeFiles(runtime.Files)
	bodyEditor.SetColorscheme(colorscheme)
	bodyEditor.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlS:
			tools.SaveBuffer(buffer, fn)
			app.SetRoot(flex, true).SetFocus(requestForm)
			return nil
		}

		return event
	})

	settingsEditor := editor.NewView(bufferSettings)
	settingsEditor.SetRuntimeFiles(runtime.Files)
	settingsEditor.SetColorscheme(colorscheme)
	settingsEditor.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlS:
			tools.SaveBuffer(bufferSettings, tools.SettingsFile())
			app.SetRoot(flex, true).SetFocus(requestForm)
			return nil
		}

		return event
	})

	// response outputs
	responseView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetChangedFunc(func() {
			app.Draw()
		})

	statusView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetChangedFunc(func() {
			app.Draw()
		})

	// headers inputs
	headers := tview.NewTextView()

	// auth inputs
	token := tview.NewInputField().
		SetLabel("Token").
		SetFieldWidth(20)

	username := tview.NewInputField().
		SetLabel("Username").
		SetFieldWidth(20)

	password := tview.NewInputField().
		SetLabel("Password").
		SetFieldWidth(20)

	flex.AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(requestForm, 0, 1, false).
		AddItem(authForm, 20, 1, false).
		AddItem(headersForm, 15, 1, false), 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(responseView, 0, 3, false).
			AddItem(statusView, 7, 1, false), 0, 2, false).
		AddItem(tview.NewBox().SetBorder(true), 0, 0, false)

	Input := func(text, label string, width int, formToReturn *tview.Form, doneFunc func(text string)) {
		fileNameInput := tview.NewPages()

		input := tview.NewInputField().SetText(text)
		input.SetBorder(true)
		input.SetLabel(label).SetLabelWidth(width).SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				doneFunc(input.GetText())
				fileNameInput.RemovePage("input")
			} else if key == tcell.KeyEsc {
				app.SetRoot(flex, true).SetFocus(formToReturn)
			}
		})

		fileNameInput.AddAndSwitchToPage("input", tview.NewGrid().
			SetColumns(0, 0, 0).
			SetRows(0, 3, 0).
			AddItem(input, 1, 1, 1, 1, 0, 0, true), true).ShowPage("main")

		app.SetRoot(fileNameInput, true).SetFocus(input)
	}

	headersForm.AddButton("Add Header", func() {
		Input("", "header to add", 14, headersForm, func(text string) {
			headersForm.AddInputField(text, "", 20, nil, nil)

			headersCount++

			app.SetRoot(flex, true).SetFocus(headersForm)
		})
	}).AddButton("Remove Header", func() {
		Input("", "header to remove", 17, headersForm, func(text string) {
			headersForm.RemoveFormItem(headersForm.GetFormItemIndex(text))

			headersCount--

			app.SetRoot(flex, true).SetFocus(headersForm)
		})
	})

	send := func() {
		responseView.Clear()
		statusView.Clear()

		httpURL = urlField.GetText()

		b, e := os.Open(fn)

		if e != nil {
			fmt.Println(e)
		}

		defer b.Close()

		currentBody, err := ioutil.ReadAll(b)
		if err != nil {
			panic(err)
		}

		if cType == "application/json" {
			var r map[string]interface{}
			json.Unmarshal([]byte(currentBody), &r)
			body = string(pretty.Pretty([]byte(currentBody)))
		} else {
			body = string(currentBody)
		}

		if method == "POST" || method == "PUT" || method == "PATCH" || method == "DELETE" {
			respone, status, requestHeaders, _ = api.BasicRequestWithBody(
				httpURL,
				method,
				cType,
				body,
				authType,
				token.GetText(),
				username.GetText(),
				password.GetText(),
				false,
				headersCount,
				headersForm,
			)
		} else {
			body = ""

			respone, status, requestHeaders, _ = api.BasicGet(
				httpURL,
				method,
				authType,
				token.GetText(),
				username.GetText(),
				password.GetText(),
				false,
				headersCount,
				headersForm,
			)
		}

		headers.Clear()
		requestHeaders += "\n\nTo Exit Press 'Esc' Key"

		fmt.Fprintf(responseView, "%s ", respone)
		fmt.Fprintf(statusView, "%s ", status)
		fmt.Fprintf(headers, "%s", requestHeaders)
	}

	requestForm.AddFormItem(requestMethods).
		AddFormItem(urlField).
		AddFormItem(contentType).
		AddButton("Headers", func() {
			app.SetRoot(flex, true).SetFocus(headersForm)
		}).
		AddButton("Body", func() {
			app.SetRoot(bodyEditor, true).SetFocus(bodyEditor).Run()
		}).
		AddButton("Authorization", func() {
			app.SetRoot(flex, true).SetFocus(authForm)
		}).
		AddButton("Send", func() {
			send()

			app.SetRoot(flex, true).SetFocus(responseView)
		})

	responseView.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyTab {
			app.SetRoot(flex, true).SetFocus(requestForm)
		}
	})

	headers.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEsc {
			app.SetRoot(flex, true).SetFocus(requestForm)
		}
	})

	panelModal := tview.NewModal().
		SetText("What task do you want to do?").
		AddButtons([]string{
			"Request Form",
			"Send Request",
			"Body",
			"Headers",
			"Authorization",
			"Show Response Headers",
			"Save Response in File",
			"Return",
			"Quit From App",
		}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			switch buttonLabel {
			case "Request Form":
				app.SetRoot(flex, true).SetFocus(requestForm)

			case "Send Request":
				send()
				app.SetRoot(flex, true).SetFocus(requestForm)

			case "Body":
				app.SetRoot(bodyEditor, true).SetFocus(bodyEditor)

			case "Headers":
				app.SetRoot(flex, true).SetFocus(headersForm)

			case "Authorization":
				app.SetRoot(flex, true).SetFocus(authForm)

			case "Show Response Headers":
				app.SetRoot(headers, true).SetFocus(headers)

			case "Save Response in File":
				data := []byte(respone)

				Input("response.json", "file name", 5, requestForm, func(fn string) {
					err := os.WriteFile(fn, data, 0644)
					if err != nil {
						panic(err)
					}

					app.SetRoot(flex, true).SetFocus(requestForm)
				})

			case "Return":
				app.SetRoot(flex, true).SetFocus(requestForm)

			case "Quit From App":
				app.Stop()
			}
		})

	authForm.AddDropDown("Authentication Type", []string{"none", "basic auth", "bearer token"}, 0, func(option string, optionIndex int) {
		tokenIndex := authForm.GetFormItemIndex("Token")
		usernameIndex := authForm.GetFormItemIndex("Username")
		passwordIndex := authForm.GetFormItemIndex("Password")

		if option == "basic auth" {
			if tokenIndex != -1 {
				authForm.RemoveFormItem(authForm.GetFormItemIndex("Token"))
			} else if usernameIndex != -1 && passwordIndex != -1 {
				authForm.RemoveFormItem(authForm.GetFormItemIndex("Username"))
				authForm.RemoveFormItem(authForm.GetFormItemIndex("Password"))
			}

			authForm.AddFormItem(username)
			authForm.AddFormItem(password)

			authType = "basic"
		} else if option == "bearer token" {
			if usernameIndex != -1 && passwordIndex != -1 {
				authForm.RemoveFormItem(authForm.GetFormItemIndex("Username"))
				authForm.RemoveFormItem(authForm.GetFormItemIndex("Password"))
			} else if tokenIndex != -1 {
				authForm.RemoveFormItem(authForm.GetFormItemIndex("Token"))
			}

			authForm.AddFormItem(token)

			authType = "bearer"
		} else {
			if usernameIndex != -1 && passwordIndex != -1 {
				authForm.RemoveFormItem(authForm.GetFormItemIndex("Username"))
				authForm.RemoveFormItem(authForm.GetFormItemIndex("Password"))
			}

			if tokenIndex != -1 {
				authForm.RemoveFormItem(authForm.GetFormItemIndex("Token"))
			}

			token.SetText("")
			username.SetText("")
			password.SetText("")
		}
	})

	authForm.AddButton("Headers", func() {
		app.SetRoot(flex, true).SetFocus(headersForm)
	}).AddButton("Request", func() {
		app.SetRoot(flex, true).SetFocus(requestForm)
	})

	helpText.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEsc {
			app.SetRoot(flex, true).SetFocus(requestForm)
		}
	})

	// set borders
	authForm.SetBorder(true)
	headersForm.SetBorder(true)
	requestForm.SetBorder(true)
	responseView.SetBorder(true)
	statusView.SetBorder(true)

	// set titles
	authForm.SetTitle("Authentication").SetTitleAlign(tview.AlignCenter)
	headersForm.SetTitle("Headers").SetTitleAlign(tview.AlignCenter)
	requestForm.SetTitle("Request Form").SetTitleAlign(tview.AlignCenter)
	responseView.SetTitle("Response").SetTitleAlign(tview.AlignCenter)
	statusView.SetTitle("Status").SetTitleAlign(tview.AlignCenter)

	newReleaseModal := tview.NewModal()

	enableMouse := gjson.Get(tools.SettingsContent(), "rs_settings.enable_mouse").Bool()

	if enableMouse {
		app.EnableMouse(true)
	}

	if version != api.GetLatest() && gjson.Get(tools.SettingsContent(), "rs_settings.show_update").Bool() != false {
		newReleaseModal.SetText("There's a new version of resto is avalaible: " + version + " → " + api.GetLatest()).
			AddButtons([]string{"How to Update ?", "Don't show again", "Cancel"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				if buttonLabel == "How to Update ?" {
					app.SetRoot(updatePage, true).SetFocus(updatePage)
				} else if buttonLabel == "Don't show again" {
					tools.UpdateSettings(false)
					app.SetRoot(flex, true).SetFocus(requestForm)
				} else if buttonLabel == "Cancel" {
					app.SetRoot(flex, true).SetFocus(requestForm)
				}
			})

		if err := app.
			SetRoot(newReleaseModal, true).
			SetFocus(newReleaseModal).
			Sync().
			SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				switch event.Key() {
				case tcell.KeyCtrlP:
					app.SetRoot(panelModal, true).SetFocus(panelModal)
					return nil

				case tcell.KeyCtrlH:
					app.SetRoot(helpPage, true).SetFocus(helpPage)
					return nil

				case tcell.KeyCtrlU:
					app.SetRoot(newReleaseModal, true).SetFocus(newReleaseModal)

				case tcell.KeyCtrlE:
					app.SetRoot(settingsEditor, true).SetFocus(settingsEditor)
					return nil

				case tcell.KeyCtrlQ:
					app.Stop()
					return nil
				}

				return event
			}).
			Run(); err != nil {
			panic(err)
		}
	} else {
		newReleaseModal.SetText("All good, you're using the latest version of resto 👊").
			AddButtons([]string{"Ok"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				if buttonLabel == "Ok" {
					app.SetRoot(flex, true).SetFocus(requestForm)
				}
			})

		if err := app.
			SetRoot(flex, true).
			SetFocus(requestForm).
			Sync().
			SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				switch event.Key() {
				case tcell.KeyCtrlP:
					app.SetRoot(panelModal, true).SetFocus(panelModal)
					return nil

				case tcell.KeyCtrlW:
					app.SetRoot(helpPage, true).SetFocus(helpPage)
					return nil

				case tcell.KeyCtrlU:
					app.SetRoot(newReleaseModal, true).SetFocus(newReleaseModal)

				case tcell.KeyCtrlE:
					app.SetRoot(settingsEditor, true).SetFocus(settingsEditor)
					return nil

				case tcell.KeyCtrlQ:
					app.Stop()
					return nil

				case tcell.KeyCtrlJ:
					if requestForm.HasFocus(){
						if (requestForm.GetButton(requestForm.GetButtonCount() - 1).HasFocus()){
							app.SetFocus(authForm.GetFormItem(0))
						}else{
							fII, bII := requestForm.GetFocusedItemIndex()
							if bII == -1 && fII != requestForm.GetFormItemCount() - 1{
								app.SetFocus(requestForm.GetFormItem(fII + 1))
							}else if fII == -1 || fII == requestForm.GetFormItemCount() - 1{
								app.SetFocus(requestForm.GetButton(bII + 1))
							}
						}
					} else if authForm.HasFocus(){
						if (authForm.GetButton(authForm.GetButtonCount() - 1).HasFocus()){
							// have to check if there is a form item else transfer it to button
							if headersForm.GetFormItemCount() == 0{
								app.SetFocus(headersForm.GetButton(0))
							}else{
								app.SetFocus(headersForm.GetFormItem(0))
							}
						}else{
							fII, bII := authForm.GetFocusedItemIndex()
							if bII == -1 && fII != authForm.GetFormItemCount() - 1{
								app.SetFocus(authForm.GetFormItem(fII + 1))
							}else if fII == -1 || fII == authForm.GetFormItemCount() - 1{
								app.SetFocus(authForm.GetButton(bII + 1))
							}
						}
					} else if headersForm.HasFocus(){
						fII, bII := headersForm.GetFocusedItemIndex()
						if (headersForm.GetButton(headersForm.GetButtonCount() - 1).HasFocus()){
							app.SetFocus(requestForm.GetFormItem(0))
						}else if fII == -1 || bII != headersForm.GetFormItemCount() - 1{
							app.SetFocus(headersForm.GetButton(bII+1))
						}else if bII == -1 {
							app.SetFocus(headersForm.GetButton(0))
						}
					} else if statusView.HasFocus(){
						app.SetFocus(responseView)
					} else if responseView.HasFocus(){
						app.SetFocus(statusView)
					}
					return nil
				
				case tcell.KeyCtrlK:
					if headersForm.HasFocus(){
						fII, bII := headersForm.GetFocusedItemIndex()
						formItemCount := headersForm.GetFormItemCount()
						if formItemCount != 0 && fII == 0{
							app.SetFocus(authForm.GetButton(authForm.GetButtonCount() - 1))
						}else if headersForm.GetButton(0).HasFocus() &&  formItemCount != 0{
							app.SetFocus(headersForm.GetFormItem(formItemCount - 1))
						}else if headersForm.GetButton(0).HasFocus() && formItemCount == 0{
							app.SetFocus(authForm.GetButton(authForm.GetButtonCount() - 1))
						} else if fII == -1 {
							app.SetFocus(headersForm.GetButton(bII - 1))
						}
					} else if authForm.HasFocus(){
						fII, bII := authForm.GetFocusedItemIndex()
						if authForm.GetFormItem(0).HasFocus(){
							app.SetFocus(requestForm.GetButton(requestForm.GetButtonCount() - 1))
						}else if bII == 0 {
							app.SetFocus(authForm.GetFormItem(authForm.GetFormItemCount() - 1))
						}else if fII == -1 {
							app.SetFocus(authForm.GetButton(bII - 1))
						} else if bII == -1 && fII != 0 || fII != -1 {
							app.SetFocus(authForm.GetFormItem(fII - 1))
						}
					} else if requestForm.HasFocus(){
						fII, bII := requestForm.GetFocusedItemIndex()
						if requestForm.GetFormItem(0).HasFocus(){
							app.SetFocus(headersForm.GetButton(headersForm.GetButtonCount() - 1))
						}else if bII == 0 {
							app.SetFocus(requestForm.GetFormItem(requestForm.GetFormItemCount() - 1))
						}else if fII == -1 {
							app.SetFocus(requestForm.GetButton(bII - 1))
						} else if bII == -1 && fII != 0 || fII != -1 {
							app.SetFocus(requestForm.GetFormItem(fII - 1))
						}
					} else if statusView.HasFocus(){
						app.SetFocus(responseView)
					} else if responseView.HasFocus(){
						app.SetFocus(statusView)
					}

				case tcell.KeyCtrlL:
					if requestForm.HasFocus() || authForm.HasFocus(){
						app.SetFocus(responseView)
						return nil
					}
					if headersForm.HasFocus(){
						app.SetFocus(statusView)
						return nil
					}
					if responseView.HasFocus(){
						app.SetFocus(requestForm.GetFormItem(0))
						return nil
					}
					if statusView.HasFocus(){
						app.SetFocus(headersForm.GetButton(0))
						return nil
					}

				case tcell.KeyCtrlH:
					if requestForm.HasFocus() || authForm.HasFocus(){
						app.SetFocus(responseView)
						return nil
					}
					if headersForm.HasFocus(){
						app.SetFocus(statusView)
						return nil
					}
					if responseView.HasFocus(){
						app.SetFocus(requestForm.GetFormItem(0))
						return nil
					}
					if statusView.HasFocus(){
						app.SetFocus(authForm.GetFormItem(0))
						return nil
					}
				}

				return event
			}).
			Run(); err != nil {
			panic(err)
		}
	}
}

// AddItem(requestForm, 0, 1, false).
// AddItem(authForm, 20, 1, false).
// AddItem(headersForm, 15, 1, false), 0, 1, false).
// AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
// 	AddItem(responseView, 0, 3, false).
// 	AddItem(statusView, 7, 1, false), 0, 2, false).
// AddItem(tview.NewBox().SetBorder(true), 0, 0, false)
