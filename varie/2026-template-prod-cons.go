// Davide Chirichella - 0001222371

package main

import (
    "fmt"
    "math/rand"
    "time"
)

/* Costanti */
const MAX_DEPOSITO = 5

const TOT_OPERAZIONI = 10

const NUM_PRODUTTORI = 3
const NUM_CONSUMATORI = 2

// Identificativi classi
const TIPO_A = 0
const TIPO_B = 1

/* Canali Globali */
// Canali di richiesta
var req_produzione = make(chan int, 100)
var req_consumo    = make(chan int, 100)

// Canali di ACK 
var ack_prod [NUM_PRODUTTORI]chan int
var ack_cons [NUM_CONSUMATORI]chan int

// Canali di sincronizzazione terminazione
var done    = make(chan bool)
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
 * Processo client produttore
 */
func processoProduttore(id int) {
    fmt.Printf("[PRODUTTORE %d]: Avviato\n", id)
    for {
        // Simulazione lavoro locale
        time.Sleep(time.Duration(rand.Intn(2)+1) * time.Second)

        // Richiesta al server
        req_produzione <- id
        risposta := <-ack_prod[id]

        if risposta == -1 { // Segnale di terminazione
            fmt.Printf("[PRODUTTORE %d]: Ricevuto segnale di fine. Esco.\n", id)
            done <- true
            return
        }

        fmt.Printf("[PRODUTTORE %d]: Operazione completata con successo\n", id)
    }
}

/*
 * Processo client consumatore
 */
func processoConsumatore(id int) {
    fmt.Printf("[CONSUMATORE %d]: Avviato\n", id)
    for {
        req_consumo <- id
        risposta := <-ack_cons[id]

        if risposta == -1 {
            fmt.Printf("[CONSUMATORE %d]: Ricevuto segnale di fine. Esco.\n", id)
            done <- true
            return
        }

        // Simulazione utilizzo risorsa
        time.Sleep(time.Duration(rand.Intn(2)+1) * time.Second)
        fmt.Printf("[CONSUMATORE %d]: Risorsa elaborata\n", id)
    }
}

/*
 * Server centralizzato
 */
func server() {
    // Stato interno
    var risorseDisponibili = 0
    var contatoreTotale = 0
    var fine = false

    for {
        select {
		
        /* PRODUZIONE */
        case id := <-when(!fine && risorseDisponibili < MAX_DEPOSITO, req_produzione):
            risorseDisponibili++
            ack_prod[id] <- 1
            fmt.Printf("[SERVER]: Prodotto da %d. Disponibili: %d\n", id, risorseDisponibili)

        /* CONSUMO */ 
        case id := <-when(!fine && risorseDisponibili > 0, req_consumo):
            risorseDisponibili--
            contatoreTotale++
            ack_cons[id] <- 1
            fmt.Printf("[SERVER]: Consumato da %d. Totale completati: %d\n", id, contatoreTotale)

        /* GESTIONE TERMINAZIONE */
        case id := <-when(fine, req_produzione):
            ack_prod[id] <- -1
        case id := <-when(fine, req_consumo):
            ack_cons[id] <- -1
        
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

		// Ulteriori controlli dopo ogni operazione
		// che verranno eseguiti indipendentemente dal tipo di operazione

        // Verifica condizione di uscita
        if contatoreTotale >= TOT_OPERAZIONI {
            fine = true
            fmt.Println("[SERVER]: Soglia raggiunta, avvio procedura di terminazione...")
        }
    }
}

/*
 * Funzione main
 */
func main() {
    rand.Seed(time.Now().UnixNano())

    // Inizializzazione array di canali ACK
    for i := 0; i < NUM_PRODUTTORI; i++ {
        ack_prod[i] = make(chan int)
    }
    for i := 0; i < NUM_CONSUMATORI; i++ {
        ack_cons[i] = make(chan int)
    }

    /* Avvio Goroutines */
    go server()

    for i := 0; i < NUM_PRODUTTORI; i++ {
        go processoProduttore(i)
    }
    for i := 0; i < NUM_CONSUMATORI; i++ {
        go processoConsumatore(i)
    }

    // Attesa terminazione processi client
    for i := 0; i < (NUM_PRODUTTORI + NUM_CONSUMATORI); i++ {
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
