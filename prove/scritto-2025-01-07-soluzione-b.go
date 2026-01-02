//esame 7 Gennaio 2025 -tema B

package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const MAXBUFF = 100
const MAXPROC = 10
const MAXCICLI = 4

const AREAPESI = 0
const AREACORSI = 1
const NumAree = 2 //numero di tipi di risorse

const NP = 15  // Numero Massimo di Persone per l'area Pesi
const NT = 5   //  numero di personal trainer
const MAX = 18 // capacità num massimo di utenti

type Richiesta struct {
	id   int
	tipo int
	ack  chan bool
}

type Trainer struct {
	dentro          bool
	vuoleUscire     bool
	utenteAssegnato int
	ackUscita       chan bool
}

var IngressoArea [NumAree]chan Richiesta
var Uscita = make(chan Richiesta, MAXBUFF)

var IngressoPT = make(chan Richiesta, MAXBUFF)
var UscitaPT = make(chan Richiesta)

var done = make(chan bool)
var termina = make(chan bool)
var terminaServer = make(chan bool)

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

func getTipo(t int) string {
	if t == AREAPESI {
		return "Area Pesi"
	} else if t == AREACORSI {
		return "Area Corsi"
	} else {
		return ""
	}
}

func utente(id int) {
	fmt.Printf("[UTENTE %d] Start...\n", id)
	r := Richiesta{id, -1, make(chan bool, MAXBUFF)}

	cicli := rand.Intn(MAXCICLI) + 1

	for i := 0; i < cicli; i++ {
		//Sceglie il tipo di attività da svolgere
		tipo := rand.Intn(NumAree)
		r.tipo = tipo

		fmt.Printf("[UTENTE %d] chiedo di entrare in %s\n", id, strings.ToUpper(getTipo(tipo)))
		IngressoArea[tipo] <- r
		<-r.ack

		fmt.Printf("[UTENTE %d] mi sto allenando in %s...\n", id, strings.ToUpper(getTipo(tipo)))
		sleepRandTime(5)

		fmt.Printf("[UTENTE %d] esco da %s\n", id, strings.ToUpper(getTipo(tipo)))
		Uscita <- r
		<-r.ack

	} //fine ciclo

	fmt.Printf("[UTENTE %d] me ne vado\n", id)
	done <- true
}
func trainer(id int) {

	var ric Richiesta
	ric.id = id
	ric.tipo = AREAPESI
	ric.ack = make(chan bool, MAXBUFF)

	for {
		sleepRandTime(5)

		fmt.Printf("[PT %d] chiedo di entrare in AREA CORSI ...\n", id)

		IngressoPT <- ric
		<-ric.ack

		fmt.Printf("[PT %d] sono entrato ...\n", id)

		sleepRandTime(15)

		UscitaPT <- ric
		<-ric.ack

		fmt.Printf("[PT %d] sono uscito ...\n", id)

		select {
		case <-termina:
			{
				fmt.Printf("[PT %d] ho finito!\n", id)
				done <- true
				return
			}
		default:
			{
				// non terminare
				sleepRandTime(2)
			}
		}
	}

}

func palestra() {

	utentiInPalestra := 0
	utentiInAP := 0                //quante persone in area Pesi
	trainer := make([]Trainer, NT) //stato dei trainer

	for i := 0; i < NT; i++ {
		trainer[i].dentro = false
		trainer[i].vuoleUscire = false
		trainer[i].utenteAssegnato = -1
		trainer[i].ackUscita = nil
	}

	trainerLiberi := 0
	trainerDentro := 0

	fmt.Printf("[PALESTRA] Apertura! \n")
	for {
		select {
		case r := <-when(utentiInPalestra < MAX && utentiInAP < NP && len(IngressoArea[AREACORSI]) == 0, IngressoArea[AREAPESI]):
			utentiInPalestra++
			utentiInAP++
			fmt.Printf("[PALESTRA] utente %d  entrato in area Pesi.\n", r.id)
			r.ack <- true
		case r := <-when(utentiInPalestra < MAX && trainerLiberi > 0 && len(IngressoPT) == 0, IngressoArea[AREACORSI]):
			utentiInPalestra++
			//ricerca e assegnazione PT:
			found := false
			i := 0
			for i = 0; i < NT && !found; i++ {
				if trainer[i].utenteAssegnato == -1 && trainer[i].dentro {
					found = true
					trainer[i].utenteAssegnato = r.id
				}
			}
			trainerLiberi--
			fmt.Printf("[PALESTRA] l' utente %d è in area Corsi e sta allenandosi con il trainer %d.\n", r.id, i)
			r.ack <- true
		case r := <-IngressoPT:
			fmt.Printf("[PALESTRA] entrato il trainer %d.\n", r.id)
			trainer[r.id].dentro = true
			trainer[r.id].vuoleUscire = false
			trainer[r.id].utenteAssegnato = -1
			trainer[r.id].ackUscita = nil
			trainerDentro++
			trainerLiberi++
			r.ack <- true
		case r := <-Uscita:

			utentiInPalestra--
			fmt.Printf("[PALESTRA] utente %d in uscita da %s\n", r.id, strings.ToUpper(getTipo(r.tipo)))

			if r.tipo == AREACORSI {
				found := false
				for i := 0; i < NT && !found; i++ {
					if trainer[i].utenteAssegnato == r.id {
						found = true
						trainer[i].utenteAssegnato = -1
						trainerLiberi++
						if trainer[i].vuoleUscire && trainer[i].dentro {
							//permetto al trainer i di uscire:
							fmt.Printf("[PALESTRA] Il trainer %d può uscire dalla palestra...\n", i)
							trainer[i].dentro = false
							trainer[i].vuoleUscire = false
							trainer[i].ackUscita <- true
							trainer[i].ackUscita = nil
							trainerDentro--
							trainerLiberi--
						}
					}
				}
			} else {
				utentiInAP--
			}
			r.ack <- true
		case ric := <-UscitaPT:
			fmt.Printf("[PALESTRA] il trainer %d chiede di uscire..\n", ric.id)
			if trainer[ric.id].utenteAssegnato == -1 {
				// il trainer è libero e puo' uscire
				fmt.Printf("[PALESTRA] il trainer %d è libero ed  esce dalla palestra...\n", ric.id)
				trainer[ric.id].dentro = false
				trainer[ric.id].vuoleUscire = false
				trainer[ric.id].ackUscita = nil
				trainerLiberi--
				trainerDentro--
				ric.ack <- true
			} else {
				// il trainer è impegnato e attende
				fmt.Printf("[PALESTRA] Il trainer %d è occupato e attende di uscire dalla palestra\n", ric.id)
				trainer[ric.id].vuoleUscire = true
				trainer[ric.id].ackUscita = ric.ack
			}

		case <-terminaServer:
			fmt.Printf("[PALESTRA] Termino\n")
			done <- true
			return
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	var nUtenti int
	var nTrainer int

	nUtenti = 50
	nTrainer = NT

	for i := 0; i < len(IngressoArea); i++ {
		IngressoArea[i] = make(chan Richiesta, MAXBUFF)
	}

	go palestra()
	//creazione trainer:
	for i := 0; i < nTrainer; i++ {
		go trainer(i)
	}
	//creazione utenti:
	for i := 0; i < nUtenti; i++ {
		go utente(i)
	}
	//attesa terminazione utenti:
	for i := 0; i < nUtenti; i++ {
		<-done
	}
	// terminazione Personal trainer:
	for i := 0; i < nTrainer; i++ {
		termina <- true
		<-done
	}
	// terminazione server:
	terminaServer <- true
	<-done

	fmt.Printf("\n\n[MAIN] Chiusura palestra!\n")
}
