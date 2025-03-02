package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

const jsonFilePath = "/home/spektrus/Escritorio/AtlassianAyudas/assets/jsons/datos.json"

// readJSONFile lee el fichero JSON y devuelve su contenido como un mapa.
func readJSONFile(filePath string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error leyendo fichero JSON: %w", err)
	}
	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil, fmt.Errorf("error parseando JSON: %w", err)
	}
	return jsonData, nil
}

// writeJSONFile escribe el mapa proporcionado en el fichero JSON.
func writeJSONFile(filePath string, data map[string]interface{}) error {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("error formateando JSON: %w", err)
	}
	return os.WriteFile(filePath, jsonBytes, 0644)
}

// mergeMaps actualiza el mapa original (oldData) con los valores del mapa newData.
func mergeMaps(oldData, newData map[string]interface{}) map[string]interface{} {
	for key, newValue := range newData {
		oldData[key] = newValue
	}
	return oldData
}

// Handler para ejecutar la API de Jira, fusionando los datos nuevos con los existentes en el fichero JSON.
func handleJiraExecution(w http.ResponseWriter, r *http.Request) {
	var form RequestData

	// Decodificar la solicitud JSON
	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		http.Error(w, "Error al decodificar JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Verificar que se han enviado los datos mínimos en el formulario
	if form.Domain == "" || form.Correo == "" || form.Token == "" {
		http.Error(w, "Faltan el dominio, el correo o el token", http.StatusBadRequest)
		return
	}

	// Obtener las credenciales de la conexión activa almacenada
	conn, err := getCredentials()
	if err != nil {
		http.Error(w, "No hay conexión activa: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Ejecutar la consulta a Jira usando las credenciales de la conexión activa
	resultados, err := ejecutarConsultaJira(conn.Domain, conn.Correo, conn.Token, form.Proyectos, form.Workflows, form.Estados)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Generar el nombre del archivo JSON a partir del dominio activo
	basePath := "/home/spektrus/Escritorio/AtlassianAyudas/assets/jsons/"
	fileName := generateFileName(conn.Domain)
	filePath := basePath + fileName

	// Leer el JSON existente (si existe) en un mapa
	existingData := make(map[string]interface{})
	if data, err := os.ReadFile(filePath); err == nil {
		if err := json.Unmarshal(data, &existingData); err != nil {
			log.Println("Error al parsear JSON existente, se sobrescribirá:", err)
		}
	} else {
		log.Println("No se encontró el fichero existente, se creará uno nuevo")
	}

	// Fusionar los datos nuevos (obtenidos de Jira) con los existentes
	updatedData := mergeMaps(existingData, resultados)

	// Convertir el mapa actualizado a JSON formateado y escribirlo en el archivo correspondiente
	dataToWrite, err := json.MarshalIndent(updatedData, "", "  ")
	if err != nil {
		log.Println("Error al formatear JSON:", err)
	} else {
		if err := os.WriteFile(filePath, dataToWrite, 0644); err != nil {
			log.Println("Error al guardar el fichero JSON:", err)
		} else {
			log.Println("Fichero JSON actualizado guardado en:", filePath)
		}
	}

	// Enviar la respuesta actualizada al cliente
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedData)
}

// handleGetJSONKey lee el fichero JSON y devuelve el valor asociado a la clave pasada como query parameter "key".
func handleGetJSONKey(w http.ResponseWriter, r *http.Request) {
	// Obtener la clave a buscar desde la query string, por ejemplo: ?key=estados
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Falta el parámetro 'key'", http.StatusBadRequest)
		return
	}

	// Ruta absoluta del fichero JSON
	filePath := "/home/spektrus/Escritorio/AtlassianAyudas/assets/jsons/credenciales.json"

	data, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, "Error leyendo fichero JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Deserializamos el contenido en un mapa para extraer la clave solicitada
	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		http.Error(w, "Error parseando JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Buscar la clave especificada
	value, ok := jsonData[key]
	if !ok {
		http.Error(w, fmt.Sprintf("No se encontró la clave %q en el JSON", key), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(value)
}

// handleTestJira prueba la conexión y, si es exitosa, guarda o actualiza la conexión en el fichero JSON.
// handleTestJira prueba la conexión y guarda o actualiza la conexión en el fichero JSON.
func handleTestJira(w http.ResponseWriter, r *http.Request) {
	var reqData struct {
		Domain string `json:"domain"`
		Correo string `json:"correo"`
		Token  string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		http.Error(w, "Error al decodificar JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	client := conectarAJira(reqData.Domain, reqData.Correo, reqData.Token)
	resp, err := client.R().Get("/rest/api/3/myself")
	if err != nil {
		http.Error(w, "Error en petición a Jira: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if resp.StatusCode() != http.StatusOK {
		http.Error(w, fmt.Sprintf("Error en la petición de prueba: %d - %s", resp.StatusCode(), resp.Status()), http.StatusInternalServerError)
		return
	}

	// Guardar o actualizar la conexión y saber si ya existía
	exists, err := addOrUpdateConnection(reqData.Domain, reqData.Correo, reqData.Token)
	if err != nil {
		http.Error(w, "Error al guardar conexión: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Actualizar el campo "active" en el JSON a true
	jsonData, err := readJSONFile(jsonFilePath)
	if err != nil {
		log.Println("No se pudo leer el JSON para actualizar el campo active:", err)
	} else {
		jsonData["active"] = true
		if err := writeJSONFile(jsonFilePath, jsonData); err != nil {
			log.Println("Error al actualizar campo active en el JSON:", err)
		}
	}

	message := "¡Conexión exitosa y guardada!"
	if exists {
		message = "Ya existe la conexión!"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": message,
	})
}

// addOrUpdateConnection revisa si ya existe la conexión (comparando dominio y correo) y, si es así,
// actualiza el token (si es necesario) y marca la conexión como actual. Devuelve true si la conexión ya existía.
func addOrUpdateConnection(domain, correo, token string) (bool, error) {
	filePath := jsonFilePath // definido en una constante
	// Leer datos existentes; si no existe, se crea un nuevo mapa.
	jsonData, err := readJSONFile(filePath)
	if err != nil {
		log.Println("No se encontró el fichero existente, se creará uno nuevo")
		jsonData = make(map[string]interface{})
	}

	var conns []interface{}
	if temp, ok := jsonData["connections"].([]interface{}); ok {
		conns = temp
	} else {
		conns = []interface{}{}
	}

	// Verificar si ya existe una conexión con el mismo dominio y correo.
	for i, c := range conns {
		if connMap, ok := c.(map[string]interface{}); ok {
			if connMap["domain"] == domain && connMap["correo"] == correo {
				// Si ya existe y el token es diferente, lo actualizamos.
				if connMap["token"] != token {
					connMap["token"] = token
				}
				// Marcar esta conexión como la actual.
				jsonData["current"] = i
				jsonData["connections"] = conns
				if err := writeJSONFile(filePath, jsonData); err != nil {
					return true, err
				}
				return true, nil
			}
		}
	}

	// Si no existe, se agrega la nueva conexión y se marca como actual.
	newConn := map[string]string{
		"domain": domain,
		"correo": correo,
		"token":  token,
	}
	conns = append(conns, newConn)
	jsonData["connections"] = conns
	jsonData["current"] = len(conns) - 1
	if err := writeJSONFile(filePath, jsonData); err != nil {
		return false, err
	}
	return false, nil
}
