package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("¿Desea leer o escribir? (r/w)")
	operation, _ := reader.ReadString('\n')
	operation = operation[:len(operation)-1]

	if operation == "w" {
		fmt.Println("Ingrese la clave:")
		key, _ := reader.ReadString('\n')
		key = key[:len(key)-1]

		fmt.Println("Ingrese el valor:")
		value, _ := reader.ReadString('\n')
		value = value[:len(value)-1]

		writeData(key, value)
	} else if operation == "r" {
		fmt.Println("Ingrese la clave a leer:")
		key, _ := reader.ReadString('\n')
		key = key[:len(key)-1]

		readData(key)
	} else {
		fmt.Println("Operación no válida")
	}
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
