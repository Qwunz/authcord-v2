//Credits to https://github.com/atotto/clipboard/blob/master/clipboard_windows.go
//Credits to https://github.com/denisbrodbeck/machineid
//
package main
//
import (
	//"image/color"
	"log"
	"fmt"
	"os"
	"strings"
	"time"
	////
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
	//"fyne.io/fyne/v2/canvas"
	//"fyne.io/fyne/v2/container"
	//"fyne.io/fyne/v2/layout"
	////
	"github.com/bwmarrin/discordgo"
	"github.com/denisbrodbeck/machineid"
)
//
var (
	channel_id string
	user_key   string
	user_hwid  string
	check      bool = false
)
//
func response_check(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "!user.check=valid "+user_key {
		check = true
	} else if m.Content == "!user.register=hwid_added "+user_key+" "+user_hwid {
		log.Println("Your HWID has been set you can restart now!")
	}
}
//
func login_check(s *discordgo.Session, e *discordgo.Ready) {
	s.ChannelMessageSend(channel_id, "!user.login=check "+user_key+" "+user_hwid)
}
//
func register_check(s *discordgo.Session, e *discordgo.Ready) {
	s.ChannelMessageSend(channel_id, "!user.register=check "+user_key+" "+user_hwid)
}
//
func authcord_start(option, key, channel, token string) bool {
	hwid, e := machineid.ID()
	if e != nil {
		panic("Error Getting HWID")
	}

	user_hwid = hwid
	user_key = key
	channel_id = channel

	dg, e := discordgo.New("Bot " + token)
	if e != nil {
		panic("Error Initializing Bot")
	}
	defer dg.Close()

	dg.AddHandler(response_check)
	if option == "login" {
		dg.AddHandler(login_check)
	} else if option == "register" {
		dg.AddHandler(register_check)
	} else {
		return false
	}

	dg.Identify.Intents = discordgo.IntentsGuildMessages

	e = dg.Open()
	if e != nil {
		fmt.Printf("error: %v\n", e)
	}

	var i int
	for start := time.Now(); time.Since(start) < time.Second; {
		if i == 5000 && check == false {
			log.Println("Invalid Login...")
			time.Sleep(50000)
			os.Exit(0)
		}
		if check == true {
			//fmt.Println("Welcome User!")
			e = dg.Close()
			if e != nil {
				fmt.Printf("error: %v\n", e)
			}

			break
		}

		time.Sleep(10000)
		i++
	}

	//fmt.Println("\n\nWelcome User!")
	//time.Sleep(1000000000)

	return true
}
//
func parse_cmd(content string, token string, start, end int) string {
	splice := strings.Split(content, token)
	for q := 0; q < 9; q++ {
		splice = append(splice, " ")
	}

	return strings.Join(splice[start:end], "")
}
//
func main() {
	//var key string
	//var option string

	startApp := app.New()
	mainWindow := startApp.NewWindow("login-example")

	entry := widget.NewEntry()
	input := &widget.Form{
		Items: []*widget.FormItem{ // we can specify items in the constructor
			{Text: "User:", Widget: entry}},
		OnSubmit: func() { // optional, handle form submission
			log.Println("Checking Key:", entry.Text)
			mainWindow.Hide() // hide window
			authcord_start(
			// 	put your bots info here (cannot be same bot as server host bot)
				"login", entry.Text, 
				"channel-id", 
				"bot token")
			//	add register option before releasing to the hub
			//var key string = entry.Text
			for {
				if check == true {
					// valid login, you can start coding here or u can
					//	file stream or whatever... (add file-streaming)

					log.Println("                                                              ")
					log.Println("                                                              ")
					log.Println("                                                              ")
					log.Println("   Welcome back User!                                         ")
					log.Println("   ------------------                                         ")
					log.Println("    Demo V2 : Sexxxy                                          ")
					log.Println("                                                              ")
					log.Println("                                                              ")

					break
		
				} else if check == false {
					log.Println("Invalid User....")
					break
				}
				time.Sleep(20000)
			}

			mainWindow.Close() // kills app
		},
	}
	//
	// content := container.New(top_ui, bottom_ui)
	//
	//mainWindow.CenterOnScreen()
	mainWindow.Resize(fyne.NewSize(250, 100))
	mainWindow.SetContent(input)
	//
	// layout ^^^
	//
	mainWindow.ShowAndRun()
	//
	// w.SetContent(widget.NewLabel("Hello World!"))
	// w.ShowAndRun()
}

