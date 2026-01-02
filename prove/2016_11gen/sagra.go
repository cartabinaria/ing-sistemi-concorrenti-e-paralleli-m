package main

import (
	"fmt"
	"math/rand"
	"time"
)


const MAXBUFF int = 10
const MAXPROC int = 3
const MAX int = 5 // capacità

const CARTA int = 0
const CONTANTI int = 1
const PIADA int = 0
const CRESCIONE int = 1

//definizione canali
var done = make(chan bool)
var termina = make(chan bool)

var pagamento[2] chan int // necessità di accodamento per priorità
var richiesta[2] chan int
var deposito[2] chan bool

var pagamento_ack [MAXPROC]chan int 
var richiesta_ack [MAXPROC] chan int
var deposito_ack = make(chan bool, MAXBUFF)

func when(b bool, c chan int) chan int {
	if !b {
		return nil
	}
	return c
}

func whenBool(b bool, c chan bool) chan bool {
	if !b {
		return nil
	}
	return c
}

func cliente(myid int) {
	var tt int
	tt = rand.Intn(5) + 1
	var tipoPag int = rand.Intn(2)
	var quantità int = rand.Intn(5)+1

	time.Sleep(time.Duration(tt) * time.Second)
	fmt.Printf("[CLIENTE %d] Voglio pagare %d prodotti\n", myid, quantità)
	pagamento[tipoPag] <- myid
	<-pagamento_ack[myid]
 	time.Sleep(time.Duration(tt) * time.Second) //tempo di pagamento
	
	fmt.Printf("[CLIENTE %d] Vado al bancone a ritirare %d prodotti\n", myid, quantità)

	for i := 0; i < quantità; i++ {
		var tipo = rand.Intn(2)
		fmt.Printf("[CLIENTE %d] Voglio ritirare un prodotto tipo %d [0 PIADA, 1 CRESCIONE]\n", myid, tipo)
		richiesta[tipoPag] <- myid
		<-richiesta_ack[myid]
	}
	fmt.Printf("[CLIENTE %d] Ho ritirato tutto, vado a mangiare\n", myid)

    done<-true
}

func cuoco(){
	fmt.Printf("[CUOCO] Sono pronto, inizio a preparare i piatti\n")
	for{
		var tt int
		tt = rand.Intn(5) + 1
		time.Sleep(time.Duration(tt) * time.Second)
		var tipo = rand.Intn(2)
		fmt.Printf("[CUOCO] Preparo un piatto tipo %d [0 PIADA, 1 CRESCIONE]\n", tipo)
		deposito[tipo] <- true
		<-deposito_ack
	}
	done<-true
}


func server(){
	var countCR int = 0
	var countPI int = 0

	for {
		select {
        //CASSIERE
		case x := <- pagamento[CARTA]:
			fmt.Printf("[CASSIERE] Cliente %d paga con carta, rilascio scontrino\n", x)
			pagamento_ack[x] <- 1 // termine "call"
        case x := <-when( len(pagamento[CARTA])==0, pagamento[CONTANTI]):
			fmt.Printf("[CASSIERE] Cliente %d paga in contanti, rilascio scontrino\n", x)
			pagamento_ack[x] <- 1 // termine "call"
		//ADDETTO AL BANCO - CLIENTE
		case x := <- when(countCR < MAX && countCR > 0,richiesta[CRESCIONE]):
			countCR--
			fmt.Printf("[ADDETTO] Servo CRESCIONE al cliente %d\n", x)
			richiesta_ack[x]<-1
		case x := <- when(countPI < MAX && countPI > 0 && len(richiesta[CRESCIONE])==0,richiesta[PIADA]):
			countPI--
			fmt.Printf("[ADDETTO] Servo PIADA al cliente %d\n", x)
			richiesta_ack[x]<-1
		//ADDETTO AL BANCO - CUOCO
		case <- whenBool(countPI+countCR<MAX && countPI<=countCR, deposito[PIADA]):
			countPI++
			fmt.Printf("[ADDETTO] Il cuoco ha preparato una PIADA, rifornisco il bancone (%d CRESCIONE, %d PIADE)\n", countCR, countPI)
			deposito_ack<-true
		case <- whenBool(countPI+countCR<MAX && countCR<=countPI, deposito[CRESCIONE]):
			countCR++
			fmt.Printf("[ADDETTO] Il cuoco ha preparato un CRESCIONE, rifornisco il bancone(%d CRESCIONE, %d PIADE)\n", countCR, countPI)
			deposito_ack<-true
		case <-termina: // quando tutti i processi hanno finito
			fmt.Printf("\n\n\nFINE!!!!!!\n\n\n")
			done <- true
			return
		}
	}
}

func main() {
        
	//inizializzazione canali per le auto a nord e a sud 
	for i := 0; i < MAXPROC; i++ {
		pagamento_ack[i] = make(chan int, MAXBUFF)
		richiesta_ack[i] = make(chan int, MAXBUFF)
	}

    for i := 0; i < 2; i++ {
		pagamento[i] = make(chan int, MAXBUFF)
		richiesta[i] = make(chan int, MAXBUFF)
		deposito[i] = make(chan bool, MAXBUFF)
	}

	rand.Seed(time.Now().Unix())
	go server()
	go cuoco()

	for i := 0; i < MAXPROC; i++ {
		go cliente(i)
	}
	
	
	for i := 0; i < MAXPROC; i++ {
		<-done
	}
	
    termina <- true
	<-done
}


