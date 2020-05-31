package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"

	"github.com/kylelemons/godebug/diff"
	"github.com/nsf/termbox-go"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

var termXSize int
var termYSize int
var lDisplay *widgets.Paragraph
var rDisplay *widgets.Paragraph

type stringData struct {
	Strings  string `json:"strings"`
	Filename string `json:"filename"`
}

var diffs = []stringData{}

const (
	defaultColor = termbox.ColorDefault
)

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Println("mdiff: Multi Diff Tool.")
		fmt.Println("usage) mdiff [Master] [Diffs...]")
		os.Exit(1)
	}

	for i := 0; i < flag.NArg(); i++ {
		diffs = append(diffs, stringData{Strings: readFile(flag.Arg(i)), Filename: flag.Arg(i)})
	}

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	termXSize, termYSize = termbox.Size()

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	lDisplay = widgets.NewParagraph()
	lDisplay.TextStyle.Fg = ui.ColorWhite
	lDisplay.SetRect(0, 0, int(termXSize/2), termYSize)
	lDisplay.BorderStyle.Fg = ui.ColorCyan
	lDisplay.BorderStyle.Bg = ui.ColorBlack

	rDisplay = widgets.NewParagraph()
	rDisplay.TextStyle.Fg = ui.ColorWhite
	rDisplay.SetRect(int(termXSize/2)-1, 0, termXSize, termYSize)
	rDisplay.BorderStyle.Fg = ui.ColorCyan
	rDisplay.BorderStyle.Bg = ui.ColorBlack

	ui.Render(lDisplay, rDisplay)

	cursor := 1

	for {
		cursor = showDiff(cursor)
		if cursor == 0 {
			cursor = flag.NArg() - 1
		}
		if cursor == flag.NArg() {
			cursor = 1
		}
	}
}

func readFile(filename string) string {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}

	return string(bytes)
}

func printToBuffer(diffStr []string, diffTop int) {
	lDisplay.Text = ""
	rDisplay.Text = ""
	termbox.Clear(defaultColor, defaultColor)

	cnt := diffTop + 1
	for i := diffTop; i < diffTop+termYSize; i++ {
		if len(diffStr[i]) > 0 {
			switch string(diffStr[i][0]) {
			case "+":
				lDisplay.Text += fmt.Sprintf("[%d: %s](fg:blue)\n", cnt, diffStr[i])
				rDisplay.Text += fmt.Sprintf("%d: \n", cnt)
			case "-":
				lDisplay.Text += fmt.Sprintf("%d: \n", cnt)
				rDisplay.Text += fmt.Sprintf("[%d: %s](fg:red)\n", cnt, diffStr[i])
			default:
				lDisplay.Text += fmt.Sprintf("%d: %s\n", cnt, diffStr[i])
				rDisplay.Text += fmt.Sprintf("%d: %s\n", cnt, diffStr[i])
			}
		}
		cnt = cnt + 1
	}

	ui.Render(lDisplay, rDisplay)
}

func lineDown(diffTop, diffLen int) int {
	if diffTop+termYSize+termYSize >= diffLen {
		return diffLen - termYSize
	}
	return diffTop + termYSize
}

func lineUp(diffTop int) int {
	if diffTop-termYSize <= 0 {
		return 0
	}
	return diffTop - termYSize
}

func oneDown(diffTop, diffLen int) int {
	if diffTop+termYSize+1 >= diffLen {
		return diffTop
	}
	return diffTop + 1
}

func oneUp(diffTop int) int {
	if diffTop > 0 {
		diffTop = diffTop - 1
	}
	return diffTop
}

func showDiff(cursol int) int {
	diffStrTmp := diff.Diff(diffs[0].Strings, diffs[cursol].Strings)

	lDisplay.Title = diffs[0].Filename
	rDisplay.Title = diffs[cursol].Filename

	if len(diffStrTmp) == 0 {
		diffStrTmp = ""
		for _, v := range regexp.MustCompile("\r\n|\n\r|\n|\r").Split(diffs[0].Strings, -1) {
			diffStrTmp += " " + v + "\n"
		}
		rDisplay.Title = "No Diffs: " + rDisplay.Title
	} else {
		rDisplay.Title = "Exists Diffs: " + rDisplay.Title
	}
	diffStr := stringToArray(diffStrTmp)
	diffTop := 0
	diffLen := len(diffStr)

	termbox.SetInputMode(termbox.InputEsc)

	for {
		printToBuffer(diffStr, diffTop)
		termbox.Flush()

		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case 9: //Tab
				return cursol + 1
			case 8: //Backspace
				return cursol - 1
			case 13: //Enter
				if commtDiff(diffs[cursol].Filename) == true {
					diffs[cursol].Strings = diffs[0].Strings
				}
				return cursol
			case 65516: //Down
				diffTop = oneDown(diffTop, diffLen)
			case 65517:
				diffTop = oneUp(diffTop) // Upper
			case 65515, 65519: // Left
				diffTop = lineUp(diffTop)
			case 32, 65514, 65518: // Right, and Space
				diffTop = lineDown(diffTop, diffLen)
			case 27: //ESC
				termbox.Flush()
				ui.Close()
				os.Exit(0)
			default:
			}

			switch ev.Ch {
			case 'x', 'X':
				return cursol + 1
			case 'z', 'Z':
				return cursol - 1
			case 'c', 'C':
				if commtDiff(diffs[cursol].Filename) == true {
					diffs[cursol].Strings = diffs[0].Strings
				}
				return cursol
			case 'j', 'J':
				diffTop = oneDown(diffTop, diffLen)
			case 'k', 'K':
				diffTop = oneUp(diffTop)
			case 'l', 'L':
				diffTop = lineUp(diffTop)
			case 'h', 'H':
				diffTop = lineDown(diffTop, diffLen)
			case 'q', 'Q':
				termbox.Flush()
				ui.Close()
				os.Exit(0)
			default:
			}
		}
	}
	return cursol
}

func commtDiff(dstFilename string) bool {
	termbox.SetInputMode(termbox.InputEsc)
	termbox.Flush()
	lDisplay.Text = ""
	lDisplay.Text += fmt.Sprintf("src: [%s]\n", diffs[0].Filename)
	lDisplay.Text += fmt.Sprintf("dst: [%s]\n", dstFilename)
	lDisplay.Text += fmt.Sprintf("Commit? (y/n)\n")
	ui.Render(lDisplay, rDisplay)

	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case 13:
				writeFile(dstFilename, diffs[0].Strings)
				return true
			case 27: //ESC
				return false
			default:
			}

			switch ev.Ch {
			case 'y', 'Y':
				writeFile(dstFilename, diffs[0].Strings)
				return true
			case 'n', 'N', 'q':
				return false
			default:
			}
		}
	}
}

func stringToArray(strs string) []string {
	var result []string

	for _, v := range regexp.MustCompile("\r\n|\n\r|\n|\r").Split(strs, -1) {
		result = append(result, v)
	}

	return result
}

func writeFile(dstFilename, strs string) bool {
	file, err := os.Create(dstFilename)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer file.Close()

	_, err = file.WriteString(strs)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}
