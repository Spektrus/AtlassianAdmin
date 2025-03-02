package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-resty/resty/v2"
)

// 2. Función actualizada para obtener los estados usando el endpoint /rest/api/3/statuses/search
func obtenerEstadosJira(client *resty.Client) ([]JiraStatus, error) {
	resp, err := client.R().Get("/rest/api/3/statuses/search")
	if err != nil {
		return nil, fmt.Errorf("error en petición a Jira (estados): %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("error en la petición (estados): %d - %s", resp.StatusCode(), resp.Status())
	}

	var resBody JiraStatusSearchResponse
	if err := json.Unmarshal(resp.Body(), &resBody); err != nil {
		return nil, fmt.Errorf("error al parsear JSON (estados): %w", err)
	}
	return resBody.Values, nil
}

// 3. Función para obtener todos los proyectos de Jira
func obtenerProyectosJira(client *resty.Client) ([]JiraProject, error) {
	var allProjects []JiraProject
	startAt := 0
	maxResults := 50
	hasMore := true

	for hasMore {
		resp, err := client.R().
			SetQueryParam("startAt", strconv.Itoa(startAt)).
			SetQueryParam("maxResults", strconv.Itoa(maxResults)).
			Get("/rest/api/2/project/search")
		if err != nil {
			return nil, fmt.Errorf("error en petición a Jira (proyectos): %w", err)
		}
		if resp.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("error en la petición (proyectos): %d - %s", resp.StatusCode(), resp.Status())
		}

		var resBody JiraProjectSearchResponse
		if err := json.Unmarshal(resp.Body(), &resBody); err != nil {
			return nil, fmt.Errorf("error al parsear JSON (proyectos): %w", err)
		}

		allProjects = append(allProjects, resBody.Values...)
		hasMore = !resBody.IsLast
		if hasMore {
			startAt += maxResults
		}
	}

	log.Println("Total proyectos descargados:", len(allProjects))
	return allProjects, nil
}

// Función para obtener todos los workflows y sus transiciones de Jira usando la API v3
func obtenerWorkflowsJira(client *resty.Client) ([]JiraWorkflow, error) {
	var allWorkflows []JiraWorkflow
	startAt := 0
	maxResults := 50

	// Parámetros opcionales (ajusta estos valores según tus necesidades)
	workflowNames := []string{} // Ejemplo: []string{"Workflow1", "Workflow2"}
	expand := "transitions,transitions.rules,transitions.properties,statuses,statuses.properties,default,schemes,projects,hasDraftWorkflow,operations"
	queryString := "" // Cadena de búsqueda opcional
	orderBy := "name" // Ordena por nombre (puede ser "created", "updated", etc.)
	//isActive := true  // Filtra workflows activos

	hasMore := true
	for hasMore {
		// Construir los parámetros de consulta usando url.Values para poder repetir claves (por ejemplo, workflowName)
		params := url.Values{}
		params.Add("startAt", strconv.Itoa(startAt))
		params.Add("maxResults", strconv.Itoa(maxResults))
		params.Add("expand", expand)
		// Agregar workflowName para cada nombre proporcionado
		for _, name := range workflowNames {
			params.Add("workflowName", name)
		}
		if queryString != "" {
			params.Add("queryString", queryString)
		}
		if orderBy != "" {
			params.Add("orderBy", orderBy)
		}
		//params.Add("isActive", strconv.FormatBool(isActive))

		// Realizar la petición a la API v3 de workflows
		resp, err := client.R().
			SetQueryParamsFromValues(params).
			Get("/rest/api/3/workflow/search")
		if err != nil {
			return nil, fmt.Errorf("error en petición a Jira (workflows): %w", err)
		}
		if resp.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("error en la petición (workflows): %d - %s", resp.StatusCode(), resp.Status())
		}

		var resBody JiraWorkflowSearchResponse
		if err := json.Unmarshal(resp.Body(), &resBody); err != nil {
			return nil, fmt.Errorf("error al parsear JSON (workflows): %w", err)
		}

		allWorkflows = append(allWorkflows, resBody.Values...)
		hasMore = !resBody.IsLast
		if hasMore {
			startAt += maxResults
		}
	}

	log.Println("Total workflows descargados:", len(allWorkflows))
	return allWorkflows, nil
}

// Función principal que ejecuta la consulta a Jira y agrupa todos los datos
func ejecutarConsultaJira(domain, correo, token string, incluirProyectos, incluirWorkflows, incluirEstados bool) (map[string]interface{}, error) {
	client := conectarAJira(domain, correo, token)

	// Creamos el JSON maestro donde se guardarán todos los datos
	maestro := make(map[string]interface{})

	// Consultamos estados si se requiere
	if incluirEstados {
		estados, err := obtenerEstadosJira(client)
		if err != nil {
			return nil, err
		}
		maestro["estados"] = estados
	}

	// Consultamos proyectos si se requiere
	if incluirProyectos {
		proyectos, err := obtenerProyectosJira(client)
		if err != nil {
			return nil, err
		}
		maestro["proyectos"] = proyectos
	}

	// Consultamos workflows si se requiere
	if incluirWorkflows {
		workflows, err := obtenerWorkflowsJira(client)
		if err != nil {
			return nil, err
		}
		maestro["workflows"] = workflows
	}

	return maestro, nil
}

// conectarAJira configura la conexión a Jira y devuelve el cliente.
func conectarAJira(domain, correo, token string) *resty.Client {
	client := resty.New()
	client.SetBaseURL(domain)
	credenciales := correo + ":" + token
	auth := base64.StdEncoding.EncodeToString([]byte(credenciales))
	client.SetHeader("Authorization", "Basic "+auth)
	client.SetHeader("Accept", "application/json")
	return client
}
