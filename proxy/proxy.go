package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

var leaderURL = "http://localhost:9081" // Se actualizará dinámicamente

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
	// Verificar dinámicamente quién es el líder
	leaderURL = detectLeader()

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

	// Redirigir a un follower basado en la clave
	var followerURLs = []string{"http://localhost:9082", "http://localhost:9083"}
	firstFollower := followerURLs[key[len(key)-1]%2]      // Determina cuál intentar primero
	secondFollower := followerURLs[(key[len(key)-1]+1)%2] // El otro follower

	// Primero intenta con el primer follower
	fmt.Println("Redirecting read to follower:", firstFollower)
	resp, err := http.Get(firstFollower + "/read/" + key)
	if err != nil {
		// Si falla, intenta con el segundo follower
		fmt.Printf("Error al leer del follower %s: %v. Intentando con el otro follower.\n", firstFollower, err)
		resp, err = http.Get(secondFollower + "/read/" + key)
		if err != nil {
			// Si también falla, devolver error
			http.Error(w, "Failed to read from both followers", http.StatusInternalServerError)
			return
		}
	}

	defer resp.Body.Close()

	// Copiar el cuerpo de la respuesta
	w.WriteHeader(resp.StatusCode)
	body, _ := ioutil.ReadAll(resp.Body)
	w.Write(body)
}

func detectLeader() string {
	// Lista de posibles nodos (líder o followers)
	nodes := []string{"http://localhost:9081", "http://localhost:9082", "http://localhost:9083"}

	for _, node := range nodes {
		resp, err := http.Get(node + "/is_leader")
		if err != nil {
			fmt.Printf("Error al consultar el nodo %s: %v\n", node, err)
			continue
		}
		defer resp.Body.Close()

		var result map[string]bool
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err == nil && result["is_leader"] {
			fmt.Printf("Nodo líder detectado en: %s\n", node)
			return node
		}
	}

	// Si no se encuentra un líder, devuelve el valor predeterminado (8081)
	fmt.Println("Líder no detectado, usando valor predeterminado.")
	return "http://localhost:9081"
}
