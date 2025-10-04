package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/gocolly/colly"
)

type play struct {
	lotto_zahlen  []int
	zusatz_zahlen []int
}

var play_list []play
var datum string
var lottotype string
var anzahl_spiele int
var wg sync.WaitGroup

func init() {
	now := time.Now().UnixNano()
	rand.New(rand.NewSource(now))

	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "?" || os.Args[1] == "--help") {
		fmt.Println("Hilfe angefordert !!!")
		fmt.Println("USAGE :")
		fmt.Printf("\tlotto [-flag : value] ...\n")
		fmt.Println("\t-g	: game. Momentan lotto und eurojackpot")
		fmt.Println("\t		: lotto -g lotto")
		fmt.Println("\t		: lotto -g eurojackpot")
		fmt.Println("\t-d	: datum. ")
		fmt.Println("\t		: lotto -g lotto -d 01.10.2025")
		fmt.Println("\t-c	: count. Wie viele Spielfelder sollen erstellt werden?")
		fmt.Println("\t		: lotto -g lotto -d 01.10.2025 -c 10000")
		fmt.Println("\t		: Erstellt 10000 Spielfelder und vergleicht sie mit dem")
		fmt.Printf("\t		: Ergebnis vom angegebenen Datum\n\n")
		os.Exit(0)
	}

	flag.StringVar(&lottotype, "g", "eurojackpot", "") // -g lotto / -g eurojackpot
	flag.StringVar(&datum, "d", "03.10.2025", "")      // -d 27.09.2025 / -d 01.10.2025
	flag.IntVar(&anzahl_spiele, "c", 2, "")            // -c 10 / -c 1000000
	flag.Parse()

	// if len(os.Args) > 1 {
	// 	anzahl := os.Args[1]
	// 	anzahl_spiele, _ = strconv.Atoi(anzahl)
	// 	datum = os.Args[2]
	// } else {
	// 	anzahl_spiele = 10
	// }

}

func newPlay() play {
	return play{
		lotto_zahlen:  make([]int, 0),
		zusatz_zahlen: make([]int, 0),
	}
}

func (p *play) sortSpielZahlen() {
	sort.Ints(p.lotto_zahlen)
	sort.Ints(p.zusatz_zahlen)
}

func (p *play) printSpielZahlen() {
	fmt.Println("####### Lotto ##########")
	for i := 0; i < len(p.lotto_zahlen); i++ {
		fmt.Print(p.lotto_zahlen[i])
		fmt.Print(" ")
	}
	fmt.Println()
	fmt.Println("####### Zusatz ##########")
	for i := 0; i < len(p.zusatz_zahlen); i++ {
		fmt.Print(p.zusatz_zahlen[i])
		fmt.Print(" ")
	}
	fmt.Println()
	fmt.Println()
}

func main() {
	fmt.Println("Hello World!!!")

	fmt.Println("datum: ", datum)
	if lottotype != "" && lottotype == "lotto" {
		fmt.Println(" LOTTO FLAG: ", lottotype)
		lotto()
		return
	}
	field := make(chan play)
	for x := 1; x <= anzahl_spiele; x++ {
		wg.Add(1)
		go fill_up_spiel(&wg, field)
	}

	go recieveData(field)

	wg.Wait()

	printInfo("Erstellt Daten: ")
	gewinn_zahlen := get_lotto_gewinnzahlen()
	check_match(gewinn_zahlen)
}

func printInfo(s string) {
	fmt.Printf("\n%s %d\n\n", s, len(play_list))

}

func recieveData(field chan play) {
	for one_play := range field {
		one_play.sortSpielZahlen()
		play_list = append(play_list, one_play)
	}
}

func fill_up_spiel(wg *sync.WaitGroup, f chan play) {
	defer wg.Done()
	min_lotto := 1
	max_lotto := 50
	max_zusatz := 10
	one_play := newPlay()
	zahlen_map := make(map[int]bool)
	zusatz_map := make(map[int]bool)
	for i := 0; i < 5; i++ {
		zahl := get_random_number(max_lotto, min_lotto)
		if check_number_exist(zahlen_map, zahl) {
			i = i - 1
			//			fmt.Printf("Zahl %d ist schon da. \n", zahl)
			continue
		}
		zahlen_map[zahl] = true
		one_play.lotto_zahlen = append(one_play.lotto_zahlen, zahl)
	}

	for i := 0; i < 2; i++ {
		zahl := get_random_number(max_zusatz, min_lotto)
		if check_number_exist(zusatz_map, zahl) {
			i = i - 1
			//fmt.Printf("Zusatzzahl %d ist schon da. \n", zahl)
			continue
		}
		zusatz_map[zahl] = true
		one_play.zusatz_zahlen = append(one_play.zusatz_zahlen, zahl)
	}
	//one_play.printSpielZahlen()
	f <- one_play
	//one_play.sortSpielZahlen()
	//play_list = append(play_list, one_play)

}

func check_match(gz play) {
	for i := 0; i < len(play_list); i++ {
		match_counter := 0
		for j := 0; j < len(gz.lotto_zahlen); j++ {
			if gz.lotto_zahlen[j] != play_list[i].lotto_zahlen[j] {
				//fmt.Printf("%d - BREAK !\n", i)
				break
			}
			match_counter++
		}
		if match_counter > 4 {
			fmt.Printf(" Match gefunden: %d !\n", i)
			play_list[i].printSpielZahlen()
			z_counter := 0
			for z := 0; z < 2; z++ {
				if play_list[i].zusatz_zahlen[z] != gz.zusatz_zahlen[z] {
					break
				}
				z_counter++
			}
			if z_counter == 2 {
				fmt.Println("Zusatzzahlen ok!")
			}
		}
	}
}

func get_lotto_gewinnzahlen() play {
	gewinnzahlen := newPlay()
	c := colly.NewCollector()
	//datum := "20.08.2024"
	c.OnHTML("div.col-md-12", func(e *colly.HTMLElement) {
		counter := 1
		coll := e.ChildAttrs("p", "class")
		for k := range coll {
			if coll[k] == "heading-h5" && e.ChildText("span.polygon-label") != "" {
				//fmt.Printf("Superzahl: %s\n", e.ChildText("span.polygon-label"))
				z, _ := strconv.Atoi(e.ChildText("span.polygon-label"))
				gewinnzahlen.zusatz_zahlen = append(gewinnzahlen.zusatz_zahlen, z)
			}

		}
		e.ForEach("li", func(_ int, e *colly.HTMLElement) {
			if e.Text != "" && len(gewinnzahlen.lotto_zahlen) < 5 {
				//fmt.Printf("Die Zahl: %s\n", e.Text)
				z, _ := strconv.Atoi(e.Text)
				gewinnzahlen.lotto_zahlen = append(gewinnzahlen.lotto_zahlen, z)
				counter++
			} else if e.Text != "" && len(gewinnzahlen.lotto_zahlen) > 4 {
				z, _ := strconv.Atoi(e.Text)
				//fmt.Printf("\t\tDie SZahl: %d\n", z)
				gewinnzahlen.zusatz_zahlen = append(gewinnzahlen.zusatz_zahlen, z)
				return
			}

		})
	})
	c.Visit(fmt.Sprintf("https://www.westlotto.de/eurojackpot/gewinnzahlen/gewinnzahlen.html?datum=%s", datum))
	gewinnzahlen.sortSpielZahlen()
	fmt.Println("*** Gewinnzahlen ****")
	gewinnzahlen.printSpielZahlen()

	return gewinnzahlen
}

func get_random_number(max, min int) int {
	return rand.Intn(max-min+1) + min
}

func check_number_exist(mappe map[int]bool, number int) bool {
	_, ok := mappe[number]
	return ok
}

func lotto() {
	gewinnzahlen := newPlay()
	c := colly.NewCollector()
	//datum := "20.08.2024"
	c.OnHTML("div.col-md-12", func(b *colly.HTMLElement) {
		//fmt.Println("gefunden ", b.Text)
		if b.ChildText("p.heading-h5") == "Superzahl" { //&& b.ChildText("span.polygon-label") == "Superzahl" {
			fmt.Println("gefunden Superzahl !!! ", b.Attr("class"))
			fmt.Printf("Superzahl: %s\n", b.ChildText("span.polygon-label"))
			z, _ := strconv.Atoi(b.ChildText("span.polygon-label"))
			gewinnzahlen.zusatz_zahlen = append(gewinnzahlen.zusatz_zahlen, z)
		}
	})
	c.Visit(fmt.Sprintf("https://www.westlotto.de/infos-und-zahlen/gewinnzahlen/lotto/gewinnzahlen-lotto.html?date=%s", datum))

	c = colly.NewCollector()
	c.OnHTML("div.col-md-12", func(e *colly.HTMLElement) {
		counter := 1
		e.ForEach("li", func(_ int, e *colly.HTMLElement) {
			if e.Text != "" && len(gewinnzahlen.lotto_zahlen) < 6 {
				//fmt.Printf("Die Zahl: %s\n", e.Text)
				z, _ := strconv.Atoi(e.Text)
				gewinnzahlen.lotto_zahlen = append(gewinnzahlen.lotto_zahlen, z)
				counter++
			}
		})
	})
	c.Visit(fmt.Sprintf("https://www.westlotto.de/infos-und-zahlen/gewinnzahlen/lotto/gewinnzahlen-lotto.html?date=%s", datum))
	gewinnzahlen.sortSpielZahlen()
	fmt.Println("*** Gewinnzahlen ****")
	gewinnzahlen.printSpielZahlen()

}
