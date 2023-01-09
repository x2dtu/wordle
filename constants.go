package main

// for formatting of console:
const WORD_LEN = 5
const NUM_TRIES = 6
const WORD_START = 8
const SPACE = "        "

// ansi color codes:
const GREEN = "\u001b[32m"
const YELLOW = "\u001b[33m"
const RED = "\u001b[31m"
const BLUE = "\u001b[34m"
const CYAN = "\u001b[36m"
const RESET = "\u001b[0m"
const GRAY = "\u001b[30;1m"

const KEYBOARD_START = "Q  W  E  R  T  Y  U  I  O  P" // used to get width of keyboard view
const ALPHABET_LEN = 26

var KEYBOARD = [ALPHABET_LEN]string{
	"Q", "W", "E", "R", "T", "Y", "U", "I", "O", "P",
	"\n\n A", "S", "D", "F", "G", "H", "J", "K", "L",
	"\n\n    Z", "X", "C", "V", "B", "N", "M"}

// colored keyboard is initialized in init() to be a copy of KEYBOARD that will have ansi color codes embedded within
var COLORED_KEYBOARD = make([]string, ALPHABET_LEN)

// keyboard positions has index for letter in keyboard array. For example:
// KEYBOARD_POSITIONS['a' - 'a'] = 10, and KEYBOARD[10] = 'A'
var KEYBOARD_POSITIONS = [ALPHABET_LEN]int{10, 23, 21, 12, 2, 13, 14, 15, 7, 16, 17, 18, 25, 24, 8, 9, 0, 3, 11, 4, 6, 22, 1, 20, 5, 19}

func init() {
	initKeyboard()
}

// turns each letter in keyboard blue, signifying that they havent been used in a word yet
func initKeyboard() {
	for i := 0; i < len(KEYBOARD); i++ {
		// COLORED_KEYBOARD[i] = BLUE + KEYBOARD[i] + RESET
		COLORED_KEYBOARD[i] = CYAN + KEYBOARD[i] + RESET
	}
}
