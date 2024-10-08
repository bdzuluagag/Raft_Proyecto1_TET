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

var mutex = &sync.Mutex{}

//const dbFile = "database.txt"

func main() {
	http.HandleFunc("/read/", readHandler)

	fmt.Println("Seguidor escuchando en el puerto 8082")
	http.ListenAndServe(":8082", nil)
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

// ReadFromFile lee el valor correspondiente a la clave del archivo
func readFromFile(key string) (string, error) {
	file, err := os.Open("C:/Users/Usuario/Documents/go_projects/Raft_Proyecto1_TET/database.txt")
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
