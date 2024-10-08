package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	leaderURL    = "http://localhost:8081" // Asegúrate de que este puerto sea el del líder
	followerURL1 = "http://localhost:8082" // Primer seguidor
	followerURL2 = "http://localhost:8084" // Segundo seguidor
)

func main() {
	http.HandleFunc("/write", writeHandler)
	http.HandleFunc("/read/", readHandler) // Nota el cambio aquí para incluir la barra

	fmt.Println("Proxy server running on port 8080")
	http.ListenAndServe(":8080", nil)
}

func writeHandler(w http.ResponseWriter, r *http.Request) {
	/*var record map[string]string

	err := json.NewDecoder(r.Body).Decode(&record)
	if err != nil {
		http.Error(w, "Invalid write data", http.StatusBadRequest)
		return
	}*/

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

	// Redirigir al primer seguidor
	resp, err := http.Get(followerURL1 + "/read/" + key) // Usar la clave correcta
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
