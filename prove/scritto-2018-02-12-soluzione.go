// Davide Chirichella - 0001222371

package main

import (
    "fmt"
    "math/rand"
    "time"
)

/* Costanti */
const MAX_S = 20              
const MAX_P = 20              

const NUM_AL = 4
const NUM_OO = 2

/* Strutture Dati */
type Richiesta struct {
    id     int
    qty	   int 
}	

/* Canali Globali */
// Canali di richiesta separati per tipo
var req_entrataAL_small = make(chan Richiesta, 100)
var req_entrataAL_big = make(chan Richiesta, 100)
var req_entrataOO_small = make(chan Richiesta, 100)
var req_entrataOO_big = make(chan Richiesta, 100)

// Canali di ACK 
var ack_al [NUM_AL]chan int
var ack_oo [NUM_OO]chan int

// Canali di sincronizzazione terminazione
var done    = make(chan bool)
var termina = make(chan bool)

/*
 * Guardia logica
 */
func when(cond bool, ch chan int) chan int {
    if !cond {
        return nil
    }
    return ch
}

func whenR(cond bool, ch chan Richiesta) chan Richiesta {
    if !cond {
        return nil
    }
    return ch
}

/*
 * Processo client AddettoAL
 */
func processoAddettoAL(id int, M int) {
	fmt.Printf("[AddettoAL %d]: Avviato per M %d\n", id, M)

	// entra
	if ( M > MAX_P / 2 ) {
		req_entrataAL_big   <- Richiesta{id, M}
		<- ack_al[id]
	} else {
		req_entrataAL_small <- Richiesta{id, M}
		<- ack_al[id]
	}

	//fmt.Printf("[AddettoAL %d]: Operazione completata con successo\n", id)
	
	// esce
	done <- true
	return
}

/*
 * Processo client OperatoreOO
 */
func processoOperatoreOO(id int, N int) {
    fmt.Printf("[OperatoreOO %d]: Avviato per N %d\n", id, N)

	// entra
	if ( N > MAX_S / 2 ) {
		req_entrataOO_big   <- Richiesta{id, N}
		<-ack_oo[id]
	} else {
		req_entrataOO_small <- Richiesta{id, N}
		<-ack_oo[id]
	}

	//fmt.Printf("[OperatoreOO %d]: Finito. Esco.\n", id)
	
	// esce
	done <- true
	return
}

/*
 * Deposito centralizzato
 */
func deposito() {
    // Stato interno

    var lenzuola_sporche = MAX_S / 2
    var lenzuola_pulite = MAX_P / 2

    for {
        select {
        
        /* entrata OO */
		// sopra la soglia
        case req := <-whenR( lenzuola_pulite > lenzuola_sporche || ( len(req_entrataAL_big) == 0 && len(req_entrataAL_small) == 0 ) , req_entrataOO_big):
			if ( lenzuola_sporche + req.qty < MAX_S && lenzuola_pulite - req.qty > 0 && len(req_entrataOO_small) == 0 ) {
				
				// restituisco e ritiro
				lenzuola_sporche += req.qty
				lenzuola_pulite  -= req.qty
				
				// finito
				ack_oo[req.id] <- 1
				
				fmt.Printf("[Deposito]: Entrato OO %d, N=%d.\t[S:%d/%d, P:%d/%d]\n", req.id, req.qty, lenzuola_sporche, MAX_S, lenzuola_pulite, MAX_P)
			} else {				
				//fmt.Printf("[Deposito]: NON Entrato OO %d, N=%d.\t[S:%d/%d, P:%d/%d]\n", req.id, req.qty, lenzuola_sporche, MAX_S, lenzuola_pulite, MAX_P)
				req_entrataOO_big <- req
			}
		
		// sotto la soglia (prioritari)
        case req := <-whenR( lenzuola_pulite > lenzuola_sporche || ( len(req_entrataAL_big) == 0 && len(req_entrataAL_small) == 0 ) , req_entrataOO_small):
			if ( lenzuola_sporche + req.qty < MAX_S && lenzuola_pulite - req.qty > 0 ) {
				
				// restituisco e ritiro
				lenzuola_sporche += req.qty
				lenzuola_pulite  -= req.qty
				
				// finito
				ack_oo[req.id] <- 1
				
				fmt.Printf("[Deposito]: Entrato OO %d, N=%d.\t[S:%d/%d, P:%d/%d]\n", req.id, req.qty, lenzuola_sporche, MAX_S, lenzuola_pulite, MAX_P)
			} else {				
				//fmt.Printf("[Deposito]: NON Entrato OO %d, N=%d.\t[S:%d/%d, P:%d/%d]\n", req.id, req.qty, lenzuola_sporche, MAX_S, lenzuola_pulite, MAX_P)

				req_entrataOO_small <- req
			}
		
        /* entrata AL */
		// sopra la soglia
        case req := <-whenR( lenzuola_pulite <= lenzuola_sporche || ( len(req_entrataOO_big) == 0 && len(req_entrataOO_small) == 0 ) , req_entrataAL_big):
			if ( lenzuola_pulite + req.qty < MAX_P && lenzuola_sporche - req.qty > 0 && len(req_entrataAL_small) == 0 ) {
				
				// restituisco e ritiro
				lenzuola_sporche -= req.qty
				lenzuola_pulite  += req.qty
				
				// finito
				ack_al[req.id] <- 1
				
				fmt.Printf("[Deposito]: Entrato AL %d, N=%d.\t[S:%d/%d, P:%d/%d]\n", req.id, req.qty, lenzuola_sporche, MAX_S, lenzuola_pulite, MAX_P)
			} else {				
				//fmt.Printf("[Deposito]: NON Entrato OO %d, N=%d.\t[S:%d/%d, P:%d/%d]\n", req.id, req.qty, lenzuola_sporche, MAX_S, lenzuola_pulite, MAX_P)

				req_entrataAL_big <- req
			}
		
		// sotto la soglia (prioritari)
        case req := <-whenR( lenzuola_pulite <= lenzuola_sporche || ( len(req_entrataOO_big) == 0 && len(req_entrataOO_small) == 0 ) , req_entrataAL_small):
			if ( lenzuola_pulite + req.qty < MAX_P && lenzuola_sporche - req.qty > 0 ) {
				
				// restituisco e ritiro
				lenzuola_sporche -= req.qty
				lenzuola_pulite  += req.qty
				
				// finito
				ack_al[req.id] <- 1
				
				fmt.Printf("[Deposito]: Entrato AL %d, N=%d.\t[S:%d/%d, P:%d/%d]\n", req.id, req.qty, lenzuola_sporche, MAX_S, lenzuola_pulite, MAX_P)
			} else {				
				//fmt.Printf("[Deposito]: NON Entrato OO %d, N=%d.\t[S:%d/%d, P:%d/%d]\n", req.id, req.qty, lenzuola_sporche, MAX_S, lenzuola_pulite, MAX_P)

				req_entrataAL_small <- req
			}

        // Chiusura definitiva del Deposito
        case <-termina:
            fmt.Println("[Deposito]: Spegnimento sistema.")
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
    for i := 0; i < NUM_AL; i++ {
        ack_al[i] = make(chan int)
    }
    for i := 0; i < NUM_OO; i++ {
        ack_oo[i] = make(chan int)
    }

    /* Avvio Goroutines */
    go deposito()

    // Lancio entità
    for i := 0; i < NUM_AL; i++ {
		var N = rand.Intn(7)+1
		go processoAddettoAL(i, N)
	}
    for i := 0; i < NUM_OO; i++ {
		var M = rand.Intn(7)+1
		go processoOperatoreOO(i, M)
	}

    // Attesa terminazione processi client
    for i := 0; i < (NUM_AL + NUM_OO); i++ {
        <-done
    }

    // Terminazione Deposito
    termina <- true
    <-done

    fmt.Println("[MAIN]: Programma terminato correttamente.")
}