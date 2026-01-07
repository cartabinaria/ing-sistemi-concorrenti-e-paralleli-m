// Davide Chirichella - 0001222371

package main

import (
    "fmt"
    "math/rand"
    "time"
)

/* Costanti */
const MAX_serverA = 5

const NUM_UTENTI = 10       

const TIPO_1 = 0
const TIPO_2 = 1

/* Canali Globali */
// Canali di richiesta ServerA (Entrata per tipo e Uscita)
var req_entra_A_T1 = make(chan int, 100)
var req_entra_A_T2 = make(chan int, 100)
var req_uscita_A = make(chan int)

// Canali di richiesta ServerB (Entrata per tipo e Uscita)
var req_entra_B_T1 = make(chan int, 100)
var req_entra_B_T2 = make(chan int, 100)
var req_uscita_B = make(chan int)

// Canali di ACK
var ack_utente [NUM_UTENTI]chan int

// Canali di sincronizzazione terminazione
var done = make(chan bool)
var termina_serverA = make(chan bool)
var termina_serverB = make(chan bool)

/* 
 * Utils 
 */
func printTipo(val int) string {
	switch val {
	case TIPO_1:
		return "TIPO_1"
	case TIPO_2:
		return "TIPO_2"
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
 * Processo Utente
 */
func processoUtente(id int, tipo_iniziale int) {
    fmt.Printf("[Utente %d]: Avviato con tipo %d\n", id, tipo_iniziale)

    /* serverA */ 
    // entra in base al suo tipo
    if tipo_iniziale == TIPO_1 {
        req_entra_A_T1 <- id
    } else {
        req_entra_A_T2 <- id
    }   
    <-ack_utente[id]
    
    // è entrato in serverA, esegue attività
    time.Sleep(time.Duration(rand.Intn(3)+1) * time.Second)
    fmt.Printf("[Utente %d]: Ha finito nel Server A\n", id)

    // esce da serverA
    req_uscita_A <- id
    <-ack_utente[id]

    /* serverB */
    // definisce un nuovo tipo o mantiene il precedente
    var tipo_B = rand.Intn(2)
    if tipo_B == TIPO_1 {
        req_entra_B_T1 <- id
    } else {
        req_entra_B_T2 <- id
    }
    <-ack_utente[id]

    // è entrato in serverB, esegue attività
    time.Sleep(time.Duration(rand.Intn(3)+1) * time.Second)

    // esce da serverB
    req_uscita_B <- id
    <-ack_utente[id]
     
    fmt.Printf("[Utente %d]: Terminato percorso completo\n", id)

    done <- true
}


/*
 * server serverA
 */
func serverA() {

    // Stato interno
    var n_t1 = 0 
    var n_t2 = 0 
    var n_tot = 0

    for {
        select {

        /* USCITA */
        case id := <-req_uscita_A:
            n_tot--
            ack_utente[id] <- 1

        /* ENTRATA */
        // tipo_1
        case id := <-when( n_tot < MAX_serverA && ( n_t1 <= n_t2 || len(req_entra_A_T2) == 0 ) , req_entra_A_T1):
            n_t1++
            n_tot++
            ack_utente[id] <- 1
            fmt.Printf("[ServerA]: Entrato T1 %d. Stato: [T1:%d T2:%d]\n", id, n_t1, n_t2)

        // tipo_2
        case id := <-when( n_tot < MAX_serverA && ( n_t2 <= n_t1 || len(req_entra_A_T1) == 0 ) , req_entra_A_T2):
            n_t2++
            n_tot++
            ack_utente[id] <- 1
            fmt.Printf("[ServerA]: Entrato T2 %d. Stato: [T1:%d T2:%d]\n", id, n_t1, n_t2)

        // se ci sono alcuni CLIENTI ad avere un funzionamento ciclico la cui terminazione dipende da altri:
        /* SEGNALAZIONE FINE 
        case <- segnalaFine:
            fine = true   
        */
            
        /* TERMINAZIONE */
        case <-termina_serverA:
            fmt.Println("[ServerA]: Spegnimento sistema.")
            done <- true
            return
        }
    }
}

/*
 * server serverB
 */
func serverB() {

    // Stato interno
    var occupato = false

    for {
        select {

        /* USCITA */
        case id := <-req_uscita_B:
            occupato = false
            ack_utente[id] <- 1

        /* ENTRATA */
        // tipo_1
        case id := <-when( !occupato && len(req_entra_B_T2) == 0 , req_entra_B_T1):
            occupato = true
            ack_utente[id] <- 1
            fmt.Printf("[ServerB]: Entrato utente T1 %d\n", id)

        // tipo_2
        case id := <-when( !occupato, req_entra_B_T2):
            occupato = true
            ack_utente[id] <- 1
            fmt.Printf("[ServerB]: Entrato utente T2 %d\n", id)
        
        // se ci sono alcuni CLIENTI ad avere un funzionamento ciclico la cui terminazione dipende da altri:
        /* SEGNALAZIONE FINE 
        case <- segnalaFine:
            fine = true   
        */

        /* TERMINAZIONE */
        case <-termina_serverB:
            fmt.Println("[ServerB]: Spegnimento sistema.")
            done <- true
            return
        }
    }
}


/*
 * Funzione main
 */
func main() {
    rand.Seed(time.Now().UnixNano())

    // Inizializzazione array di canali ACK
    for i := 0; i < NUM_UTENTI; i++ {
        ack_utente[i] = make(chan int)
    }

    /* Avvio Goroutines */
    go serverA()
    go serverB()

    for i := 0; i < NUM_UTENTI; i++ {
        var tipo = rand.Intn(2)
        go processoUtente(i, tipo)
    }

    // Attesa terminazione processi utenti
    for i := 0; i < NUM_UTENTI; i++ {
        <-done
    }

    // dopo il done di tutti i CLIENTI NON CICLICI, segnalo la fine al server
    //	segnalaFine <- true
    //  for i := 0; i < CLIENTI_CICLICI; i++ {
    //      <-done
    //   }

    // Terminazione server A e B
    termina_serverA <- true
    <-done
    termina_serverB <- true
    <-done

    fmt.Println("[MAIN]: Programma terminato correttamente.")
}