package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	//"flag"
	"io/ioutil"
	"log"
	_ "log"
	_ "runtime"
	"strconv"
	"strings"
	"time"

	"github.com/tarm/serial"
	_ "github.com/timtadh/data-structures/tree/avl"
)

//commands records the available commands.
//var commands = map[string]func(){
//"download": 	*comm.downloadBin,
// "banner":    cmdbanner,
// "bootstrap": cmdbootstrap,
// "clean":     cmdclean,
// "env":       cmdenv,
// "install":   cmdinstall,
// "list":      cmdlist,
// "test":      cmdtest,
// "version":   cmdversion,
//}

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	cmd := os.Args[1]
	os.Args = os.Args[1:] // for flag parsing during cmd
	// if f, ok := commands[cmd]; ok {
	// 	f()
	// } else {
	// 	xprintf("unknown command %s\n", cmd)
	// 	usage()
	// }

	// s := ""
	// for _, arg := range os.Args {
	// 	s += arg
	// 	//sep = " "
	// }
	// fmt.Println(s)

	// sep := strings.Split(s, "-")
	// for _, str := range sep {
	// 	fmt.Println(str)
	// }

	switch cmd {

	case "download":
		var com comm
		com.downloadBin()
	case "help", "-h":
		usage()
	case "format":
		convertToAsm()
	default:
		fmt.Println("unkown command")
	}

	//readFile()
	//convertToAsm()
	//

}

type comm struct {
	Port, path string
	Baudrate   int
	s          *serial.Port
	done       chan int
}

func (this *comm) init() {
	this.done = make(chan int)

	for i, arg := range os.Args {
		if arg == "-r" {
			var err error
			this.Baudrate, err = strconv.Atoi(os.Args[i+1])
			if err != nil {
				fmt.Printf("input baud rate is error %s", os.Args[i+1])
				os.Exit(-1)
			}
		}
		if arg == "-p" {
			this.Port = os.Args[i+1]
		}
		if arg == "-f" {
			this.path = os.Args[i+1]
		}
	}
	if this.Port == "" {
		fmt.Println("set port :")
		fmt.Scanln(&this.Port)
	}
	if this.Baudrate == 0 {
		fmt.Println("set baut rate :")
		fmt.Scanln(&this.Baudrate)
	}
	if this.path == "" {
		fmt.Println("set file path :")
		filselect := make(map[int]string)

		filename, _ := ioutil.ReadDir(".")
		i := 0
		for _, na := range filename {
			if strings.HasSuffix(na.Name(), ".bin") {
				filselect[i] = na.Name()
				fmt.Println(i, " ", na.Name())
				i++
			}
		}
		n := 0
		fmt.Println("select file in current dir :")
		fmt.Scanln(&n)
		fmt.Println("you select: ", filselect[n])
		this.path = filselect[n]
		fmt.Println("now start send file ...")
		time.Sleep(200 * time.Millisecond)
		//fmt.Scanln(&this.path)
	}

	//filename := os.Readdirnames(10)

	//this.path = "HT45F4842CHECK.bin"
	fmt.Printf("Currently all ports are opened with 8 data bits, 1 stop bit, no parity, no hardware flow control, and no software flow control.  serial port is %s,baud rate is %d\n", this.Port, this.Baudrate)
}

func (this *comm) downloadBin() {
	this.init()
	//c := &serial.Config{Name: "COM17", Baud: 115200}
	c := new(serial.Config)
	c.Name = this.Port
	c.Baud = this.Baudrate
	c.ReadTimeout = time.Millisecond * 500
	var err error
	//s, err := serial.OpenPort(c)
	this.s, err = serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}
	defer this.s.Close()
	//HT45F4842CHECK.bin
	f, _ := os.Open(this.path)
	defer f.Close()
	buff := bufio.NewReader(f)

	go func() {

		data := []byte{0xfc, 0xcf, 0x00, 0x04, 0x0d, 0x09, 0xc3, 0x40, 0x1d, 0xed}
		_, err := this.s.Write(data[:10])
		if err != nil {
			log.Fatal(err)
		}

		cnt := 0
		for {

			time.Sleep(100 * time.Millisecond)
			//data, err := buff.Peek(128)
			data := make([]byte, 128)
			n, err := buff.Read(data)
			// if err!=nil || err==io.EOF{
			// 	this.done <- 1
			// 	break
			// }
			// if cnt > 2000 {
			// 	this.done <- 1
			// 	break
			// }

			if err == io.EOF || err != nil {
				//log.Fatal(err)
				fmt.Printf("counted %d , send complete", cnt)
				this.done <- 1
				break
			}
			// fmt.Print("\n")
			// log.Printf("total tx size %d %d", len(data), n)
			// for _, i := range data {
			// 	fmt.Printf("%x", i)
			// }
			i, err := this.s.Write(data[:n])
			cnt += i
			// for i,n:=0,0;n <= len(data);n +=i {

			// }

		}
	}()

	go func() {
		for {
			// select {
			// case <-this.done:
			// 		fmt.Println("exiting...")
			// 		this.done <- 1
			// 		break
			// default:
			time.Sleep(100 * time.Millisecond)
			buf := make([]byte, 128)
			n, _ := this.s.Read(buf)
			log.Printf("total rx size %d %d", len(buf), n)
			for _, i := range buf[:n] {
				fmt.Printf("%x", i)
			}
			fmt.Print("\n")
			//}
		}
	}()

	<-this.done
	//close(this.done)
}

func convertToAsm() {
	path := ""
	for i, arg := range os.Args {
		if arg == "-f" {
			path = os.Args[i+1]
		}
	}
	if path == "" {
		fmt.Println("set file path :")
		filselect := make(map[int]string)

		filename, _ := ioutil.ReadDir(".")
		i := 0
		for _, na := range filename {
			fmt.Println(i, " ", na.Name())
			filselect[i] = na.Name()
			i++
		}
		n := 0
		fmt.Println("select file in current dir :")
		fmt.Scanln(&n)
		fmt.Println("you select: ", filselect[n])
		path = filselect[n]
		//fmt.Scanln(&this.path)
	}
	newName := "HT45F4842"
	fmt.Println("new file name:")
	fmt.Scanln(&newName)
	newName += ".asm"

	fmt.Println("now format file ...")
	time.Sleep(1000 * time.Millisecond)

	f, _ := os.Open(path)
	defer f.Close()

	fileasm, _ := os.Create(newName)
	defer fileasm.Close()

	buff := bufio.NewReader(f)
	for {
		line, err := buff.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if ok := strings.HasPrefix(line, "0"); ok {
			split := strings.Fields(line)
			line = ";"
			line += split[0]
			line += ","
			line += split[1]
			line += "\n"
			line += "    "
			for _, str := range split[2:] {
				line += str
			}
		}
		fmt.Println(line)
		line += "\n"
		fileasm.WriteString(line)
	}
	fmt.Println("format complete!")
}

func readFile() {
	// files, err := ioutil.ReadDir(".")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// for _, file := range files {
	// 	fmt.Println(file.Name())
	// }

	// content, err := ioutil.ReadFile("main.go")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Printf("File contents: %s", content)
	f, err := os.Open("HT45F4842CHECK.list")
	defer f.Close()
	printError(err)

	i := 0
	buff := bufio.NewReader(f)
	for {
		line, err := buff.ReadString('\n')
		if err != nil {
			break
		}
		i++

		//line = strings.TrimLeft(strings.TrimSpace(line),";")
		line = strings.TrimSpace(line)
		if ok := strings.HasPrefix(line, ";"); ok {
			continue
		}

		split := strings.Fields(line)
		if len(split) > 1 {

			if ok := strings.ContainsAny(split[0], "_ghijklmnopqrstuvwxyz;GHIJKLMNOPQRSTUVWXYZ"); ok {
				continue
			}

			// if ok := strings.ContainsAny(split[1],"_&.");ok{
			// 	continue
			// }

			// if i := len(split[1]);i<3{
			// 	continue
			// }
			//fmt.Println(split[1])
			fmt.Printf("%s : %s \n", split[0], split[1])
		}
	}
}

func printError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func xprintf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

var atexits []func()

// xexit exits the process with return code n.
func xexit(n int) {
	for i := len(atexits) - 1; i >= 0; i-- {
		atexits[i]()
	}
	os.Exit(n)
}

func usage() {
	xprintf(`usage: go tool dist [command]
	Commands are:

	download       IAP function ,download file to flash
	[-p]           name of serial port
	[-r]		   baut rate
	[-f]		   file name
	clean          deletes all built files
	list           list all file current directory
	format         format list file to assemble
	help[-h]       more help infomation
	version        print  version
	All commands take -v flags to emit extra information.
	example :
	go download -p com17  -r 115200
	`)
	xexit(2)
}
