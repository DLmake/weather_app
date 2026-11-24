// Package vreme stores weather stations and their temperature
// measurements, and reports simple statistics about them.
//
// Measurements are validated against the package-level limits MinTemp and
// MaxTemp, and statistics are only computed once a station has at least
// StMeritev measurements.
package vreme

import (
	"fmt"
)

// Postaja is a weather station. Its measurements are the temperatures
// (in °C) recorded with DodajMeritev.
type Postaja struct {
	ime     string
	kraj    string
	meritve []float64
}



// slovar is an unused package-level station map.
var slovar = make(map[string]Postaja)
// MinTemp is the lowest temperature, in °C, that DodajMeritev accepts.
var	MinTemp   float64 = -50

// MaxTemp is the highest temperature, in °C, that DodajMeritev accepts.
var	MaxTemp   float64 = 50
// StMeritev is the minimum number of measurements a station needs
// before statistics are computed for it.
var	StMeritev int = 3

// DodajMeritev appends the temperature temperatura, in °C, to the station
// with key oznaka in postaje. It returns an error if the temperature is
// outside [MinTemp, MaxTemp] or if no station has that key. A rejected
// measurement is not stored.
func DodajMeritev(postaje map[string]Postaja, oznaka string, temperatura float64) error{
	//preveri če je mertev validna
	if temperatura > MaxTemp || temperatura < MinTemp{
		return fmt.Errorf("invalid range for temperature \n")
		
	}

	//prveri če obstaja v postajah
	postaja, obstaja := postaje[oznaka]
	if obstaja {
		//doda postaji meritev
		postaja.meritve = append(postaja.meritve, temperatura)
		postaje[oznaka] = postaja
		fmt.Printf("added the temperature \n")
		return nil

	} else {
		return fmt.Errorf("postaja doesnt exist \n")
		

	}
	return nil
}

// statistika returns the lowest, highest and average measurement of the
// station with key oznaka. ok is false if the station does not exist or
// has fewer than StMeritev measurements; the other results are then zero.
func statistika(postaje map[string]Postaja, oznaka string) (najnizja, najvisja, povprecna float64, ok bool) {
	postaja, obstaja := postaje[oznaka]

	if obstaja {
		var copi_mertive = postaja.meritve
		if len(copi_mertive) < StMeritev {
			ok = false
			fmt.Errorf("the postaja ima premalo meritev %d \n",len(copi_mertive))
			return
		}
		najnizja = copi_mertive[0]
		najvisja = copi_mertive[0]
		povprecna = 0
		ok = true

		for i := range copi_mertive {

			if copi_mertive[i] > najvisja {
				najvisja = copi_mertive[i]
			}

			if copi_mertive[i] < najnizja {
				najnizja = copi_mertive[i]
			}

			povprecna += copi_mertive[i]

		}
		povprecna = povprecna / float64(len(copi_mertive))

		return

	} else {
		fmt.Print("the postaja doesnt exist \n")
		ok = false
		return
	}
}

// IzpisMeritev prints every station's measurements to standard output,
// followed by its statistics when it has enough measurements.
func IzpisMeritev(postaje map[string]Postaja) {
	fmt.Println("MERITVE:")
	for kljuc, postaja := range postaje {
		fmt.Printf("%s - %s (%s): %v\n", kljuc, postaja.ime, postaja.kraj, postaja.meritve)

		najnizja, najvisja, povprecna, ok := statistika(postaje, kljuc)
		if !ok {
			fmt.Printf("   premalo meritev za statistiko (%d/%d)\n", len(postaja.meritve), StMeritev)
			continue
		}
		fmt.Printf("   min: %.1f  max: %.1f  povprečje: %.1f\n", najnizja, najvisja, povprecna)
	}
}
// DodajPostajo adds a new station with name ime_postaje, town ime_kraj and
// key kljuc to postaje. It returns an error if kljuc is already in use.
func DodajPostajo(ime_postaje string,ime_kraj string,kljuc string, postaje map[string]Postaja) error{

	//postaje so kot ze referenca dodajanje in umikanje caller vidi
	if _, obstaja := postaje[kljuc]; obstaja {
        return fmt.Errorf("postaja s ključem %q že obstaja", kljuc)
    }
	postaje[kljuc] = Postaja{ime : ime_postaje, kraj : ime_kraj}
	return nil




}


// IzpisPorocila prints a report for every station that has enough
// measurements: its maximum, minimum and average temperature, and a
// category for the average (Vroče, toplo, hladno or mrzlo).
func IzpisPorocila(postaje map[string]Postaja) {

	for i := range postaje {
		var postaja = postaje[i]
		var najnizja, najvisja, povprecna, uspeh = statistika(postaje, i)
		if !uspeh {
			fmt.Printf("statistika ni usepla zaradi preamlo meritev %s \n", i)
			continue
		}
		var temperatura = "error"
		switch {
		case povprecna >= 25:
			temperatura = "Vroče"
		case povprecna >= 15:
			temperatura = "toplo"
		case povprecna >= 5:
			temperatura = "hladno"
		default:
			temperatura = "mrzlo"
		}

		fmt.Printf("%s - %s (%s): max %.1f, min %.1f, povp %.1f — %s\n",
			i, postaja.ime, postaja.kraj, najvisja, najnizja, povprecna, temperatura)

	}

}

