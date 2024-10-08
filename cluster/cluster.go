package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
)

var (
	mutex    = &sync.Mutex{}
	isLeader = false // Variable que define si este nodo es líder
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Uso: cluster [puerto] [role: leader/follower]")
		return
	}
	port := os.Args[1]
	role := os.Args[2]

	// Definir si es líder o follower
	if role == "leader" {
		isLeader = true
	}

	http.HandleFunc("/write", writeHandler)
	http.HandleFunc("/read/", readHandler)

	fmt.Printf("Cluster %s corriendo en el puerto %s\n", role, port)
	http.ListenAndServe(":"+port, nil)
}

func writeHandler(w http.ResponseWriter, r *http.Request) {
	var data map[string]string

	// Leer y decodificar el cuerpo de la solicitud
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		fmt.Println("Error al decodificar JSON en leader:", err)
		http.Error(w, "Datos inválidos", http.StatusBadRequest)
		return
	}
	key := data["key"]
	value := data["value"]

	// Guardar en el archivo
	mutex.Lock()
	err = appendToFile(key, value)
	mutex.Unlock()

	if err != nil {
		http.Error(w, "Error al escribir en la base de datos", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Escritura recibida: %s = %s\n", key, value)
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
}

// AppendToFile agrega la clave y el valor al archivo
func appendToFile(key, value string) error {
	file, err := os.OpenFile("/home/brayan/Documentos/Raft_Proyecto1_TET/database.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("%s=%s\n", key, value))
	return err
}

// ReadFromFile lee el valor correspondiente a la clave del archivo
func readFromFile(key string) (string, error) {
	file, err := os.Open("/home/brayan/Documentos/Raft_Proyecto1_TET/database.txt")
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
