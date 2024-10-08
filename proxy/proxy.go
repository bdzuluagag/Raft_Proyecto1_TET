package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	leaderURL    = "http://localhost:8081" // Asegúrate de que este puerto sea el del líder
	followerURL1 = "http://localhost:8082" // Primer seguidor
	followerURL2 = "http://localhost:8083" // Segundo seguidor
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Por favor, provee un puerto.")
		return
	}
	port := os.Args[1]

	http.HandleFunc("/write", writeHandler)
	http.HandleFunc("/read/", readHandler)

	fmt.Printf("Proxy server running on port %s\n", port)
	http.ListenAndServe(":"+port, nil)
}

func writeHandler(w http.ResponseWriter, r *http.Request) {
	// Redirigir al líder
	fmt.Println("Redirecting write to leader")
	resp, err := http.Post(leaderURL+"/write", "application/json", r.Body)
	if err != nil {
		http.Error(w, "Failed to write to leader", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Copiar el cuerpo de la respuesta
	w.WriteHeader(resp.StatusCode)
	body, _ := ioutil.ReadAll(resp.Body)
	w.Write(body)
}

func readHandler(w http.ResponseWriter, r *http.Request) {
	// Obtener la clave desde la URL
	key := r.URL.Path[len("/read/"):] // Extraer la clave de la ruta
	fmt.Println("Redirecting read to follower")

	// Redirigir a un follower aleatorio
	var followerURL string
	if key[len(key)-1]%2 == 0 { // Simple condición para redirigir entre followers
		followerURL = followerURL1
	} else {
		followerURL = followerURL2
	}

	// Redirigir al primer seguidor
	resp, err := http.Get(followerURL + "/read/" + key) // Usar la clave correcta
	if err != nil {
		http.Error(w, "Failed to read from follower", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Copiar el cuerpo de la respuesta
	w.WriteHeader(resp.StatusCode)
	body, _ := ioutil.ReadAll(resp.Body)
	w.Write(body)
}
