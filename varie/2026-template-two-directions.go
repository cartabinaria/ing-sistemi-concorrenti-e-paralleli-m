// Davide Chirichella - 0001222371

package main

import (
	"fmt"
	"math/rand"
	"time"
)

/* Costanti */
const MAX_CAPACITY = 5

const TOT_OPERAZIONI = 15 // Condizione di terminazione

const NUM_TIPO_A = 4
const NUM_TIPO_B = 6

/* Canali Globali */
// Canali di richiesta (Entrata e Uscita per ogni tipo)
var req_entra_t1 = make(chan int, 100)
var req_esce_t1 = make(chan int, 100)
var req_entra_t2 = make(chan int, 100)
var req_esce_t2 = make(chan int, 100)

// Canali di ACK
var ack_t1 [NUM_TIPO_A]chan int
var ack_t2 [NUM_TIPO_B]chan int

// Canali di sincronizzazione terminazione
var done = make(chan bool)
var termina = make(chan bool)

/* 
 * Utils 
 */
func printTipo(val int) string {
	switch val {
	case TIPO_A:
		return "TIPO_A"
	case TIPO_B:
		return "TIPO_B"
	default:
		return ""
	}
}

/*
 * Guardia logica
 */
func when(cond bool, ch chan int) chan int {
	if !cond {
		return nil
	}
	return ch
}

/*
 * Processo Utente TIPO_A
 */
func processoTIPO_A(id int) {
	fmt.Printf("[TIPO_A %d]: Avviato\n", id)
	for {
		// Simulazione tempo fuori dall'area
		time.Sleep(time.Duration(rand.Intn(3)+1) * time.Second)

		// Richiesta Entrata
		req_entra_t1 <- id
		risposta := <-ack_t1[id]

		if risposta == -1 {
			fmt.Printf("[TIPO_A %d]: Ricevuto segnale di fine. Esco.\n", id)
			done <- true
			return
		}

		fmt.Printf("[TIPO_A %d]: Entrato nell'area\n", id)
		// Simulazione permanenza nell'area
		time.Sleep(time.Duration(rand.Intn(2)+1) * time.Second)

		// Richiesta Uscita
		req_esce_t1 <- id
		<-ack_t1[id]
		fmt.Printf("[TIPO_A %d]: Uscito dall'area\n", id)
	}
}

/*
 * Processo Utente TIPO_B
 */
func processoTIPO_B(id int) {
	fmt.Printf("[TIPO_B %d]: Avviato\n", id)
	for {
		time.Sleep(time.Duration(rand.Intn(3)+1) * time.Second)

		// Richiesta Entrata
		req_entra_t2 <- id
		risposta := <-ack_t2[id]

		if risposta == -1 {
			fmt.Printf("[TIPO_B %d]: Ricevuto segnale di fine. Esco.\n", id)
			done <- true
			return
		}

		fmt.Printf("[TIPO_B %d]: Entrato nell'area\n", id)
		time.Sleep(time.Duration(rand.Intn(2)+1) * time.Second)

		// Richiesta Uscita
		req_esce_t2 <- id
		<-ack_t2[id]
		fmt.Printf("[TIPO_B %d]: Uscito dall'area\n", id)
	}
}

/*
 * Server Gestore Area Condivisa
 */
func server() {
	// Stato interno
	var n1 = 0 // TIPO_A
	var n2 = 0 // TIPO_B
	var contatoreOperazioniTotali = 0
	var fine = false

	for {
		select {

		/* USCITA */
		case id := <-req_esce_t1:
			n1--
			ack_t1[id] <- 1

		case id := <-req_esce_t2:
			n2--
			ack_t2[id] <- 1

		/* ENTRATA */
		// TIPO_A
		case id := <-when(!fine && (n1+n2) < MAX_CAPACITY && n2 == 0, req_entra_t1):
			n1++
			contatoreOperazioniTotali++
			ack_t1[id] <- 1
			fmt.Printf("[SERVER]: Entrato T1. Stato: [T1:%d T2:%d]\n", n1, n2)

		// TIPO_B
		case id := <-when(!fine && (n1+n2) < MAX_CAPACITY && n1 == 0 && len(req_entra_t1) == 0, req_entra_t2):
			n2++
			contatoreOperazioniTotali++
			ack_t2[id] <- 1
			fmt.Printf("[SERVER]: Entrato T2. Stato: [T1:%d T2:%d]\n", n1, n2)

		/* SEGNALE TERMINAZIONE */
		case id := <-when(fine, req_entra_t1):
			ack_t1[id] <- -1
		case id := <-when(fine, req_entra_t2):
			ack_t2[id] <- -1
		
		// se ci sono alcuni CLIENTI ad avere un funzionamento ciclico la cui terminazione dipende da altri:
        /* SEGNALAZIONE FINE 
        case <- segnalaFine:
            fine = true   
        */
		
		/* TERMINAZIONE SERVER */
		case <-termina:
			fmt.Println("[SERVER]: Spegnimento sistema.")
			done <- true
			return
		}

		// Verifica condizione di uscita (Area deve essere vuota)
		if contatoreOperazioniTotali >= TOT_OPERAZIONI && n1 == 0 && n2 == 0 {
			fine = true
			fmt.Println("[SERVER]: Limite operazioni raggiunto, chiusura in corso...")
		}
	}
}

/*
 * Funzione main
 */
func main() {
	rand.Seed(time.Now().UnixNano())

	// Inizializzazione array di canali ACK
	for i := 0; i < NUM_TIPO_A; i++ {
		ack_t1[i] = make(chan int)
	}
	for i := 0; i < NUM_TIPO_B; i++ {
		ack_t2[i] = make(chan int)
	}

	/* Avvio Goroutines */
	go server()

	for i := 0; i < NUM_TIPO_A; i++ {
		go processoTIPO_A(i)
	}
	for i := 0; i < NUM_TIPO_B; i++ {
		go processoTIPO_B(i)
	}

	// Attesa terminazione processi client
	for i := 0; i < (NUM_TIPO_A + NUM_TIPO_B); i++ {
		<-done
	}

    // dopo il done di tutti i CLIENTI NON CICLICI, segnalo la fine al server
    //	segnalaFine <- true
    //  for i := 0; i < CLIENTI_CICLICI; i++ {
    //      <-done
    //   }

	// Terminazione server
	termina <- true
	<-done

	fmt.Println("[MAIN]: Programma terminato correttamente.")
}