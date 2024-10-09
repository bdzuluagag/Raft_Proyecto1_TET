package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

const raftTimeout = 10 * time.Second

var (
	mutex    = &sync.Mutex{}
	dbPath   string //db de cada nodo
	raftNode *raft.Raft
)

// FSM implementa la Finite State Machine de Raft
type FSM struct{}

// Apply maneja la replicación de logs (implementaremos esto más adelante)
func (f *FSM) Apply(raftLog *raft.Log) interface{} {
	var data map[string]string
	if err := json.Unmarshal(raftLog.Data, &data); err != nil {
		log.Fatalf("Error al aplicar log: %v", err)
		return nil
	}

	key := data["key"]
	value := data["value"]

	mutex.Lock()
	err := appendToFile(key, value)
	mutex.Unlock()

	if err != nil {
		log.Printf("Error al replicar datos: %v", err)
		return err
	}

	fmt.Printf("Datos replicados: %s = %s\n", key, value)
	return nil
}

// Snapshot maneja los snapshots de Raft
func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	return nil, nil
}

// Restore se usa para restaurar el estado de un snapshot
func (f *FSM) Restore(io.ReadCloser) error {
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: cluster [puerto_raft] [puerto_http]")
		return
	}
	portRaft := os.Args[1] // Puerto para Raft
	portHTTP := os.Args[2] // Puerto para HTTP

	initializeRaft(portRaft)
	setupHTTPServer(portHTTP)

	select {}
}

func setupHTTPServer(port string) {

	http.HandleFunc("/write", writeHandler)
	http.HandleFunc("/read/", readHandler)
	http.HandleFunc("/is_leader", isLeaderHandler) // Nueva ruta para chequear el estado del líder

	fmt.Printf("Nodo corriendo en el puerto %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Error al iniciar el servidor HTTP: %v", err)
	}
}

func initializeRaft(port string) {
	// Configuración por defecto de Raft
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID("node-" + port) // Asignar un ID único basado en el puerto

	// Asignar el archivo de base de datos basado en el puerto
	dbPath = fmt.Sprintf("database_%s.txt", port)

	fmt.Printf("Base de datos para este nodo: %s\n", dbPath)

	// Almacenamiento de logs
	logStore, err := raftboltdb.NewBoltStore(fmt.Sprintf("raft-log-%s.db", port))
	if err != nil {
		log.Fatalf("Error al inicializar logStore: %v", err)
	}

	// Almacenamiento de snapshots
	snapshotStore, err := raft.NewFileSnapshotStore(".", 1, os.Stdout)
	if err != nil {
		log.Fatalf("Error al crear snapshot store: %v", err)
	}

	// Transporte TCP
	address := fmt.Sprintf("localhost:%s", port)
	transport, err := raft.NewTCPTransport(address, nil, 3, raftTimeout, os.Stdout)
	if err != nil {
		log.Fatalf("Error al crear TCP transport: %v", err)
	}

	// Crear FSM
	fsm := &FSM{}

	// Inicializar Raft
	raftNode, err = raft.NewRaft(config, fsm, logStore, logStore, snapshotStore, transport)
	if err != nil {
		log.Fatalf("Error al inicializar Raft: %v", err)
	}

	// Definir los nodos conocidos (nodos en el clúster)
	raftNode.BootstrapCluster(raft.Configuration{
		Servers: []raft.Server{
			{ID: "node-8081", Address: raft.ServerAddress("localhost:8081")},
			{ID: "node-8082", Address: raft.ServerAddress("localhost:8082")},
			{ID: "node-8083", Address: raft.ServerAddress("localhost:8083")},
		},
	})
}

func writeHandler(w http.ResponseWriter, r *http.Request) {

	if raftNode.State() != raft.Leader {
		http.Error(w, "Este nodo no es el líder", http.StatusForbidden)
		return
	}

	var data map[string]string

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Datos inválidos", http.StatusBadRequest)
		return
	}

	// Marshalizar los datos a JSON para replicación
	jsonData, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Error al preparar datos para replicación", http.StatusInternalServerError)
		return
	}

	// Proponer los datos al log de Raft para replicación
	future := raftNode.Apply(jsonData, raftTimeout)
	if err := future.Error(); err != nil {
		http.Error(w, "Error al replicar datos", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Escritura replicada: %s = %s\n", data["key"], data["value"])
	w.WriteHeader(http.StatusOK)
}

func readHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[len("/read/"):]

	value, err := readFromFile(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	response := map[string]string{"key": key, "value": value}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	fmt.Printf("Lectura recibida: %s = %s\n", key, value)
}

// AppendToFile agrega la clave y el valor al archivo
func appendToFile(key, value string) error {
	file, err := os.OpenFile(dbPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("%s=%s\n", key, value))
	return err
}

// ReadFromFile lee el valor correspondiente a la clave del archivo
func readFromFile(key string) (string, error) {
	file, err := os.Open(dbPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var value string
	var found bool

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, key+"=") {
			value = strings.Split(line, "=")[1]
			found = true
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	if !found {
		return "", fmt.Errorf("Registro no encontrado")
	}

	return value, nil
}

func isLeaderHandler(w http.ResponseWriter, r *http.Request) {
	status := map[string]bool{"is_leader": raftNode.State() == raft.Leader}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
