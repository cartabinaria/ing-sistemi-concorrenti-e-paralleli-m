package main

import(
	"fmt"
	"time"
	"math/rand"
)

//definizione tipo
type nuovo_tipo struct{
	id int
	dato int
}

//costanti
const N_VIAGGIATORI int = 30
const MAX_MOLO int = 3
const MAX_TEMPO_CHIUSURA_MOLO int = 3
const MAX_ATTESA int = 4
const MAX_BUFF int = 10
const IN int = 0
const OUT int = 1

//canali
var done = make(chan bool)
var termina = make(chan bool)
var termina_addetto = make(chan bool)
var viaggiatore_in_ingresso = make(chan int, MAX_BUFF)
var viaggiatore_out_ingresso = make(chan int, MAX_BUFF)
var viaggiatore_in_uscita = make(chan int, MAX_BUFF)
var viaggiatore_out_uscita = make(chan int, MAX_BUFF)
var chiusura_molo = make(chan bool)
var apertura_molo = make(chan bool)

//ack
var ack_viaggiatore_in_ingresso[N_VIAGGIATORI] chan bool
var ack_viaggiatore_out_ingresso[N_VIAGGIATORI] chan bool
var ack_viaggiatore_in_uscita[N_VIAGGIATORI] chan bool
var ack_viaggiatore_out_uscita[N_VIAGGIATORI] chan bool
var ack_chiusura_molo = make(chan bool)
var ack_apertura_molo = make(chan bool)

//funzioni utility
func when(b bool, c chan int) chan int{
	if(!b){
		return nil
	}
	return c
}

func when_bool(b bool, c chan bool) chan bool{
	if(!b){
		return nil
	}
	return c
}

//ridefinire una when per ogni nuovo tipo definito (se occorre)

func viaggiatore(id int){
	
	var direzione int
	
	//scelgo se arrivare o partire
	direzione = rand.Intn(2)
	time.Sleep(time.Duration(rand.Intn(MAX_ATTESA)+1) * time.Second)
	
	switch direzione{
	
		case IN:
			viaggiatore_in_ingresso <- id
			fmt.Printf("[Viaggiatore %d] scelta direzione IN\n", id)
			<- ack_viaggiatore_in_ingresso[id]
			
			fmt.Printf("[Viaggiatore %d] in transito sul ponte\n", id)
			
			viaggiatore_in_uscita <- id
			<- ack_viaggiatore_in_uscita[id]		
			
		case OUT:
			viaggiatore_out_ingresso <- id
			fmt.Printf("[Viaggiatore %d] scelta direzione OUT\n", id)
			<- ack_viaggiatore_out_ingresso[id]
			
			fmt.Printf("[Viaggiatore %d] in transito sul ponte\n", id)
			
			viaggiatore_out_uscita <- id
			<- ack_viaggiatore_out_uscita[id]
	}
	
	done <- true
}

func addetto(){

	for{
		select{
			case <-termina_addetto:
				fmt.Printf("[Addetto] Finito il lavoro! Addio!\n")
				done <- true
			
			default:
				//attendo un tempo random
				time.Sleep(time.Duration(rand.Intn(MAX_ATTESA)+3) * time.Second)
				fmt.Printf("[Addetto] Tempo di chiusura!\n")
				//chiudo il molo
				chiusura_molo <- true
				<- ack_chiusura_molo
				
				time.Sleep(time.Duration(rand.Intn(MAX_TEMPO_CHIUSURA_MOLO)+1) * time.Second)
				
				fmt.Printf("[Addetto] Tempo di apertura!\n")
				apertura_molo <-true
				<- ack_apertura_molo
		}
	}
}

func passerella(){
	
	var viaggiatori_sul_ponte int = 0
	var molo_aperto bool = true
	
	//ciclo di vita del server
	for{
		select{
			//se arriva un viaggiore IN e c'è spazio sul molo, non ci sono OUT in attesa ed il molo è aperto! passa!
			case x:=<-when(((viaggiatori_sul_ponte < MAX_MOLO) && molo_aperto && (len(viaggiatore_out_ingresso) == 0)), viaggiatore_in_ingresso):
				fmt.Printf("[Passerella] viaggiatore %d entrato in direzione IN\n", x)
				viaggiatori_sul_ponte++
				ack_viaggiatore_in_ingresso[x] <- true
				
			case x:=<-viaggiatore_in_uscita:
				fmt.Printf("[Passerella] viaggiatore %d in uscita in direzione IN\n",x)
				viaggiatori_sul_ponte--
				ack_viaggiatore_in_uscita[x] <- true
			
			//se arriva un viaggiatore OUT e c'è spazio sul molo passa!
			case x:=<-when(((viaggiatori_sul_ponte < MAX_MOLO) && molo_aperto), viaggiatore_out_ingresso):
				fmt.Printf("[Passerella] viaggiatore %d entrato in direzione OUT\n", x)	
				viaggiatori_sul_ponte++
				ack_viaggiatore_out_ingresso[x] <- true
							
			case x:= <-viaggiatore_out_uscita:
				fmt.Printf("[Passerella] viaggiatore %d in uscita in direzione OUT\n",x)
				viaggiatori_sul_ponte--
				ack_viaggiatore_out_uscita[x] <- true
			
			//se il molo viene chiuso
			case <- when_bool(molo_aperto, chiusura_molo):
				fmt.Printf("[Passerella] ricevuto ordine di chiusura...\n")
				molo_aperto = false
				ack_chiusura_molo <- true
				
			case <- when_bool(!molo_aperto, apertura_molo):
				fmt.Printf("[Passerella] ricevuto ordine di apertura...\n")
				molo_aperto = true
				ack_apertura_molo <- true
				
			case <-termina:
				fmt.Printf("[Passerella] termino\n")
				done <- true
				return
		}
		fmt.Printf("[Passerella] molo aperto: %t, viaggiatori sul molo: %d, viaggiatori IN in attesa:%d, viaggiatori OUT in attesa: %d\n", molo_aperto, viaggiatori_sul_ponte, len(viaggiatore_in_ingresso), len(viaggiatore_out_ingresso))
	}
}

func main(){
	
	fmt.Printf("Programma avviato\n")
	rand.Seed(time.Now().Unix())
	
	//inizializzo canali ack
	for i:=0; i<N_VIAGGIATORI;i++{
		ack_viaggiatore_in_ingresso[i] = make(chan bool, MAX_BUFF)
		ack_viaggiatore_out_ingresso[i] = make(chan bool, MAX_BUFF)
		ack_viaggiatore_in_uscita[i] = make(chan bool, MAX_BUFF)
		ack_viaggiatore_out_uscita[i] = make(chan bool, MAX_BUFF)
	}	
	
	//lancio threads
	for i:=0; i<N_VIAGGIATORI;i++{
		go viaggiatore(i)
	}

	go addetto()
	
	//lancio il server
	go passerella()
	
	//attendo la terminazione dei clients
	for i:=0; i < N_VIAGGIATORI; i++{
		<-done
	}
	
	termina_addetto <- true
	<- done
	
	//avviso la passerella
	termina <- true
	
	//attendo la terminazione della passerella
	<-done
	
	fmt.Printf("Programma terminato\n")
}
