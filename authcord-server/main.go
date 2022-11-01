package main

import (
	"bufio"
	"fmt"
	"strings"
	"time"

	//"log"
	//"syscall"
	//"strconv"
	"io"
	"os"

	//"io/ioutil"
	"math/rand"
	"net/http"

	//"path/filepath"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/net/html"

	"github.com/jroimartin/gocui"
)

var cnt int = 1
var cnt_max int = 12
var cnt_msgs int = 0
var recv_msgs = make([]string, 12)

var ( // TODO: Allow all of this shit to be set on jit on jit bc they wont have sauce nigga i on fo'nem
	cmd_channel_id string = "channel-id" 
	pay_channel_id string = "channel-id" 
	wallet_addr    string = "https://etherscan.io/address/your-eth-address"
	bot_token      string = "bot token"

	//users_key      string
	//users_sys_id   string
)

/* helpers */
func msg_handler(content string) {
	if cnt_msgs >= 12 {
		recv_msgs = nil
		cnt_msgs = 0
		return
	}

	cnt_msgs++
	recv_msgs[cnt_msgs] = content
}

func valid_check(hash string, hashes []string) bool { // valid_check now checks all hashes so they aren't overwritten by new purchases
	for _, v := range hashes {
		if hash == v {
			return true
		}
	}

	return false
}

func cred_check(key, hwid string) bool {
	file, e := os.Open("db.ini")
	if e != nil {
		fmt.Println("db.ini seems to be missing... Remake that file then retry!")
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		content := strings.Split(string(scanner.Text()), "\n")
		parsed_content := strings.Join(content, "\n")
		if strings.Contains(parsed_content, key+" "+hwid) {
			return true
		}
	}

	return false
}

func reg_check(key, hwid string) bool {
	file, e := os.Open("db.ini")
	if e != nil {
		fmt.Println("db.ini seems to be missing... Remake that file then retry!")
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var i int
	for start := time.Now(); time.Since(start) < time.Second; scanner.Scan() {
		if i == 10 {
			break
		}

		content := strings.Split(string(scanner.Text()), "\n")
		parsed_content := strings.Join(content, "\n")
		if strings.Contains(parsed_content, key+" "+hwid) {
			return true
		}

		time.Sleep(1000)
		i++
	}

	return false
}

func parse_cmd(content string, token string, start, end int) string {
	splice := strings.Split(content, token)
	for q := 0; q < 9; q++ {
		splice = append(splice, " ")
	}

	return strings.Join(splice[start:end], "")
}

func rand_str(num int) string {
	abc := "abcdefghijklmpqrstuvwxyz"
	bytes := make([]byte, num)

	for q := 0; q < num; q++ {
		bytes[q] = abc[rand.Intn(len(abc))]
	}
	return string(bytes)
}

func rand_num(num int) string {
	one_two_three := "123456789"
	bytes := make([]byte, num)

	for q := 0; q < num; q++ {
		bytes[q] = one_two_three[rand.Intn(len(one_two_three))]
	}
	return string(bytes)
}

func html_parse(body string) []string {
	tokenizer := html.NewTokenizer(strings.NewReader(body))
	var data []string
	var check int = 0
	for {
		tn := tokenizer.Next()
		switch {
		case tn == html.ErrorToken:
			return data
		case tn == html.StartTagToken:
			t := tokenizer.Token()
			if t.Data == "a" {
				check = 1
			}
		case tn == html.TextToken:
			t := tokenizer.Token()
			if check == 1 && strings.Contains(t.Data, "0x") {
				data = append(data, t.Data)
			}
			check = 0
		}
	}
}

func send_file(s *discordgo.Session, file string, channel_id string) {
	fss, _ := os.Open(file)
	msg_data := discordgo.MessageSend{
		Content: "",
		TTS:     false,
		File:    &discordgo.File{Name: rand_str(5) + ".go~", ContentType: "file", Reader: fss},
	}
	s.ChannelMessageSendComplex(channel_id, &msg_data)

	fss.Close()
}

/* helpers */

/* callbacks */
func msg_callback(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.ChannelID == cmd_channel_id {
		msg_handler(m.Content)

		user_key := parse_cmd(m.Content, " ", 1, 2)
		user_hwid := parse_cmd(m.Content, " ", 2, 2)

		//users_key = user_key
		//users_sys_id = user_hwid

		if strings.Contains(m.Content, "!user.login=check ") {

			if cred_check(user_key, user_hwid) {
				s.ChannelMessageSend(m.ChannelID, "!user.check=valid "+user_key+" "+user_hwid)

			}
		} else if strings.Contains(m.Content, "!user.register=check") {
			if reg_check(user_key, user_hwid) {
				s.ChannelMessageSend(m.ChannelID, "!user.check=valid "+user_key+" "+user_hwid)
			} else if reg_check(user_key, "NO_HWID") {
				// Open file using READ & WRITE permission.
				var file, e = os.OpenFile("db.ini", os.O_APPEND, 0755)
				if e != nil {
					fmt.Printf("failed to open db.ini\n")
				}
				defer file.Close()

				// Write some text line-by-line to file.
				_, e = file.WriteString("\n" + user_key + " " + user_hwid + "\n")
				if e != nil {
					fmt.Printf("failed to write key.\n")
				}
				// Save file changes.
				e = file.Sync()
				if e != nil {
					fmt.Printf("failed to save key.\n")
				}
				//fmt.Printf("key added Successfully.\n")
				s.ChannelMessageSend(m.ChannelID, "!user.register=hwid_set "+user_key+" "+user_hwid)
			}

		}
	} else if m.ChannelID == pay_channel_id {
		if strings.Contains(m.Content, "$verify ") {
			msg_handler(m.Content)
			hash := parse_cmd(m.Content, " ", 1, 2)
			oniichan, _ := s.UserChannelCreate(m.Author.ID) // create msg with user

			http_get, e := http.Get(wallet_addr)
			if e != nil {
				s.ChannelMessageSend(oniichan.ID, "[0x04] Error Finding Transaction "+"["+hash+"]")
				return
			}
			defer http_get.Body.Close()

			reader, e := io.ReadAll(http_get.Body)
			if e != nil {
				s.ChannelMessageSend(oniichan.ID, "[0x08] Error Finding Transaction "+"["+hash+"]")
				return
			}
			data := html_parse(string(reader))

			if valid_check(hash, data) { // TODO: don't just check latest because a purchase could go through essentially overwriting the last
				file, e := os.Open("db.ini")
				if e != nil {
					fmt.Println("db.ini seems to be missing... Remake that file then retry!")
					return
				}
				defer file.Close()

				scanner := bufio.NewScanner(file)
				for scanner.Scan() {
					content := strings.Split(string(scanner.Text()), "\n")
					parsed_content := strings.Join(content, "\n")

					if strings.Contains(parsed_content, hash) {
						s.ChannelMessageSend(oniichan.ID, "Hash has been used already!")
						return
					}
				}
				// Open file using READ & WRITE permission.
				file, e = os.OpenFile("db.ini", os.O_APPEND, 0755)
				if e != nil {
					fmt.Printf("failed to add key.\n")
				}
				defer file.Close()

				// make license code shit
				license_code := rand_str(5) + rand_num(5) + rand_str(6) + rand_num(4)
				fmt.Printf("fjfjfj: = %s", license_code)
				// Write some text line-by-line to file.
				_, e = file.WriteString("\n" + hash + " " + license_code + " NO_HWID" + "\n")
				if e != nil {
					fmt.Printf("\nfailed to add key.\n")
				}
				// Save file changes.
				e = file.Sync()
				if e != nil {
					fmt.Printf("\nfailed to add key.\n")
				}
				fmt.Printf("\nkey added Successfully.\n")
				s.ChannelMessageSend(oniichan.ID, "Thank you for purchasing!\nYour key: "+license_code)

				//now send software
				//send_file(s, "clientlib.zip", oniichan.ID)
				//s.ChannelMessageSend("1004788298553770030", "tb-addrole " + m.Author.ID + " 1011464055459946597 29d")
			}
		}
	}
}

/* callbacks */

/* cui */
func quit(gui *gocui.Gui, view *gocui.View) error {
	return gocui.ErrQuit
}

func update(gui *gocui.Gui, view *gocui.View) error {
	if cnt >= cnt_max {
		cnt_max += 12
		view.Clear()
	}
	view.FgColor = gocui.ColorCyan
	fmt.Fprintf(view, "[%d] %s: %s\n", cnt, "logs", recv_msgs)
	cnt++

	return nil
}

func layout(gui *gocui.Gui) error {
	x, y := gui.Size()
	main_view, e := gui.SetView("authcord", 0, 0, x-1, y-1)
	if e != gocui.ErrUnknownView {
		return e
	}
	main_view.Title = "AuthCord-Server"

	logs_view, e := gui.SetView("logs", 2, 2, x-3, y-3)
	if e != gocui.ErrUnknownView {
		return e
	}
	logs_view.Title = "Last Login Attempt"
	//draw_logs(logs_view)

	return nil
}

/* cui made by skidfaulted (thats why it sucksss) */

func main() {

	/*
		fmt.Println("Your wallet url: ")
		fmt.Scanln(&wallet_addr)
		fmt.Println("Your bot token: ")
		fmt.Scanln(&bot_token)
	*/

	//rand.Seed(time.Now().UnixNano())

	dg, e := discordgo.New("Bot " + bot_token)
	if e != nil {
		fmt.Printf("error New(): %v\n", e)
	}
	dg.AddHandler(msg_callback)
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	e = dg.Open()
	if e != nil {
		fmt.Printf("error: %v\n", e)
	}

	gui, e := gocui.NewGui(gocui.OutputNormal)
	if e != nil {
		panic("Failed to init GUI")
	}
	defer gui.Close()

	gui.FgColor = gocui.ColorRed
	gui.BgColor = gocui.ColorBlue
	gui.Mouse = true

	gui.SetManagerFunc(layout)

	if e := gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); e != nil {
		panic("Failed to quit")
	}

	if e := gui.SetKeybinding("logs", gocui.MouseLeft, gocui.ModNone, update); e != nil {
		panic("Failed to setup keybind")
	}

	e = gui.MainLoop()
	if e != nil {
		panic("Failed to start main loop")
	}

	dg.Close()
}
