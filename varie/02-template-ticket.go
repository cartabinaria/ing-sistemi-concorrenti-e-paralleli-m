// Davide Chirichella - 0001222371

package main

import (
    "fmt"
    "math/rand"
    "time"
)

/* Costanti */
const MAX_RICHIESTE = 10  
const MAX_DEPOSITO = 5    
const NUM_PRODUTTORI = 3  
const NUM_CONSUMATORI = 5 

/* Classi */
const TIPO_A = 0
const TIPO_B = 0

/* Strutture Dati */
type Messaggio struct {
    id     int
    ticket int 
}

/* Canali */
// Canali di richiesta 
var req_produzione = make(chan int, 100)      
var req_fase1      = make(chan int, 100)      
var req_fase2      = make(chan Messaggio, 100) 

// Canali di ACK
var ack_prod [NUM_PRODUTTORI]chan bool
var ack_cons_b [NUM_CONSUMATORI]chan bool
var ack_cons_i [NUM_CONSUMATORI]chan int

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
 * Guardie Logiche
 */
func when(cond bool, ch chan int) chan int {
    if !cond { 
        return nil 
    }
    return ch
}

func whenM(cond bool, ch chan Messaggio) chan Messaggio {
    if !cond { 
        return nil 
    }
    return ch
}

/*
 * Processo Produttore
 */
func processoProduttore(id int) {
    fmt.Printf("[PRODUTTORE %d]: Avviato\n", id)
    
    for {
        time.Sleep(time.Duration(rand.Intn(2)+1) * time.Second)

        req_produzione <- id
        esito := <-ack_prod[id]

        if !esito { 
            fmt.Printf("[PRODUTTORE %d]: Fine.\n", id)
            done <- true
            return
        }
        fmt.Printf("[PRODUTTORE %d]: Consegna effettuata.\n", id)
    }
}

/*
 * Processo Consumatore
 */
func processoConsumatore(id int) {
    fmt.Printf("[CONSUMATORE %d]: Avviato\n", id)

    // Fase 1: Richiesta Ticket
    req_fase1 <- id
    ticket_ricevuto := <-ack_cons_i[id]
    
    if ticket_ricevuto == -1 {
        fmt.Printf("[CONSUMATORE %d]: Richiesta respinta.\n", id)
        done <- true
        return
    }

    time.Sleep(time.Duration(rand.Intn(2)+1) * time.Second)
    
    // Fase 2: Utilizzo Ticket
    req_fase2 <- Messaggio{id, ticket_ricevuto}
    successo := <-ack_cons_b[id]

    if !successo {
        fmt.Printf("[CONSUMATORE %d]: Ticket non valido.\n", id)
    } else {
        fmt.Printf("[CONSUMATORE %d]: Operazione conclusa.\n", id)
    }

    done <- true
}

/*
 * Server
 */
func server() {
    var contatore_fase1 = 0
    var disp = 0
    var prog_ticket = 1
    var ticket_validi [100]bool 
    
    var completati = 0
    var fine = false

    for {
        select {
        
        /* EMISSIONE TICKET */
        case id := <-when(!fine, req_fase1):
            risposta := -1
            if contatore_fase1 < MAX_RICHIESTE {
                risposta = prog_ticket
                ticket_validi[prog_ticket] = true
                prog_ticket++
                contatore_fase1++
            }
            ack_cons_i[id] <- risposta

        /* VALIDAZIONE E CONSUMO */
        case req := <-whenM(!fine && disp > 0 && len(req_fase1) == 0, req_fase2):
            if ticket_validi[req.ticket] {
                ticket_validi[req.ticket] = false 
                disp--
                completati++
                ack_cons_b[req.id] <- true
            } else {
                ack_cons_b[req.id] <- false
            }

        /* PRODUZIONE */
        case id := <-when(!fine && disp < MAX_DEPOSITO && len(req_fase1) == 0 && len(req_fase2) == 0, req_produzione):
            disp++
            ack_prod[id] <- true

        /* TERMINAZIONE DI TUTTI QUELLI IN ATTESA */
        case id := <-when(fine, req_produzione): 
            ack_prod[id] <- false
        case id := <-when(fine, req_fase1):    
            ack_cons_i[id] <- -1
        case req := <-whenM(fine, req_fase2):  
            ack_cons_b[req.id] <- false
        
        // se ci sono alcuni CLIENTI ad avere un funzionamento ciclico la cui terminazione dipende da altri:
        /* SEGNALAZIONE FINE 
        case <- segnalaFine:
            fine = true   
        */

        /* Segnale terminazione server */
        case <-termina:
            done <- true
            return
        }

        // Condizione di fine
        if completati >= NUM_CONSUMATORI {
            fine = true
        }
    }
}

/*
 * Main
 */
func main() {
    rand.Seed(time.Now().UnixNano())
    for i := 0; i < NUM_PRODUTTORI; i++ { 
        ack_prod[i] = make(chan bool) 
    }
    
    for i := 0; i < NUM_CONSUMATORI; i++ { 
        ack_cons_b[i] = make(chan bool)
        ack_cons_i[i] = make(chan int) 
    }

    go server()

    for i := 0; i < NUM_PRODUTTORI; i++ { 
        go processoProduttore(i) 
    }
    
    for i := 0; i < NUM_CONSUMATORI; i++ { 
        go processoConsumatore(i) 
    }

    for i := 0; i < NUM_CONSUMATORI; i++ { 
        <-done 
    }

    // dopo il done di tutti i CLIENTI NON CICLICI, segnalo la fine al server
    //	segnalaFine <- true
    //  for i := 0; i < CLIENTI_CICLICI; i++ {
    //      <-done
    //   }

    termina <- true
    <-done
    fmt.Println("[MAIN]: Chiusura programma.")
}
