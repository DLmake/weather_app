// Command vreme is an interactive console program for adding weather
// stations and temperature measurements and printing their statistics.
//
// Run it without arguments to start the vreme> prompt. Type "help" for the
// list of commands and "izhod" or "exit" to quit.
package main
import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"github.com/DLmake/weather_app/vreme"

    "github.com/urfave/cli/v3"
)

// zgradiUkaz builds the command tree for the vreme program. The commands
// operate on the stations in svc.
func zgradiUkaz(svc map[string]vreme.Postaja) *cli.Command {
	return &cli.Command{
		Name: "Vreme",
		Usage: "upravljanje meritev in postaj",
		Writer: os.Stdout,
		Commands: []*cli.Command{
			//first command
			{
				Name: "dodaj_postajo",
				Usage: "doda novo postajo (kakor piše)",
				ArgsUsage: "<ime_postaje> <ime_kraja> <kljuc>",
				Arguments: []cli.Argument{
					&cli.StringArg{Name: "ime_postaje", Required: true},
					&cli.StringArg{Name: "ime_kraja", Required: true},
					&cli.StringArg{Name: "kljuc", Required: true},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
						vreme.DodajPostajo(cmd.StringArg("ime_postaje"), cmd.StringArg("ime_kraja"), cmd.StringArg("kljuc"), svc)
						fmt.Println("station added")
						return nil
					},



			},

			//second command
			{
				Name: "dodaj_meritve",
				Usage: "doda postaji meritev",
				ArgsUsage: "<kljuc> <meritev>",
				Arguments: []cli.Argument{
					&cli.StringArg{Name: "kljuc", Required: true},
					&cli.Float64Arg{Name: "meritev", Required: true},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
						vreme.DodajMeritev(svc,cmd.StringArg("kljuc"),cmd.Float64Arg("meritev"))
						fmt.Println("meritev added")
						return nil
					},




			},


			//command three

			{
				Name: "izpis_meritev",
				Usage: "izpiše vse meritve",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					vreme.IzpisMeritev(svc)
					fmt.Println("koncan izpis")
					return nil
				},
			},
		},
	}
}

// zanka runs the interactive prompt, reading one command per line from
// standard input until the user types "izhod" or "exit" or input ends.
func zanka(postaje map[string]vreme.Postaja) error {
	zgradiUkaz(postaje).Run(context.Background(), []string{"Vreme", "help"})
	fmt.Println(`Vpiši ukaz, "help" za pomoč ali "izhod" za konec.`)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("vreme> ")
		if !scanner.Scan() {
			break
		}
		vrstica := strings.TrimSpace(scanner.Text())
		if vrstica == "" {
			continue
		}
		if vrstica == "izhod" || vrstica == "exit" {
			break
		}
		args := append([]string{"Vreme"}, strings.Fields(vrstica)...)
		if err := zgradiUkaz(postaje).Run(context.Background(), args); err != nil {
			fmt.Println("napaka:", err)
		}
	}
	fmt.Println("Nasvidenje!")
	return nil
}



// main parses the global flags, validates them and starts the prompt.
func main() {
	postaje := map[string]vreme.Postaja{}
/*
	slovar["LJ"] = Postaja{ime: "Bežigrad", kraj: "Ljubljana"}
	slovar["MB"] = Postaja{ime: "Tabor", kraj: "Maribor"}
	slovar["KP"] = Postaja{ime: "Koper Luka", kraj: "Koper"}
	slovar["CE"] = Postaja{ime: "Celjski Grad", kraj: "Celje"}
	slovar["AT"] = Postaja{ime: "antartika", kraj: "kranjska gora"}

	//DODAJANJE MERTIEV
	dodajMeritev(slovar, "LJ", 12.5)
	dodajMeritev(slovar, "LJ", 15.0)
	dodajMeritev(slovar, "LJ", 9.3)
	dodajMeritev(slovar, "LJ", 20.3)
	dodajMeritev(slovar, "LJ", 30.3)

	dodajMeritev(slovar, "MB", 26.0)
	dodajMeritev(slovar, "MB", 28.4)
	dodajMeritev(slovar, "MB", 25.1)
	dodajMeritev(slovar, "MB", -20.1)
	dodajMeritev(slovar, "MB", -5.1)
	dodajMeritev(slovar, "MB", 10.1)

	//permal meritev tkoda vrne error
	dodajMeritev(slovar, "KP", 20.0)
	dodajMeritev(slovar, "KP", 25.0)

	dodajMeritev(slovar, "CE", 40.0)
	dodajMeritev(slovar, "CE", 40.0)
	dodajMeritev(slovar, "CE", 40.0)
	dodajMeritev(slovar, "CE", 40.0)

	dodajMeritev(slovar, "AT", -25)
	dodajMeritev(slovar, "AT", -25)
	dodajMeritev(slovar, "AT", -25)
	dodajMeritev(slovar, "AT", -25)

	//NAPAČNE MERITVE
	dodajMeritev(slovar, "LJ", 80)   // previsoko
	dodajMeritev(slovar, "XX", 10)   // neobstaja povezava
	dodajMeritev(slovar, "LJ", -100) //prenizko

	//sprinate stvari
	fmt.Println()
	izpisMeritev(slovar)
	fmt.Println()
	izpisPorocila(slovar)
*/
	//usvarim mapo z novimi postajami
	zagon := &cli.Command{
		Name:  "Vreme",
		Usage: "upravljanje meritev in postaj",
		Flags: []cli.Flag{
			&cli.FloatFlag{Name: "minTemp", Value: -50, Usage: "najnižja dovoljena temperatura meritve (°C)"},
			&cli.FloatFlag{Name: "maxTemp", Value: 50, Usage: "najvišja dovoljena temperatura meritve (°C)"},
			&cli.IntFlag{Name: "stMeritev", Value: 3, Usage: "najmanjše število meritev za izračun statistike"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			vreme.MinTemp = cmd.Float("minTemp")
			vreme.MaxTemp = cmd.Float("maxTemp")
			vreme.StMeritev = cmd.Int("stMeritev")

			if vreme.MinTemp >= vreme.MaxTemp {
				return fmt.Errorf("minTemp mora biti manjši od maxTemp")
			}
			if vreme.StMeritev < 1 {
				return fmt.Errorf("stMeritev mora biti vsaj 1")
			}
			return zanka(postaje)
		},
	}
	if err := zagon.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "napaka:", err)
		os.Exit(1)
	}

	//definiramo naše komande
	//kle mamo main app loop kopirano iz primera


}

