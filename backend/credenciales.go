package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// ----------------------------------------------------------------
// Constantes y Helpers para gestionar el fichero JSON de conexiones
// ----------------------------------------------------------------

// ----------------------------------------------------------------
// Endpoints para la gestión de conexiones
// ----------------------------------------------------------------

// handleGetConnections devuelve todas las conexiones y el índice de la conexión actual.
func handleGetConnections(w http.ResponseWriter, r *http.Request) {
	jsonData, err := readJSONFile(jsonFilePath)
	if err != nil {
		http.Error(w, "Error leyendo JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jsonData)
}

// handleAddConnection agrega una nueva conexión al fichero JSON y la marca como actual.
func handleAddConnection(w http.ResponseWriter, r *http.Request) {
	var newConn Credentials
	if err := json.NewDecoder(r.Body).Decode(&newConn); err != nil {
		http.Error(w, "JSON inválido: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Leer datos existentes; si no existe, se crea un nuevo mapa.
	jsonData, err := readJSONFile(jsonFilePath)
	if err != nil {
		log.Println("No se encontró el fichero existente, se creará uno nuevo")
		jsonData = make(map[string]interface{})
	}

	// Obtener o inicializar el arreglo de conexiones.
	conns, ok := jsonData["connections"].([]interface{})
	if !ok {
		conns = []interface{}{}
	}

	// Crear el mapa de la nueva conexión.
	connMap := map[string]string{
		"domain": newConn.Domain,
		"correo": newConn.Correo,
		"token":  newConn.Token,
	}

	// Agregar la nueva conexión y marcarla como la actual.
	conns = append(conns, connMap)
	jsonData["connections"] = conns
	jsonData["current"] = len(conns) - 1

	if err := writeJSONFile(jsonFilePath, jsonData); err != nil {
		http.Error(w, "Error guardando JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jsonData)
}

func handleSetCurrentConnection(w http.ResponseWriter, r *http.Request) {
	indexStr := r.URL.Query().Get("index")
	if indexStr == "" {
		http.Error(w, "Falta el parámetro 'index'", http.StatusBadRequest)
		return
	}
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		http.Error(w, "Índice inválido: "+err.Error(), http.StatusBadRequest)
		return
	}

	existingData := make(map[string]interface{})
	data, err := os.ReadFile(jsonFilePath)
	if err != nil {
		http.Error(w, "Error leyendo el fichero JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.Unmarshal(data, &existingData); err != nil {
		http.Error(w, "Error parseando JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

	conns, ok := existingData["connections"].([]interface{})
	if !ok || index < 0 || index >= len(conns) {
		http.Error(w, "Índice fuera de rango", http.StatusBadRequest)
		return
	}

	// Actualizamos el índice de la conexión actual
	existingData["current"] = index

	// Si hay conexiones, marcamos active como true; de lo contrario, false.
	if len(conns) > 0 {
		existingData["active"] = true
	} else {
		existingData["active"] = false
	}

	updatedData, err := json.MarshalIndent(existingData, "", "  ")
	if err != nil {
		http.Error(w, "Error al formatear JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := os.WriteFile(jsonFilePath, updatedData, 0644); err != nil {
		http.Error(w, "Error al guardar JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Conexión actual actualizada",
	})
}

// handleDeleteConnection elimina la conexión especificada por el parámetro "index".
func handleDeleteConnection(w http.ResponseWriter, r *http.Request) {
	indexStr := r.URL.Query().Get("index")
	if indexStr == "" {
		http.Error(w, "Falta el parámetro 'index'", http.StatusBadRequest)
		return
	}
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		http.Error(w, "Índice inválido: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Leer el JSON actual
	jsonData, err := readJSONFile(jsonFilePath)
	if err != nil {
		http.Error(w, "Error leyendo JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

	conns, ok := jsonData["connections"].([]interface{})
	if !ok || index < 0 || index >= len(conns) {
		http.Error(w, "Índice fuera de rango", http.StatusBadRequest)
		return
	}

	// Eliminar la conexión indicada
	conns = append(conns[:index], conns[index+1:]...)
	jsonData["connections"] = conns

	// Leer el índice actual
	current, ok := jsonData["current"].(float64)
	if !ok {
		current = -1
	}
	currentIndex := int(current)

	// Si ya no quedan conexiones, se establece current en -1 y active en false.
	if len(conns) == 0 {
		jsonData["current"] = -1
		jsonData["active"] = false
	} else {
		// Si la conexión eliminada era la que estaba en uso, ponemos active en false.
		if currentIndex == index {
			jsonData["active"] = false
			// También podrías decidir cambiar current a otro índice, por ejemplo el último,
			// pero si prefieres dejarlo sin conexión, lo dejamos en false.
			jsonData["current"] = -1
		} else if currentIndex > index {
			// Si el índice actual es mayor, se reduce en uno.
			jsonData["current"] = currentIndex - 1
		}
	}

	if err := writeJSONFile(jsonFilePath, jsonData); err != nil {
		http.Error(w, "Error guardando JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jsonData)
}

// handleConnectionStatus devuelve la conexión actual para que el navbar pueda mostrar el estado.
func handleConnectionStatus(w http.ResponseWriter, r *http.Request) {
	jsonData, err := readJSONFile(jsonFilePath)
	if err != nil {
		http.Error(w, "Error leyendo JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Si no hay conexiones, aseguramos que active sea false.
	conns, ok := jsonData["connections"].([]interface{})
	if !ok || len(conns) == 0 {
		jsonData["active"] = false
	} else {
		// Si hay conexiones, asumimos que se ha probado la conexión y active se actualiza desde /testjira.
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jsonData)
}

// getCredentials lee el fichero JSON y devuelve la conexión actual (según el índice "current").
func getCredentials() (Credentials, error) {
	var creds Credentials
	jsonData, err := readJSONFile(jsonFilePath)
	if err != nil {
		return creds, fmt.Errorf("error leyendo fichero JSON: %w", err)
	}

	// Obtener el array de conexiones
	conns, ok := jsonData["connections"].([]interface{})
	if !ok || len(conns) == 0 {
		return creds, fmt.Errorf("no hay conexiones almacenadas")
	}

	// Leer el índice de la conexión actual
	currentIndexFloat, ok := jsonData["current"].(float64)
	if !ok {
		return creds, fmt.Errorf("no se encontró el índice de la conexión actual")
	}
	currentIndex := int(currentIndexFloat)
	if currentIndex < 0 || currentIndex >= len(conns) {
		return creds, fmt.Errorf("índice de conexión actual fuera de rango")
	}

	// Convertir la conexión actual a JSON y deserializarla en el struct Credentials
	credBytes, err := json.Marshal(conns[currentIndex])
	if err != nil {
		return creds, fmt.Errorf("error convirtiendo credenciales: %w", err)
	}
	if err := json.Unmarshal(credBytes, &creds); err != nil {
		return creds, fmt.Errorf("error parseando credenciales: %w", err)
	}
	return creds, nil
}

func generateFileName(domain string) string {
	// Eliminar prefijos http:// y https://
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "https://")
	// Eliminar posibles barras al final o dentro (puedes elegir reemplazarlas por guiones o guiones bajos)
	domain = strings.ReplaceAll(domain, "/", "_")
	return domain + ".json"
}
