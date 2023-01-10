# Wordle
This is a Wordle game I made using Go. Instead of a normal GUI design that I have usually done for my other projects, I decided to change things up and create a CUI (Character User Interface). So, the game is interacted with using a command line terminal, which makes use of the [gocui](https://pkg.go.dev/github.com/jroimartin/gocui@v0.5.0) package in Go. It has all the features of the New York Times' Wordle and more. I always wanted to remake Wordle as I usually play it every day; I believe it's a fun way to start the morning. I've wanted to also learn Go for some time now, so this gave me a great opportunity to code in a language like Go on a project I've wanted to work on for some time - Win Win! As always, please email me at 3069391@gmail.com or comment on this project page should you have any questions about the game, suggestions for further improvement, or have found any bugs.

## Screenshots
Starting the application with ``go run .`` will change the terminal to look like this: <br/>
<img src="https://user-images.githubusercontent.com/82241006/211656749-666580b0-537d-4c62-9d81-182edcbed5ca.png" alt="wordle" height="500"/> <br/><br/>
On this gif below, notice how gibberish words are highlighted in red as soon as they are written out. Pressing `enter` when a gibberish word or word less than 5 characters long will do nothing. As per the wordle rules, letters in the correct spots will show as green when submitted, yellow if they are in the word but not in the correct spot, and gray if they are not in the word. The keyboard at the bottom of the screen also mirrors this.  <br/> 
<img src="https://user-images.githubusercontent.com/82241006/211659836-23196568-78ac-446f-b3e4-d19dbd9384e0.gif" alt="wordle" height="500" /> <br/>
Inputting capital letters will turn into lowercase. Numbers, special characters, spaces, underscores, arrow keys, and delete key won't do anything. Adding more characters after 5 won't do anything until you submit the word. Backspacing characters is allowed, but backspacing a blank line won't do anything. Guessing the correct word will beat the game, but incorrectly guessing the word 6 times will lose you the game. Once the game is finished, you can press the space bar to restart the game with a new word. You can press ``^C`` at any time to end the game. <br/><br/>
Winning the game: <br/>
<img src="https://user-images.githubusercontent.com/82241006/211662270-8af781d8-e913-4554-a766-e91cc03dc3ab.gif" alt="wordle" height="500" /> <br/>
Losing the game: <br/>
<img src="https://user-images.githubusercontent.com/82241006/211662709-5093b9bc-4269-43a8-bc4f-2285a3afda23.gif" alt="wordle" height="500" /> <br/>
If your terminal screen is too short to accurately display the game, the program will print a notice message and exit: <br/>
![image](https://user-images.githubusercontent.com/82241006/211663159-3e9da29b-84b3-45bb-843b-9a8d2354160e.png)
<br/>
As told in the message above, you can run the application with ``go run . -f`` to force the game to run at any resolution, even if the terminal height is too small. 
