package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

const pidFile = "server.pid"

// -------------------
// Funciones para la web (dashboard, search y renderizado)
// -------------------

func handleConecctionSettings(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title":      "Connection Settings",
		"ActivePage": "connection_settings",
	}
	renderTemplate(w, "connection_settings", data)
}

func handleData(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title":      "Data",
		"ActivePage": "data",
	}
	renderTemplate(w, "data", data)
}

func handleStates(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title":      "States",
		"ActivePage": "States",
	}
	renderTemplate(w, "states", data)
}
func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	tmplPath := fmt.Sprintf("../pages/%s.html", tmpl)
	log.Println("Cargando plantilla:", tmplPath)

	templates, err := template.ParseFiles(
		"../pages/templates/base.html",
		"../pages/templates/aside.html",
		"../pages/templates/navbar.html",
		"../pages/templates/modal.html",
		tmplPath,
	)
	if err != nil {
		http.Error(w, "Error al cargar la plantilla: "+err.Error(), http.StatusInternalServerError)
		log.Println("Error cargando plantilla:", err)
		return
	}

	err = templates.Execute(w, data)
	if err != nil {
		http.Error(w, "Error al renderizar la plantilla", http.StatusInternalServerError)
		log.Println("Error ejecutando plantilla:", err)
	}
}

// Se asume que las funciones de Jira (handleJiraExecution y demás) ya están definidas (por ejemplo, en functions.go)

// -------------------
// Funciones para iniciar/detener el servidor
// -------------------

func runServer() {
	// Configurar el router de Gorilla Mux
	router := mux.NewRouter()

	// Ruta principal: redirige a /dashboard
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/connection_settings", http.StatusSeeOther)
	}).Methods("GET")

	router.HandleFunc("/connection_settings", handleConecctionSettings).Methods("GET")
	router.HandleFunc("/data", handleData).Methods("GET")
	router.HandleFunc("/states", handleStates).Methods("GET")
	// API POINTS
	router.HandleFunc("/connection_status", handleConnectionStatus).Methods("GET")
	router.HandleFunc("/setcurrent", handleSetCurrentConnection).Methods("GET")
	router.HandleFunc("/testjira", handleTestJira).Methods("POST")
	router.HandleFunc("/getjson", handleGetJSONKey).Methods("GET")
	router.HandleFunc("/getconnections", handleGetConnections).Methods("GET")
	router.HandleFunc("/deleteconnection", handleDeleteConnection).Methods("GET")
	// Ruta POST para ejecutar la consulta a Jira
	router.HandleFunc("/execute", handleJiraExecution).Methods("POST")
	// Servir archivos estáticos
	router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("../assets/"))))

	// Logger simple para cada solicitud
	loggedRouter := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Solicitud: %s %s", r.Method, r.URL.Path)
		router.ServeHTTP(w, r)
	})

	// Escribir el PID actual en un archivo para controlarlo
	pid := os.Getpid()
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644); err != nil {
		log.Printf("No se pudo guardar el PID: %v", err)
	}

	// Configurar el servidor
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      loggedRouter,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Servidor corriendo en http://localhost:%s (PID: %d)", port, pid)
	log.Fatal(server.ListenAndServe())
}

func stopServer() {
	// Leer el PID desde el archivo
	pidData, err := os.ReadFile(pidFile)
	if err != nil {
		fmt.Println("No se encontró el archivo de PID. ¿El servidor está en ejecución?")
		return
	}
	pid, err := strconv.Atoi(string(pidData))
	if err != nil {
		log.Fatalf("Error al convertir el PID: %v", err)
	}

	// Buscar y terminar el proceso
	process, err := os.FindProcess(pid)
	if err != nil {
		log.Fatalf("No se pudo encontrar el proceso: %v", err)
	}
	// Enviar señal SIGTERM para detener el servidor ordenadamente
	if err := process.Signal(syscall.SIGTERM); err != nil {
		log.Fatalf("Error al detener el servidor: %v", err)
	}

	// Eliminar el archivo PID
	os.Remove(pidFile)
	fmt.Println("Servidor detenido.")
}

func main() {
	if len(os.Args) > 1 {
		cmd := os.Args[1]
		switch cmd {
		case "start":
			runServer()
		case "stop":
			stopServer()
		case "toggle":
			if _, err := os.ReadFile(pidFile); err == nil {
				stopServer()
			} else {
				runServer()
			}
		default:
			fmt.Println("Uso: app [start|stop|toggle]")
		}
	} else {
		// Si no se pasan argumentos, arranca el servidor por defecto.
		runServer()
	}
}
