package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"unicode"

	"github.com/jroimartin/gocui"
	"github.com/x2dtu/wordle/wordle"
)

var currWordle *wordle.Wordle
var forcedLayout bool

func main() {
	flag.BoolVar(&forcedLayout, "f", false, "force game to play even with invalid terminal size")
	flag.Parse()

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
	// initialize y position variables for views
	startTitleY, endTitleY := 0, 2
	startDescriptionY, endDescriptionY := endTitleY, endTitleY+2
	startInputY, endInputY := endDescriptionY+2, endDescriptionY+17
	startKeyboardY := endInputY + 1
	endKeyboardY := startKeyboardY + 6

	maxX, maxY := g.Size()
	if maxY < endKeyboardY && !forcedLayout {
		fmt.Println("Your terminal height is too small to play Wordle.")
		fmt.Println("Try increasing the height and try again!")
		fmt.Println("Or, run with the -f flag to force a play session.")
		os.Exit(0)
	}

	title := "Wordle"

	if v, err := g.SetView("title", maxX/2-len(title)/2, startTitleY, maxX/2+len(title), endTitleY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		fmt.Fprintln(v, title)
	}
	description := "Guess the Hidden Word!"

	if v, err := g.SetView("description", maxX/2-len(description)/2, startDescriptionY, maxX+len(description)/2, endDescriptionY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		fmt.Fprintln(v, description)
	}

	if v, err := g.SetView("input", maxX/2-11, startInputY, maxX/2+11, endInputY); err != nil {
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

	if v, err := g.SetView("keyboard", maxX/2-len(KEYBOARD_START)/2-1, startKeyboardY, maxX/2+len(KEYBOARD_START)/2+1, endKeyboardY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		printKeyboard(v)
	}

	return nil
}

func printKeyboard(v *gocui.View) {
	fmt.Fprint(v, strings.Join(COLORED_KEYBOARD[:], "  "))
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

		// update keyboard to show this character as green
		updateCharInKeyboard(char, GREEN)
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

			// update keyboard to show this character as yellow
			updateCharInKeyboard(char, YELLOW)
		} else {
			// then make this rune white
			color_array[index] = RESET + string(char)

			// update keyboard to show this character as gray
			updateCharInKeyboard(char, GRAY)
		}
	}

	// join all the strings with RESET at end to make sure future text is white
	return SPACE + strings.Join(color_array, "") + RESET
}

// changes color of the specified character in the visual keyboard
func updateCharInKeyboard(char rune, color string) {
	original_char_str := KEYBOARD[KEYBOARD_POSITIONS[char-'a']]
	COLORED_KEYBOARD[KEYBOARD_POSITIONS[char-'a']] = color + original_char_str + RESET
}

func updateKeyboard(v *gocui.View) {
	v.Clear()
	printKeyboard(v)
}

func submitGuess(g *gocui.Gui, v *gocui.View) error {
	untrimmedGuess, err := v.Line(currWordle.Guesses)
	guess := strings.Trim(untrimmedGuess, " _")

	if err != nil {
		return err
	}
	// if this isn't a real word, don't submit the guess
	if len(guess) != WORD_LEN || currWordle.EnteredGibberish {
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
	// update keyboard
	keyboard_view, err := g.View("keyboard")
	if err != nil {
		log.Panic("No view named keyboard")
	}
	updateKeyboard(keyboard_view)

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
	fmt.Fprintf(v, "       Quit: %s^C%s\n", CYAN, RESET)
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

			untrimmedGuess, err := v.Line(currWordle.Guesses)
			if err != nil {
				log.Panic()
			}
			guess := strings.Trim(untrimmedGuess, " _")

			// if this isn't a real word, turn word red
			if len(guess) == WORD_LEN && !wordle.LegalWords[guess] {
				turnRed(g, v)
			}
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

		// update keyboard
		keyboard_view, err := g.View("keyboard")
		if err != nil {
			log.Panic("No view named keyboard")
		}
		initKeyboard()
		updateKeyboard(keyboard_view)

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
	if err := g.SetKeybinding("input", gocui.KeySpace, gocui.ModNone, handleSpace); err != nil {
		return err
	}
	for _, key := range []interface{}{gocui.KeyDelete, gocui.KeyArrowDown, gocui.KeyArrowUp, gocui.KeyArrowLeft, gocui.KeyArrowRight} {
		if err := g.SetKeybinding("input", key, gocui.ModNone, doNothing); err != nil {
			return err
		}
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
