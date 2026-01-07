// Cognome Nome - Matricola

package main

import (
    "fmt"
    "math/rand"
    "time"
)

/* Costanti */
const N = 5

const NUM_PAZIENTI = 10       

const BAMBINO = 0
const ADULTO = 1

const ROSSO = 0
const GIALLO = 1
const VERDE = 2
const BIANCO = 3

/* Canali Globali */
// Canali di richiesta Triage
var req_entra_triage = make(chan int)
var req_uscita_triage = make(chan int)

// Canali di richiesta Ambulatorio

var req_entra_A_amb_ROSSO = make(chan int, 100)
var req_entra_A_amb_GIALLO = make(chan int, 100)
var req_entra_A_amb_VERDE = make(chan int, 100)
var req_entra_A_amb_BIANCO = make(chan int, 100)

var req_entra_B_amb_ROSSO = make(chan int, 100)
var req_entra_B_amb_GIALLO = make(chan int, 100)
var req_entra_B_amb_VERDE = make(chan int, 100)
var req_entra_B_amb_BIANCO = make(chan int, 100)

var req_uscita_amb = make(chan int)

// Canali di ACK
var ack_paziente [NUM_PAZIENTI]chan int

// Canali di sincronizzazione terminazione
var done = make(chan bool)

var termina_triage = make(chan bool)
var termina_ambulatorio_A = make(chan bool)
var termina_ambulatorio_B = make(chan bool)

/*
 * Guardia logica
 */
func when(cond bool, ch chan int) chan int {
    if !cond {
        return nil
    }
    return ch
}

/* * Utils 
 */
func printTipo(val int) string {
    switch val {
    case ROSSO:
        return "ROSSO"
    case GIALLO:
        return "GIALLO"
    case VERDE:
        return "VERDE"
    case BIANCO:
        return "BIANCO"
    default:
        return ""
    }
}

func printEta(val int) string {
    switch val {
    case ADULTO:
        return "ADULTO"
    case BAMBINO:
        return "BAMBINO"
    default:
        return ""
    }
}


/*
 * Processo Paziente
 */
func processoPaziente(id int, eta int) {
    fmt.Printf("[Paziente %d]: Avviato come " + printEta(eta) + "\n", id)

    /* TRIAGE */ 
    // entra in base al suo tipo
    req_entra_triage <- id
    <-ack_paziente[id]
    
    // è entrato, esegue attività
    time.Sleep(time.Duration(rand.Intn(3)+1) * time.Second)
    
    // ha finito, esce
    req_uscita_triage <- id
    var codice = <-ack_paziente[id]

    //fmt.Printf("[Paziente %d]: Ha finito nel triage con codice: %s\n", id, printTipo(codice))

    /* AMBULATORIO */
    if eta == BAMBINO {
        // R
        if codice == ROSSO {
            req_entra_B_amb_ROSSO <- id
        } else if codice == GIALLO { // G
            req_entra_B_amb_GIALLO <- id
        } else if codice == VERDE { // V
            req_entra_B_amb_VERDE <- id
        } else if codice == BIANCO { // B
            req_entra_B_amb_BIANCO <- id
        }
    } else if eta == ADULTO {
        // R
        if codice == ROSSO {
            req_entra_A_amb_ROSSO <- id
        } else if codice == GIALLO { // G
            req_entra_A_amb_GIALLO <- id
        } else if codice == VERDE { // V
            req_entra_A_amb_VERDE <- id
        } else if codice == BIANCO { // B
            req_entra_A_amb_BIANCO <- id
        }
    }

    <-ack_paziente[id]

    // il paziente viene curato
    time.Sleep(time.Duration(rand.Intn(3)+1) * time.Second)

    // esce
    req_uscita_amb <- id
    <-ack_paziente[id]
     
    //fmt.Printf("[Paziente %d]: Terminato\n", id)

    done <- true
}


/*
 * server
 */
func triage() {

    for {
        select {

        /* USCITA */
        case id := <-req_uscita_triage:
            ack_paziente[id] <- 1

        /* ENTRATA */
        // tipo 1
        case id := <- req_entra_triage:
            
            var codice = rand.Intn(4)
            ack_paziente[id] <- codice
            
            fmt.Printf("[Triage]: Entrato Paziente %d, codice %s\n", id, printTipo(codice))

        /* TERMINAZIONE */
        case <-termina_triage:
            fmt.Println("[Triage]: Spegnimento sistema.")
            done <- true
            return
        }
    }
}

/*
 * server 
 */
func ambulatorio(tipo int) {

    if tipo == ADULTO {

        // Stato interno
        var medici_occupati = 0

        for {
            select {

            /* USCITA */
            case id := <-req_uscita_amb:
                medici_occupati --
                ack_paziente[id] <- 1

            /* ENTRATA */
            // adulto R
            case id := <-when( medici_occupati < N , req_entra_A_amb_ROSSO):
                medici_occupati ++
                ack_paziente[id] <- 1
                fmt.Printf("[Ambulatorio ADULTI]: Entrato %d.\t[Medici A: %d/%d]\n", id, medici_occupati, N)
            
            // adulto G
            case id := <-when( medici_occupati < N && len(req_entra_A_amb_ROSSO) == 0 , req_entra_A_amb_GIALLO):
                medici_occupati ++
                ack_paziente[id] <- 1
                fmt.Printf("[Ambulatorio ADULTI]: Entrato %d.\t[Medici A: %d/%d]\n", id, medici_occupati, N)
            
            // adulto V
            case id := <-when( medici_occupati < N && len(req_entra_A_amb_ROSSO) == 0 && len(req_entra_A_amb_GIALLO) == 0 , req_entra_A_amb_VERDE):
                medici_occupati ++
                ack_paziente[id] <- 1
                fmt.Printf("[Ambulatorio ADULTI]: Entrato %d.\t[Medici A: %d/%d]\n", id, medici_occupati, N)
            
            // adulto B
            case id := <-when( medici_occupati < N && len(req_entra_A_amb_ROSSO) == 0 && len(req_entra_A_amb_GIALLO) == 0  && len(req_entra_A_amb_VERDE) == 0, req_entra_A_amb_BIANCO):
                medici_occupati ++
                ack_paziente[id] <- 1
                fmt.Printf("[Ambulatorio ADULTI]: Entrato %d.\t[Medici A: %d/%d]\n", id, medici_occupati, N)
            
            /* TERMINAZIONE */
            case <-termina_ambulatorio_A:
                fmt.Println("[Ambulatorio ADULTI]: Spegnimento sistema.")
                done <- true
                return
            }
        }

    } else {

        // Stato interno
        var medici_occupati = 0

        for {
            select {

            /* USCITA */
            case id := <-req_uscita_amb:
                medici_occupati --
                ack_paziente[id] <- 1

            /* ENTRATA */
            // adulto R
            case id := <-when( medici_occupati < N , req_entra_B_amb_ROSSO):
                medici_occupati ++
                ack_paziente[id] <- 1
                fmt.Printf("[Ambulatorio BAMBINI]: Entrato %d.\t[Medici B: %d/%d]\n", id, medici_occupati, N)
            
            // adulto G
            case id := <-when( medici_occupati < N && len(req_entra_B_amb_ROSSO) == 0 , req_entra_B_amb_GIALLO):
                medici_occupati ++
                ack_paziente[id] <- 1
                fmt.Printf("[Ambulatorio BAMBINI]: Entrato %d.\t[Medici B: %d/%d]\n", id, medici_occupati, N)
            
            // adulto V
            case id := <-when( medici_occupati < N && len(req_entra_B_amb_ROSSO) == 0 && len(req_entra_B_amb_GIALLO) == 0 , req_entra_B_amb_VERDE):
                medici_occupati ++
                ack_paziente[id] <- 1
                fmt.Printf("[Ambulatorio BAMBINI]: Entrato %d.\t[Medici B: %d/%d]\n", id, medici_occupati, N)
            
            // adulto B
            case id := <-when( medici_occupati < N && len(req_entra_B_amb_ROSSO) == 0 && len(req_entra_B_amb_GIALLO) == 0  && len(req_entra_B_amb_VERDE) == 0, req_entra_B_amb_BIANCO):
                medici_occupati ++
                ack_paziente[id] <- 1
                fmt.Printf("[Ambulatorio BAMBINI]: Entrato %d.\t[Medici B: %d/%d]\n", id, medici_occupati, N)
            
            /* TERMINAZIONE */
            case <-termina_ambulatorio_B:
                fmt.Println("[Ambulatorio BAMBINI]: Spegnimento sistema.")
                done <- true
                return
            }
        }
    }
}



/*
 * Funzione main
 */
func main() {
    rand.Seed(time.Now().UnixNano())

    // Inizializzazione array di canali ACK
    for i := 0; i < NUM_PAZIENTI; i++ {
        ack_paziente[i] = make(chan int)
    }

    /* Avvio Goroutines */
    go triage()
    go ambulatorio(ADULTO)
    go ambulatorio(BAMBINO)

    for i := 0; i < NUM_PAZIENTI; i++ {
        var eta = rand.Intn(2)
        go processoPaziente(i, eta)
    }

    // Attesa terminazione client
    for i := 0; i < NUM_PAZIENTI; i++ {
        <-done
    }

    // Terminazione server
    termina_triage <- true
    <-done

    termina_ambulatorio_A <- true
    <-done

    termina_ambulatorio_B <- true
    <-done


    fmt.Println("[MAIN]: Programma terminato correttamente.")
}