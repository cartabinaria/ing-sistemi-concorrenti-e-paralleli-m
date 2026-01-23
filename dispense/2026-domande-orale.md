# Domande orale Sistemi Concorrenti e Paralleli

Queste domande sono state raccolte dagli studenti nell A.A 2025/2026 (e precedenti) raccolte in questo [documento condiviso](https://docs.google.com/document/d/1aNpplYRCrK1C7B0uB4uFsarrndah8kYi91ua7rkbuMo/edit?tab=t.0).

## Von Neumann e Parallelizzazione Low Level
* Bottleneck di von neumann e soluzioni (tassonomia flynn e low level parallelism)
* Modello von neumann esteso con enfasi sulla pipeline della cpu (quando avviene il flush)

## Modello a Memoria Comune + Semafori
* Definizione formale del semaforo e proprietà di invarianza + dimostrazioni mutua esclusione
* Dimostrazioni sulla correttezza della soluzione con semafori per la mutua esclusione tramite relazione d'invarianza
* Definizione delle primitive P e V
* Semaforo evento (vincolo di precedenza) e relativa dimostrazione
* Barriera: cos'è, a cosa serve e sua implementazione in pseudocodice
* Implementazione di un semaforo nel kernel di un sistema monoprocessore
* Implementazione dei semafori nei modelli SMP e a nuclei distinti con meccanismo di segnalazione tra i nuclei in caso di context switch

## Modello a Scambio di Messaggi
* Canali di comunicazione: sincrono, asincrono, simmetrico, asimmetrico (definizioni, vantaggi e svantaggi)
* Concetti di link, port, mailbox
* Primitive send e receive: come costruire una send sincrona mediante una asincrona e viceversa
* Send sincrone vs asincrone (vantaggi e svantaggi)
* Receive bloccante vs non bloccante (attesa attiva) e canali molti-a-uno
* Comandi con guardia: definizione, motivi di utilizzo, possibili stati della guardia
* Costrutti alternativo e ripetitivo con sintassi (riferimento linguaggio Ada)
* Determinismo: perché il comando con guardia alternativo non è deterministico
* Sincronizzazione estesa: RPC vs Rendez-vous (differenze, pro e contro, comportamenti del server)

## Algoritmi di Sincronizzazione Distribuiti
* Mutua esclusione in sistemi distribuiti: soluzioni centralizzate e decentralizzate basate su processi
* Pro e contro di un approccio decentralizzato rispetto a uno centralizzato
* Gestione del tempo in sistemi distribuiti: algoritmo di Lamport (orologio logico)
* Algoritmi di sincronizzazione distribuiti: algoritmo Ricart-Agrawala e token ring
* Algoritmi di elezione: Bully e algoritmo ad anello per la scelta del coordinatore

## Virtualizzazione
* VMM di sistema: funzionamento dettagliato, gestione trap e architettura
* Problemi nella realizzazione del VMM: ring deprivileging, aliasing, compression e loro risoluzioni
* Architetture virtualizzabili vs non virtualizzabili
* Fast Binary Translation vs Paravirtualizzazione (differenze, efficienza, pro e contro)
* Xen: architettura generale, Hypervisor, domain, gestione della memoria e balloon process
* Xen: paginazione, gestione delle interruzioni e dei driver
* Live migration in Xen: migrazione di VM tra nodi (precopy e postcopy con costi/benefici)

## Kernel di Sistemi Concorrenti
* Kernel in sistemi multiprocessore: modello SMP (Symmetric Multi-Processing) e modello a Nuclei Distinti
* Differenze tra i due modelli con schemi grafici
* Realizzazione dei semafori in entrambi i modelli con operazioni P e V su semafori condivisi
* Meccanismo di segnalazione tra i nuclei in caso di context switch
* Semafori privati vs condivisi nel modello a nuclei distinti con esempi di interazione
* Limiti del modello SMP rispetto al modello a nuclei distinti

## HPC e Parallelizzazione
* Metriche per valutare le prestazioni di applicazioni parallele: speedup e efficienza
* Strong scaling vs weak scaling
* Limiti del modello Von Neumann e come superarli a livello architetturale
* Evoluzione dal modello di Von Neumann ai sistemi distribuiti e HPC
* Sistemi operativi HPC: LWK, FWK, ibrido
* Concetto di rank nelle primitive send/receive in HPC
* Vantaggi di operazioni gather/scatter rispetto a multiple send da/verso più nodi
* Come è implementato il broadcast in MPI e sue prestazioni
* Architettura del Cineca (ibrido, quindi memoria condivisa e scambio di messaggi)
* OpenMP e MPI: cosa userei se dovessi usare più nodi (entrambi)

## Attività Progettuale
* Modalità, implementazione e test della soluzione progettuale
* Utilizzo di Slurm per la gestione dei job

