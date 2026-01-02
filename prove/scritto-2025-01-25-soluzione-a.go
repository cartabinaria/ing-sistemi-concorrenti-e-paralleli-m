package main

import (
	"fmt"
	"math/rand"
	"time"
)

// COSTANTI:
const MAXBUFF = 100
const MAXPROC = 10
const NT = 4
const MAX = 8
const FUN = 0
const FISIO = 1

type Richiesta struct {
	id  int
	ack chan int
}

var Area [2]string = [2]string{"FUN", "FISIO"}

// CANALI:
var entraUtente [2]chan Richiesta
var entraBagnino = make(chan Richiesta, MAXBUFF)
var uscitaUtente [2]chan Richiesta
var uscitaBagnino = make(chan Richiesta, MAXBUFF)

// CANALI per terminazione:
var done = make(chan bool)
var termina = make(chan bool)
var TerminaCentro = make(chan bool)

//  funzioni:
func when(b bool, c chan Richiesta) chan Richiesta {
	if !b {
		return nil
	}
	return c
}

func sleepRandTime(timeLimit int) {
	if timeLimit > 0 {
		time.Sleep(time.Duration(rand.Intn(timeLimit)+1) * time.Second)
	}
}

// GOROUTINE:
func Utente(id int) {
	fmt.Printf("[Utente %d] Inizio\n", id)
	volte := rand.Intn(4) //Numero di Entrata-Uscita che compie ogni utente
	for i := 0; i < volte; i++ {
		fmt.Printf("Utente %d , %d giro\n", id, i+1)
		tipo := rand.Intn(2) //Zona nella quale entrerà l'utente, o FUN o FISIO, può cambiare ad ogni iterazione
		r := Richiesta{id, make(chan int)}
		fmt.Printf("[Utente %d] Vuole entrare in area %s\n", id, Area[tipo])
		entraUtente[tipo] <- r
		<-r.ack

		sleepRandTime(7) //TEMPO ARBITRARIO NEL QUALE STA NELLA ZONA SCELTA

		uscitaUtente[tipo] <- r
		<-r.ack
		fmt.Printf("[Utente %d] Uscito dall'area %s\n", id, Area[tipo])
		sleepRandTime(2)

	}

	done <- true
}

func bagnino(id int) { //TERMINA CON UN SEGNALE CHE DEVO MANDARE IO NEL MAIN
	fmt.Printf("Bagnino %d inizia l'orario lavorativo\n", id)
	for true {

		r := Richiesta{id, make(chan int)}

		entraBagnino <- r
		res := <-r.ack

		if res == -1 {
			fmt.Printf("Bagnino %d deve terminare\n", id)
			done <- true
			return
		} else {
			fmt.Printf("Bagnino %d entrato ..\n", id)
		}

		sleepRandTime(10) //IL BAGNINO PRESIDIA ...
		uscitaBagnino <- r
		<-r.ack
		fmt.Printf("Bagnino %d  uscito ..\n", id)
		sleepRandTime(2)
	}
}

//SERVER:

func server() {
	// variabili di stato:
	nFUN := 0            //Numero utenti in zona FUN
	nFisio := 0          //Numero utenti in zona FISIO
	Fisioterapisti := NT //Numero di fisioterapisti liberi in zona Fisio
	nBagnini := 0        //Numero di bagnini presenti dentro l'area FUN
	fine := false

	fmt.Printf("[SERVER] Inizio\n")

	for {
		fmt.Printf("\nSTATO ATTUALE:\n Ci sono %d utenti in zona FUN e %d in zona FISIO\n", nFUN, nFisio)
		fmt.Printf("Ci sono %d fisioterapisti liberi e %d bagnini a presidiare\n", Fisioterapisti, nBagnini)
		fmt.Printf("\nCode : Entrata Bagnini %d \n Entrata Utenti FUN %d \n Entrata Utenti Fisio %d\n Uscita Bagnini %d \n\n ", len(entraBagnino), len(entraUtente[FUN]), len(entraUtente[FISIO]), len(uscitaBagnino))

		select {

		case r := <-when(fine == false, entraBagnino):
			nBagnini++
			r.ack <- 1
		case r := <-when(nFUN+nFisio < MAX && nBagnini > 0 && len(entraBagnino) == 0, entraUtente[FUN]):
			nFUN++
			r.ack <- 1
		case r := <-when(fine == true, entraBagnino): //la condizione di fine andrà solo a intaccare i bagnini in quanto diventerà vera solo quando tutti gli Utenti saranno terminati
			r.ack <- -1
		case r := <-when(nFUN+nFisio < MAX && Fisioterapisti > 0 && len(entraUtente[FUN]) == 0, entraUtente[FISIO]):
			nFisio++
			Fisioterapisti--
			r.ack <- 1
		case r := <-when(nBagnini > 1 || nFUN == 0, uscitaBagnino):
			nBagnini--
			r.ack <- 1
		case r := <-uscitaUtente[FUN]:
			nFUN--
			r.ack <- 1
		case r := <-uscitaUtente[FISIO]:
			nFisio--
			Fisioterapisti++
			r.ack <- 1
		case <-TerminaCentro: // questo messaggio arriva quando tutti gli utenti sono usciti definitamente dal centro, e posso far terminare i bagnini
			fine = true
			fmt.Printf("Il Centro sta per chiudere.. \n")
		case <-termina:
			fmt.Println("Il Centro è chiuso!\n")
			done <- true
			return
		}
	}
}

func main() {
	fmt.Printf("[MAIN] Inizio\n\n")
	rand.Seed(time.Now().UnixNano())

	// Inizializzazione canali
	for i := 0; i < 2; i++ {
		entraUtente[i] = make(chan Richiesta, MAXBUFF)
		uscitaUtente[i] = make(chan Richiesta, MAXBUFF)
	}

	// Esecuzione goroutine
	go server()

	for i := 0; i < MAXPROC; i++ {
		go Utente(i)
	}
	for i := 0; i < MAXPROC/2; i++ {
		go bagnino(i)
	}
	// Join goroutine
	for i := 0; i < MAXPROC; i++ {
		<-done
	}
	fmt.Printf("\nTutti gli utenti sono terminati!\n \n")
	TerminaCentro <- true
	for i := 0; i < MAXPROC/2; i++ { //attesa bagnini
		<-done
	}
	fmt.Printf("\nTutti i bagnini sono terminati!\n \n")
	// chiudo:
	termina <- true
	<-done
	fmt.Printf("\n\n[MAIN] Fine\n")
}
