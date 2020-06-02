package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/kylelemons/godebug/diff"
	"github.com/nsf/termbox-go"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

var termXSize int
var termYSize int
var lDisplay *widgets.Paragraph
var rDisplay *widgets.Paragraph
var masterEnterCode string

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
	termYSize = termYSize - 1

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

	termbox.SetCursor(0, termYSize)
	MenuBar()
	termbox.HideCursor()

	masterEnterCode = detectReturnCode(diffs[0].Strings)

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

func intLength(val int) int {
	strVal := strconv.Itoa(val)
	return len(strVal)
}

func makeSpace(spaces int) string {
	spaceStr := ""

	if spaces <= 0 {
		return ""
	}
	for i := 0; i < spaces; i++ {
		spaceStr += " "
	}
	return spaceStr
}

func splitDiffStr(diffStr []string) ([]string, []string) {
	diffLen := len(diffStr)
	var rStr, lStr []string

	rCnt := 1
	lCnt := 1

	for i := 0; i < len(diffStr); i++ {
		if len(diffStr[i]) > 0 {
			switch string(diffStr[i][0]) {
			case "+":
				rStr = append(rStr, makeSpace(intLength(diffLen)-intLength(rCnt))+fmt.Sprintf("[%d:](fg:red) [%s](fg:red)", rCnt, diffStr[i][1:]))
				rCnt = rCnt + 1
			case "-":
				lStr = append(lStr, makeSpace(intLength(diffLen)-intLength(lCnt))+fmt.Sprintf("[%d:](fg:blue) [%s](fg:blue)", lCnt, diffStr[i][1:]))
				lCnt = lCnt + 1
			default:
				lStr = append(lStr, makeSpace(intLength(diffLen)-intLength(lCnt))+fmt.Sprintf("%d: %s", lCnt, diffStr[i]))
				lCnt = lCnt + 1

				rStr = append(rStr, makeSpace(intLength(diffLen)-intLength(rCnt))+fmt.Sprintf("%d: %s", rCnt, diffStr[i]))
				rCnt = rCnt + 1
			}
		}
	}

	return lStr, rStr
}

func MenuBar() {
	fmt.Printf(" Next[Tab,x] Pre[BS,z] PDown[→,h,Space] PUp[←,l] Up[↑,k] Down[↓,j] Commit[Enter,c] Search[Home:/] Exit[Esc,q]")
}

func printToBuffer(diffTop int, lStr, rStr []string, lLen, rLen int) {
	lDisplay.Text = ""
	rDisplay.Text = ""
	termbox.Clear(defaultColor, defaultColor)

	for i := diffTop; i < diffTop+termYSize; i++ {
		if lLen > i {
			lDisplay.Text += fmt.Sprintf("%s\n", lStr[i])
		}
		if rLen > i {
			rDisplay.Text += fmt.Sprintf("%s\n", rStr[i])
		}
	}

	ui.Render(lDisplay, rDisplay)
	termbox.SetCursor(0, termYSize)
	MenuBar()
	termbox.HideCursor()
}

func detectReturnCode(strs string) string {
	r := regexp.MustCompile("\r\n")
	if r.MatchString(strs) == true {
		return "\r\n"
	}

	r = regexp.MustCompile("\n\r")
	if r.MatchString(strs) == true {
		return "\n\r"
	}

	r = regexp.MustCompile("\n")
	if r.MatchString(strs) == true {
		return "\n"
	}

	return "\r"
}

func emptyDiffStr() string {
	diffStrTmp := ""

	for _, v := range regexp.MustCompile("\r\n|\n\r|\n|\r").Split(diffs[0].Strings, -1) {
		diffStrTmp += " " + v + masterEnterCode
	}
	return diffStrTmp
}

func showDiff(cursol int) int {
	diffStrTmp := diff.Diff(diffs[0].Strings, diffs[cursol].Strings)

	lDisplay.Title = diffs[0].Filename
	rDisplay.Title = diffs[cursol].Filename

	if len(diffStrTmp) == 0 {
		diffStrTmp = emptyDiffStr()
		rDisplay.Title = "No Diffs: " + rDisplay.Title
	} else {
		rDisplay.Title = "Exists Diffs: " + rDisplay.Title
	}

	lStr, rStr := splitDiffStr(stringToArray(diffStrTmp))
	diffTop := 0
	lLen := len(lStr)
	rLen := len(rStr)

	termbox.SetInputMode(termbox.InputEsc)

	for {
		printToBuffer(diffTop, lStr, rStr, lLen, rLen)
		termbox.Flush()

		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case 9: //Tab
				return cursol + 1
			case 8: //Backspace
				return cursol - 1
			case 13: //Enter
				if commtDiff(diffs[cursol].Filename, cursol, diffTop) == true {
					diffs[cursol].Strings = readFile(diffs[cursol].Filename)
				}
				return cursol
			case 65521: //Home
				diffTop = searchStr(lStr, diffTop)
			case 65516: //Down
				diffTop = diffTop + 1
			case 65517:
				if diffTop > 0 {
					diffTop = diffTop - 1
				}
			case 65515, 65519: // Left
				if diffTop-termYSize < 0 {
					diffTop = 0
				} else {
					diffTop = diffTop - termYSize
				}
			case 32, 65514, 65518: // Right, and Space
				diffTop = diffTop + termYSize
			case 27: //ESC
				termbox.SetCursor(0, termYSize)
				fmt.Println("")
				termbox.Flush()
				ui.Close()
				os.Exit(0)
			default:
			}

			switch ev.Ch {
			case '/':
				diffTop = searchStr(lStr, diffTop)
			case 'x', 'X':
				return cursol + 1
			case 'z', 'Z':
				return cursol - 1
			case 'c', 'C':
				if commtDiff(diffs[cursol].Filename, cursol, diffTop) == true {
					diffs[cursol].Strings = readFile(diffs[cursol].Filename)
				}
				return cursol
			case 'j', 'J':
				diffTop = diffTop + 1
			case 'k', 'K':
				if diffTop > 0 {
					diffTop = diffTop - 1
				}
			case 'l', 'L':
				if diffTop-termYSize < 0 {
					diffTop = 0
				} else {
					diffTop = diffTop - termYSize
				}
			case 'h', 'H':
				diffTop = diffTop + termYSize
			case 'q', 'Q':
				termbox.SetCursor(0, termYSize)
				fmt.Println("")
				termbox.Flush()
				ui.Close()
				os.Exit(0)
			default:
			}
		}
	}
	return cursol
}

func inputStr() string {
	strs := ""
	cnt := 0
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case 8: //Backspace
				if cnt > 0 {
					strs = strs[0:(cnt - 1)]
					cnt = cnt - 1
				}
			case 32: //Space
				strs += " "
				break
			case 13:
				return strs
			default:
				strs += string(ev.Ch)
				cnt = cnt + 1
			}
		}
		termbox.SetInputMode(termbox.InputEsc)
		termbox.Flush()
		termbox.SetCursor(0, termYSize)
		commitStr := " search word: " + strs
		for i := len(commitStr); i < termXSize-1; i++ {
			commitStr += " "
		}
		fmt.Printf(commitStr)
		termbox.HideCursor()
	}
}

func searchStr(lStr []string, diffTop int) int {
	termbox.SetInputMode(termbox.InputEsc)
	termbox.Flush()
	termbox.SetCursor(0, termYSize)
	commitStr := " search word: "
	for i := len(commitStr); i < termXSize-1; i++ {
		commitStr += " "
	}
	fmt.Printf(commitStr)
	termbox.HideCursor()

	strs := inputStr()

	for i := 0; i < len(lStr); i++ {
		if strings.Index(lStr[i], strs) != -1 {
			return i
		}
	}
	return 0
}

func commtDiff(dstFilename string, cursol, diffTop int) bool {
	termbox.SetInputMode(termbox.InputEsc)
	termbox.Flush()
	termbox.SetCursor(0, termYSize)

	commitStr := fmt.Sprintf(" src: [%s] dst: [%s] Commit? (y/n/[a]ll)", diffs[0].Filename, dstFilename)
	for i := len(commitStr); i < termXSize-1; i++ {
		commitStr += " "
	}
	fmt.Printf(commitStr)
	termbox.HideCursor()

	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Ch {
			case 'a', 'A':
				writeFile(dstFilename, stringToArray(diffs[0].Strings))
				return true
			case 'y', 'Y':
				srcArray := stringToArray(diffs[0].Strings)
				dstArray := stringToArray(diffs[cursol].Strings)
				for i := diffTop; i < diffTop+termYSize && i < len(stringToArray(diffs[0].Strings)); i++ {
					dstArray[i] = srcArray[i]
				}
				writeFile(dstFilename, dstArray)
				return true
			case 'n', 'N', 'q':
				return false
			default:
			}
		}
	}
}

func arrayToString(strs []string) string {
	var result string
	result = ""
	for i := 0; i < len(strs); i++ {
		result += strs[i] + masterEnterCode
	}

	return result
}

func stringToArray(strs string) []string {
	var result []string

	for _, v := range regexp.MustCompile("\r\n|\n\r|\n|\r").Split(strs, -1) {
		result = append(result, v)
	}

	return result
}

func writeFile(dstFilename string, strs []string) bool {
	file, err := os.Create(dstFilename)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer file.Close()

	w := bufio.NewWriter(file)

	for i := 0; i < len(strs); i++ {
		row := strs[i]
		if i == (len(strs)-1) && len(strs[i]) == 0 {
			_, err = w.WriteString(row)
			if err != nil {
				return false
			}
		} else {
			_, err = w.WriteString(row + masterEnterCode)
			if err != nil {
				return false
			}
		}
	}

	err = w.Flush()
	if err != nil {
		return false
	}

	return true
}