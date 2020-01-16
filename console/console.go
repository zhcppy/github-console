package console

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
)

var (
	onlyWhiteSpace = regexp.MustCompile(`^\s*$`)
	exit           = regexp.MustCompile(`^\s*exit\s*;*\s*$`)
)

type Handler interface {
	ExecCommand(ctx context.Context, cmd string) error
}

// HistoryFile is the file within the data directory to store input scrollback.
const HistoryFile = ".gh_history"

// DefaultPrompt is the default prompt line prefix to use for user input querying.
const DefaultPrompt = "$🐌 "

type Console struct {
	prompt   string       // Input prompt prefix string
	prompter UserPrompter // Input prompter to allow interactive user feedback
	histPath string       // Absolute path to the console scrollback history
	history  []string     // Scroll history maintained by the console
	Printer  io.Writer    // Output writer to serialize any display strings to
	handler  Handler
	ctx      context.Context
}

func New(ctx context.Context, handler Handler) *Console {
	console := &Console{
		prompt:   DefaultPrompt,
		prompter: Stdin,
		histPath: filepath.Join(os.Getenv("HOME"), HistoryFile),
		history:  make([]string, 0),
		Printer:  os.Stdout,
		handler:  handler,
		ctx:      ctx,
	}

	if content, err := ioutil.ReadFile(console.histPath); err != nil {
		console.prompter.SetHistory(nil)
	} else {
		console.history = strings.Split(string(content), "\n")
		console.prompter.SetHistory(console.history)
	}
	return console
}

func (c *Console) SetWordCompleter(words []string) {
	c.prompter.SetWordCompleter(func(line string, pos int) (head string, completions []string, tail string) {
		if len(line) == 0 || pos == 0 {
			for _, word := range words {
				if strings.Index(word, ".") == -1 {
					completions = append(completions, word)
				}
			}
			return head, completions, tail
		}
		for _, word := range words {
			if strings.HasPrefix(strings.ToLower(word), strings.ToLower(string([]rune(line)[:pos]))) {
				completions = append(completions, word)
			}
		}
		return head, completions, string([]rune(line)[pos:])
	})
}

func (c *Console) Welcome(message string) {
	c.println(message)
}

// Interactive starts an interactive user session, where input is propted from
// the configured user prompter.
func (c *Console) Interactive() {
	var scheduler = make(chan string) // Channel to send the next prompt on and receive the input

	// Start a goroutine to listen for prompt requests and send back inputs
	go func() {
		for {
			select {
			case <-c.ctx.Done():
				return
			default:
				prompt, ok := <-scheduler
				if !ok {
					return
				}
				// Read the next user input
				line, err := c.prompter.PromptInput(prompt)
				if err != nil {
					// In case of an error, either clear the prompt or fail
					if err == abortedErr { // ctrl-C
						scheduler <- ""
						//c.println("please input 'exit/ctrl-D' shut down console.")
						continue
					}
					c.println("\nexiting.")
					close(scheduler)
					return
				}
				// User input retrieved, send for interpretation and loop
				scheduler <- line
			}
		}
	}()

	// Monitor Ctrl-C too in case the input is empty and we need to bail
	abort := make(chan os.Signal, 1)
	signal.Notify(abort, syscall.SIGINT, syscall.SIGTERM)

	// Start sending prompts to the user and reading back inputs
	for {
		// Send the next prompt, triggering an input read and process the result
		scheduler <- c.prompt
		select {
		case <-c.ctx.Done():
			return
		case sig := <-abort:
			// User forcefully quit the console (kill)
			c.println("caught interrupt: ", sig.String(), "exiting.")
			return
		default:
			input, ok := <-scheduler
			// User input was returned by the prompter, handle special cases
			if !ok || exit.MatchString(input) {
				return
			}
			if onlyWhiteSpace.MatchString(input) {
				continue
			}
			if err := c.Execute(input); err != nil {
				c.println(err.Error())
			}
		}
	}
}

// Evaluate executes func and pretty prints the result to the specified output stream.
func (c *Console) Execute(command string) error {
	appendHistory := func() {
		if len(c.history) == 0 || command != c.history[len(c.history)-1] {
			c.history = append(c.history, command)
			c.prompter.AppendHistory(command)
		}
	}
	defer func() {
		if r := recover(); r != nil {
			c.println("[native] error: ", r)
		} else {
			appendHistory()
		}
	}()
	return c.handler.ExecCommand(c.ctx, command)
}

func (c *Console) Exit() error {
	if err := ioutil.WriteFile(c.histPath, []byte(strings.Join(c.history, "\n")), 0600); err != nil {
		return err
	}
	return os.Chmod(c.histPath, 0600)
}

func (c *Console) ClearHistory() {
	c.history = make([]string, 0)
	c.prompter.ClearHistory()
	if err := os.Remove(c.histPath); err != nil {
		c.println("can't delete history file:", err.Error())
	} else {
		c.println("history file deleted.")
	}
}

func (c *Console) println(msg ...interface{}) {
	_, _ = fmt.Fprintln(c.Printer, msg...)
}
