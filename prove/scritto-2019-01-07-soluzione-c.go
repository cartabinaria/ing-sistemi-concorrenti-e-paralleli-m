/*
	Marco Bertolazzi - Mat. 0000884790
	Esame 07/01/2019 - Tema C
    voto: 30
    Lo carico con la speranza che possa essere di aiuto a tutti!
*/

package main

import (
	"fmt"
	"time"
	"math/rand"
)
func when(b bool, c chan int) chan int {
	if b {
		return c
	}	
	return nil
}


var MAX_BISCOTTI int = 10
var M int = 3
var TOT_GELATI int = 20



var done = make(chan bool)
var termina = make(chan bool)

var deposito_biscotto = make(chan int, 1)
var ack_deposito_biscotto = make(chan int)

var prelievo_biscotti = make(chan int, 1)
var ack_prelievo_biscotti = make(chan int)


//volendo si potrebbe fare anche un solo canale...
var rifornimento_serbatoio = make(chan int, 1)
var ack_rifornimento_serbatoio = make(chan int)




func mb(){

	for i:=0; i<TOT_GELATI*2; i++ {
		
		fmt.Printf("[MB] produzione biscotto\n")
		time.Sleep(time.Duration(rand.Intn(2) + 1) * time.Second)
		
		fmt.Printf("[MB] deposito biscotto\n")
		deposito_biscotto <- 1	//deposito un biscotto
		<-ack_deposito_biscotto
	}
	
	fmt.Printf("[MB] ho finito di lavorare\n")
	done <- true
	return
}//MB()

func mg(){

	var gelati_assemblabili int = 0 //TODO oppure M

	for i:=0; i<TOT_GELATI; i++{
	
		fmt.Printf("[MG] richiedo biscotti\n")
		prelievo_biscotti <- 2 //chiedo due biscotti
		<- ack_prelievo_biscotti 
	
		if gelati_assemblabili == 0 {
			fmt.Printf("[MG] richiedo rifornimento serbatoio\n")
			rifornimento_serbatoio <- 1
			gelati_assemblabili = <- ack_rifornimento_serbatoio	//l'operaio riempe per produrre M gelati
		}
		
		time.Sleep(time.Duration(rand.Intn(5) + 1) * time.Second)	
		gelati_assemblabili--

		fmt.Printf("[MG] deposito gelato numero %d\n", i+1)
	}
	
	fmt.Printf("[MG] ho finito di lavorare\n")
	done <- true
	return
}//MG()


func operaio(){

	for {
		select{
		
		case <- rifornimento_serbatoio : {
		
			fmt.Printf("[operaio] richiesta di rifornimento serbatoio\n")			
			time.Sleep(time.Duration(rand.Intn(5) + 1) * time.Second)				
			ack_rifornimento_serbatoio <- M	//riempiamo il serbatoio per produrre M gelati
		}//rifornimento serbatorio
		
		
		
		case <- termina : {
			fmt.Printf("[operaio] fine\n")
			done <- true
			return	
		}//termina	
		}//select
	}//for
	

}//operaio()

func alimentatore(){

	var numero_biscotti int = 0
	
	for {
		fmt.Printf("[alimentatore] Numbero biscotti: %d\n", numero_biscotti)
		
		select{
		
		case x := <- when(numero_biscotti < MAX_BISCOTTI && (numero_biscotti < MAX_BISCOTTI / 2 || len(prelievo_biscotti) == 0),deposito_biscotto) : {
		
			fmt.Printf("[alimentatore] ricevuto biscotto\n")
			time.Sleep(time.Duration(rand.Intn(3) + 1) * time.Second)						
			numero_biscotti += x	//aggiunta di x biscotti
			ack_deposito_biscotto <- 1
		}//deposito_biscotto
		
		case x := <- when(numero_biscotti >= 2 && (numero_biscotti >= MAX_BISCOTTI / 2 || len(deposito_biscotto) == 0) , prelievo_biscotti) : {
		
			fmt.Printf("[alimentatore] richiesta di due biscotti\n")
			time.Sleep(time.Duration(rand.Intn(3) + 1) * time.Second)						
			numero_biscotti -= x	//mi chiedono x biscotti
			ack_prelievo_biscotti <- 1
		}//prelievo_biscotti
	
		case <- termina : {
			fmt.Printf("[alimentatore] fine\n")
			done <- true
			return	
		}//termina
		}//select
	}//for
}//alimentatore



func main() {

	rand.Seed(time.Now().Unix())

	go alimentatore()
	go operaio()
	
	go mb()
	go mg()
	
	
	//attesa terminazione MB e MG
	<- done
	<- done	
	
	//attesa terminazione alimentatore e operaio
	termina <- true
	<- done

	termina <- true
	<- done


	fmt.Printf("[main] fine!!!!!!\n")
}//main
