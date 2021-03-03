package main

import (
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type output struct {
	w *widget.Label
}

func (o *output) Print(args ...interface{}) {
	o.w.SetText(o.w.Text + fmt.Sprint(args...))
}

func (o *output) Println(args ...interface{}) {
	o.w.SetText(o.w.Text + fmt.Sprintln(args...))
}

func (o *output) Printf(s string, args ...interface{}) {
	o.w.SetText(o.w.Text + fmt.Sprintf(s, args...))
}

var out output

var in *widget.Entry

var button *widget.Button

func readline(o interface{}) {
	tapped := make(chan bool)
	button.OnTapped = func() {
		switch typedout := o.(type) {
		case *string:
			// Error usually EOF, which shouldn't even happen with Stdin
			// default return value should be the right thing nevertheless.
			*typedout = strings.TrimRight(in.Text, "\n")
		case *int:
			// when ParseInt returns an error, the default value it returns
			// is still acceptable for a game situation, so ignore it
			intval, _ := strconv.ParseInt(regexp.MustCompile(`[^0-9]`).ReplaceAllString(in.Text, ""), 10, 0)
			*typedout = int(intval) // conversion from int64 to int, should work because of the last 0 argument in ParseInt
		default:
			panic("readline invoked with non-pointer type or pointer to unsupported type.")
		}
		out.w.SetText("")
		in.SetText("")
		tapped <- true
	}
	<-tapped
}

func klariguLudon(min, max, skip int) {
	out.Printf(`Bonvenon ĉe "Kaptu la muson"

La celo de la ludo estas diveni numeron de %d ĝis %d.
Vi povas provi tion unu post la alia.
%s
Se la numero fariĝas pli alta ol %d, ĝi reiras al %d.
Se la numero fariĝas malpli alta ol %d, ĝi reiras al %d.
Multan sukceson!

`, min, max, skipDoes(skip), max, min, min, max)
}

func skipDoes(skip int) string {
	switch {
	case skip > 0:
		return fmt.Sprintf("Tamen atentu! La numero povas post ĉies vico ĉu %d supreniri, ĉu tiom malsupreniri, ĉu resti sama.", skip)
	case skip == 0:
		return fmt.Sprint("Vi elektis ke la numero ĉiam restu la sama.")
	case skip < 0:
		return fmt.Sprintf("Vi elektis veran defion. La numero povas post ĉies vico en hazarda kvanto de maksimume %d supreniri, ĉu samkvante malsupreniri, ĉu resti la sama.", -skip)
	}
	panic("This should never happen.")
}

func askPlayers() (n int) {
	out.Print("Kiom da ludantoj estas? ")
	readline(&n)
	if n < 1 {
		return 2
	}
	return n
}

func askNames(n int) (names []string) {
	for i := 1; i <= n; i++ {
		name := ""
		out.Printf("Ludanto %d, kiel mi nomu vin? ", i)
		readline(&name)
		names = append(names, name)
	}
	return names
}

func getNumber(min, max, skip int, guessed chan bool) (nchan chan int) {
	nchan = make(chan int, 0)
	go func() {
		i := rand.Intn(max-(min-1)) + (min - 1)
		s := rand.Intn(3) - 1
		skipn := skip
		if skip < 0 {
			skipn = rand.Intn(-skip)
		}
		for {
			select {
			case nchan <- i:
			case <-guessed:
				guessed <- true
				return
			}
			i += s * skipn
			switch {
			case i > max:
				i = min
			case i < min:
				i = max
			}
		}
	}()
	return nchan
}

func play(name string, number int) (won bool) {
	out.Printf("Estas la vico de %s: kiu numero estas? ", name)
	var guess int
	readline(&guess)
	switch {
	case number == guess:
		out.Println("Prave, la numero estas", number)
		return true
	case number > guess:
		out.Println("Ne, la numero estas pli alta ol", guess)
	case number < guess:
		out.Println("Ne, la numero estas pli malalta ol", guess)

	}
	return false
}

func again() bool {
	for {
		out.Print("Ĉu ludi denove, jes aŭ ne? ")
		answer := ""
		readline(&answer)
		switch answer {
		case "jes":
			return true
		case "ne":
			return false
		}
	}
}

func main() {
	a := app.New()
	w := a.NewWindow("Muskapti")
	w.Resize(fyne.NewSize(640,480))

	out.w = widget.NewLabel("Bonvenon al Muskapti! Ĉi tiu ludo baldaŭ komenciĝas. Se io misfunkcias, provu ŝanĝi la grandecon de la ekrano.")
	out.w.Wrapping = fyne.TextWrapWord
	in = widget.NewEntry()
	button = widget.NewButton("Ek!", nil)
	in.OnSubmitted = func (_ string) {
		button.OnTapped()
	}
	c := container.NewVBox(out.w, in, button)
	w.SetContent(c)
	out.w.SetText("")

	go func() {
		rand.Seed(int64(time.Now().Nanosecond()))
		min, max, skip := 1, 100, 1
		switch len(os.Args) {
		case 0:
			panic("This shouldn't happen: os.Args is an empty slice.")
		case 1:
			out.Print("(Vi ludas kun la defaŭltaj valoroj. Se vi volas pli grandan defion, provu ŝanĝi la parametrojn! Nur unu parametro indikas la plej altan numeron, se vi indikas pli tio estas sinsekve minimumo, maksimumo, saltogrando. Negativa saltogrando hazardigas la saltograndon je ĝia absoluta valoro. Bonan muskaptadon!)\n\n")
		case 2:
			fmt.Sscan(os.Args[1], &max)
		case 3:
			fmt.Sscan(os.Args[1]+" "+os.Args[2], &min, &max)
		default:
			fmt.Sscan(os.Args[1]+" "+os.Args[2]+" "+os.Args[3], &min, &max, &skip)
		}
		klariguLudon(min, max, skip)
		players := askPlayers()
		names := askNames(players)
		guessed := make(chan bool)
		numbers := getNumber(min, max, skip, guessed)
		for {
			for _, name := range names {
				select {
				case number := <-numbers:
					won := play(name, number)
					if won {
						out.Printf("Gratulon %s, vi gajnis!\n", name)
						if again() {
							guessed <- true
						} else {
							out.Println("Ĝis revido!")
							button.SetText("Ĝis!")
							button.OnTapped = func() {
								w.Close()
							}
							select{}
						}
					}
				case <-guessed:
					numbers = getNumber(min, max, skip, guessed)
				}
			}
		}
	}()
	w.ShowAndRun()
}
