package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	// Ejemplo de escritura
	writeData("exampleKey57", "exampleValue57")

	// Ejemplo de lectura
	readData("exampleKey57")
}

func writeData(key, value string) {
	data := map[string]string{"key": key, "value": value}

	jsonData, err := json.Marshal(data)
	
	resp, err := http.Post("http://localhost:8080/write", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error al enviar la escritura:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("Escritura exitosa")
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("Error en la escritura:", resp.Status, string(body))
	}
}

func readData(key string) {
	resp, err := http.Get(fmt.Sprintf("http://localhost:8080/read/%s", key))
	if err != nil {
		fmt.Println("Error al enviar la lectura:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("Respuesta de lectura:", string(body))
}
