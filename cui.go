package main

import (
	"fmt"
	"log"
	"strings"
	"unicode"

	"github.com/jroimartin/gocui"
	"github.com/x2dtu/wordle/wordle"
	"golang.org/x/exp/slices"
)

const WORD_LEN = 5
const NUM_TRIES = 6
const WORD_START = 8
const SPACE = "        "
const GREEN = "\u001b[32m"
const YELLOW = "\u001b[33m"
const RED = "\u001b[31m"
const BLUE = "\u001b[34m"
const CYAN = "\u001b[36m"
const RESET = "\u001b[0m"

var currWordle *wordle.Wordle

func main() {
	currWordle = wordle.New()

	g, err := gocui.NewGui(gocui.OutputNormal)
	g.Cursor = true
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.SetManagerFunc(layout)

	if err := keybindings(g); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	title := "Wordle"
	endTitleY := 2
	if v, err := g.SetView("title", maxX/2-len(title)/2, 0, maxX/2+len(title), endTitleY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		fmt.Fprintln(v, title)
	}
	description := "Guess the Hidden Word!"
	endDescriptionY := endTitleY + 2
	if v, err := g.SetView("description", maxX/2-len(description)/2, endTitleY, maxX+len(description)/2, endDescriptionY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		fmt.Fprintln(v, description)
	}
	if v, err := g.SetView("input", maxX/2-11, endDescriptionY+2, maxX/2+11, maxY/2+2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Editable = true
		if _, err := g.SetCurrentView("input"); err != nil {
			return err
		}
		writeBlankLines(v)
		v.SetCursor(WORD_START, 0)
	}
	return nil
}

func colorLetters(guess string, target string) string {
	target_map := make(map[rune]int)
	for _, char := range target {
		target_map[char] += 1
	}

	color_array := make([]string, WORD_LEN)

	guess_runes := []rune(guess)
	target_runes := []rune(target)

	// get green letters first
	for index, char := range guess_runes {
		if index >= len(target_runes) {
			break
		}

		if target_runes[index] == char {
			color_array[index] = GREEN + string(char)
			target_map[char]-- // we 'move' the char from target to color array
		}
	}

	// get remaining yellow and white letters now
	for index, char := range guess_runes {
		if index >= len(target_runes) {
			break
		}

		if target_runes[index] == char {
			continue // we finished the greens
		} else if target_map[char] > 0 {
			// then make this rune yellow
			color_array[index] = YELLOW + string(char)
			// decrement the value in map
			target_map[char]--
		} else {
			// then make this rune white
			color_array[index] = RESET + string(char)
		}
	}

	// join all the strings with RESET at end to make sure future text is white
	return SPACE + strings.Join(color_array, "") + RESET
}

func submitGuess(g *gocui.Gui, v *gocui.View) error {
	untrimmedGuess, err := v.Line(currWordle.Guesses)
	guess := strings.Trim(untrimmedGuess, " _")

	if err != nil {
		return err
	}
	// if this isn't a real word, don't submit the guess
	if len(guess) != WORD_LEN {
		return nil
	}
	if !slices.Contains(wordle.Words, guess) {
		turnRed(g, v)
		return nil
	}

	/* buffer lines will be like this: (in this ex for guess #0)
	 * [`guess`, _____, ...., _____]
	 * So, we want to save every line after `guess`.
	 * In general, for the current guess # k, we want to skip
	 * over line k in the buffer. So, we will use all saved lines
	 * before the guess and save all lines after the guess line.
	 * By saving those, we can clear the buffer. Then, we will
	 * print what was before, then the newly colored guess word,
	 * then we will reprint what was after
	 */
	bufferLinesBefore := currWordle.PreviousGuesses
	bufferLinesAfter := v.ViewBufferLines()[currWordle.Guesses+1:]

	v.Clear() // clear buffer

	// reprint the lines before the guess
	for i := range bufferLinesBefore {
		fmt.Fprintln(v, bufferLinesBefore[i])
	}
	// print newly made guess with colors
	colored_guess := colorLetters(guess, currWordle.Target)
	currWordle.PreviousGuesses = append(currWordle.PreviousGuesses, colored_guess)
	fmt.Fprintln(v, colored_guess)
	// reprint the lines after the guess
	for i := range bufferLinesAfter {
		fmt.Fprintln(v, bufferLinesAfter[i])
	}

	currWordle.Guesses++
	// put cursor at start of next line
	v.SetCursor(WORD_START, currWordle.Guesses)

	if guess == currWordle.Target {
		finishGame(v)
		fmt.Fprintf(v, "       %sYou won!%s\n", BLUE, RESET)
		outputDirections(v)

		// fmt.Fprintln(v, "Press Ctrl+C to quit.")
	} else if currWordle.Guesses == NUM_TRIES {
		// then we lost
		finishGame(v)
		fmt.Fprintf(v, "      %sYou lost!%s\n", RED, RESET)
		fmt.Fprintln(v, "The correct word was:")
		fmt.Fprintf(v, "%s%s\n", SPACE, currWordle.Target)
		outputDirections(v)
	}

	g.Update(layout)
	return nil
}

func turnRed(g *gocui.Gui, v *gocui.View) {
	currWord, err := v.Line(currWordle.Guesses)
	if err != nil {
		log.Panic("Couldn't read line")
	}
	bufferLinesAfter := v.ViewBufferLines()[currWordle.Guesses+1:]

	v.Clear()
	// reprint the lines before the guess
	for i := range currWordle.PreviousGuesses {
		fmt.Fprintln(v, currWordle.PreviousGuesses[i])
	}
	fmt.Fprintf(v, "%s%s%s\n", RED, currWord, RESET)
	// reprint the lines after the guess
	for i := range bufferLinesAfter {
		fmt.Fprintln(v, bufferLinesAfter[i])
	}
	currWordle.EnteredGibberish = true
}

func turnWhite(g *gocui.Gui, v *gocui.View) {
	// then clear the buffer, print everything before the word,
	// print the word in white, then print everything after word
	currWord, err := v.Line(currWordle.Guesses)
	if err != nil {
		log.Panic("Couldn't read line")
	}
	bufferLinesAfter := v.ViewBufferLines()[currWordle.Guesses+1:]

	v.Clear()
	// reprint the lines before the guess
	for i := range currWordle.PreviousGuesses {
		fmt.Fprintln(v, currWordle.PreviousGuesses[i])
	}
	fmt.Fprintln(v, currWord)
	// reprint the lines after the guess
	for i := range bufferLinesAfter {
		fmt.Fprintln(v, bufferLinesAfter[i])
	}
	currWordle.EnteredGibberish = false
}

func finishGame(v *gocui.View) {
	v.Clear()
	// reprint the lines before the guess
	for i := range currWordle.PreviousGuesses {
		fmt.Fprintln(v, currWordle.PreviousGuesses[i])
	}
	fmt.Fprintln(v)
	currWordle.GameOver = true
}

func outputDirections(v *gocui.View) {
	fmt.Fprintln(v)
	fmt.Fprintf(v, "Play Again: %sspace bar%s\n", CYAN, RESET)
	fmt.Fprintf(v, "Quit: %s^C%s\n", CYAN, RESET)
}

func handleBackspace(g *gocui.Gui, v *gocui.View) error {
	x, y := v.Cursor()
	if x > WORD_START {
		v.EditDelete(true) // delete character to left
		v.EditWrite('_')
		v.SetCursor(x-1, y)
	}
	if currWordle.EnteredGibberish {
		// then clear the buffer, print everything before the word,
		// print the word in white, then print everything after word
		turnWhite(g, v)
	}
	return nil
}

func handleCharacter(char rune) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		if currWordle.GameOver {
			return nil
		}
		x, y := v.Cursor()
		if x < WORD_START+WORD_LEN {
			// delete the underscore to the right
			v.EditDelete(false)
			v.EditWrite(unicode.ToLower(char))
			v.SetCursor(x+1, y)
		}
		return nil
	}
}

func handleShift(g *gocui.Gui, v *gocui.View) error {
	fmt.Fprintln(v, v.ViewBuffer())
	return nil
}

func doNothing(g *gocui.Gui, v *gocui.View) error {
	return nil
}

func handleSpace(g *gocui.Gui, v *gocui.View) error {
	if currWordle.GameOver {
		// if game over, then space bar will restart the game
		v.Clear()
		currWordle = wordle.New()
		writeBlankLines(v)
		v.SetCursor(WORD_START, 0)
		g.Update(layout)
	}
	return nil // else do nothing
}

func writeBlankLines(v *gocui.View) {
	for i := 0; i < 6; i++ {
		fmt.Fprintf(v, "%s_____\n", SPACE)
	}
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func keybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("input", gocui.KeyEnter, gocui.ModNone, submitGuess); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}
	if err := g.SetKeybinding("input", gocui.KeyBackspace, gocui.ModNone, handleBackspace); err != nil {
		return err
	}
	if err := g.SetKeybinding("input", gocui.KeyBackspace2, gocui.ModNone, handleBackspace); err != nil {
		return err
	}
	for _, c := range "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		if err := g.SetKeybinding("", c, gocui.ModNone, handleCharacter(c)); err != nil {
			return err
		}
	}
	if err := g.SetKeybinding("input", gocui.KeyDelete, gocui.ModNone, doNothing); err != nil {
		return err
	}
	if err := g.SetKeybinding("input", gocui.KeySpace, gocui.ModNone, handleSpace); err != nil {
		return err
	}
	for _, c := range "1234567890~!@#$%^&*()-_+=[]\\{}|;':\",./<>?" {
		if err := g.SetKeybinding("", c, gocui.ModNone, doNothing); err != nil {
			return err
		}
	}
	if err := g.SetKeybinding("input", gocui.KeyCtrlSpace, gocui.ModNone, handleShift); err != nil {
		return err
	}
	return nil
}