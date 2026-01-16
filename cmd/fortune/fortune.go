package fortune

import (
	"fmt"
	"math/rand"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type Params struct {
	Category string `short:"c" help:"Category: tech, wisdom, debugging, all." default:"all"`
}

var techFortunes = []string{
	"A SQL query walks into a bar, walks up to two tables and asks... 'Can I join you?'",
	"There are only two hard things in computer science: cache invalidation, naming things, and off-by-one errors.",
	"It's not a bug, it's an undocumented feature.",
	"The best code is no code at all.",
	"Weeks of coding can save you hours of planning.",
	"In theory, theory and practice are the same. In practice, they're not.",
	"Software is like entropy: it's hard to reverse, and it only increases.",
	"The code you write makes you a programmer. The code you delete makes you a good one.",
	"Programming is 10% writing code and 90% understanding why it doesn't work.",
	"A good programmer is someone who always looks both ways before crossing a one-way street.",
	"The cloud is just someone else's computer.",
	"There's no place like 127.0.0.1",
	"SELECT finger FROM hand WHERE finger = 'pinky'; -- Returns NULL",
	"The best thing about a boolean is that even if you're wrong, you're only off by a bit.",
	"Why do programmers prefer dark mode? Because light attracts bugs.",
	"Debugging is like being the detective in a crime movie where you're also the murderer.",
	"Code never lies, comments sometimes do.",
	"First, solve the problem. Then, write the code.",
	"Any fool can write code that a computer can understand. Good programmers write code that humans can understand.",
	"Deleted code is debugged code.",
}

var wisdomFortunes = []string{
	"The best time to plant a tree was 20 years ago. The second best time is now.",
	"Simplicity is the ultimate sophistication.",
	"Done is better than perfect.",
	"Make it work, make it right, make it fast.",
	"Premature optimization is the root of all evil.",
	"If you can't explain it simply, you don't understand it well enough.",
	"The only way to go fast is to go well.",
	"Perfection is achieved not when there is nothing more to add, but when there is nothing left to take away.",
	"The best error message is the one that never shows up.",
	"Code is read more often than it is written.",
	"Walking on water and developing software from a specification are easy if both are frozen.",
	"Measuring programming progress by lines of code is like measuring aircraft building progress by weight.",
	"Talk is cheap. Show me the code.",
	"Programs must be written for people to read, and only incidentally for machines to execute.",
	"The most disastrous thing you can ever learn is your first programming language.",
}

var debuggingFortunes = []string{
	"It works on my machine!",
	"That's weird...",
	"But it worked yesterday!",
	"It must be a compiler bug.",
	"Have you tried turning it off and on again?",
	"It's probably a race condition.",
	"That should never happen.",
	"The tests pass locally!",
	"Works in dev, breaks in prod.",
	"I didn't change anything!",
	"Let me just add a print statement...",
	"99 little bugs in the code, 99 little bugs. Take one down, patch it around... 127 little bugs in the code.",
	"The bug is not in my code, it's in the framework... probably.",
	"I'll fix it in the next sprint.",
	"It's not a priority right now.",
	"We'll refactor that later.",
	"The documentation is wrong.",
	"That's undefined behavior.",
	"It's probably a caching issue.",
	"Clear your browser cache.",
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "fortune",
		Short:       "Tech fortune cookies",
		Long:        "Display a random tech-related fortune cookie or programming wisdom.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			Run(params)
		},
	}.ToCobra()
}

func Run(params *Params) {
	var pool []string

	switch params.Category {
	case "tech":
		pool = techFortunes
	case "wisdom":
		pool = wisdomFortunes
	case "debugging", "debug":
		pool = debuggingFortunes
	default:
		pool = append(pool, techFortunes...)
		pool = append(pool, wisdomFortunes...)
		pool = append(pool, debuggingFortunes...)
	}

	fortune := pool[rand.Intn(len(pool))]
	fmt.Printf("\n  %s\n\n", fortune)
}
