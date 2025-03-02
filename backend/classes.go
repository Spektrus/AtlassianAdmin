package main

// Estructura para recibir datos desde el formulario
type RequestData struct {
	Domain    string `json:"domain"`
	Correo    string `json:"correo"` // nuevo campo para el correo
	Token     string `json:"token"`
	Proyectos bool   `json:"proyectos"`
	Workflows bool   `json:"workflows"`
	Estados   bool   `json:"estados"` // flag opcional para estados
}

type JiraProjectSearchResponse struct {
	IsLast bool          `json:"isLast"`
	Values []JiraProject `json:"values"`
}
type JiraProject struct {
	Key             string           `json:"key"`
	Name            string           `json:"name"`
	ProjectCategory *ProjectCategory `json:"projectCategory,omitempty"`
}
type ProjectCategory struct {
	Name string `json:"name"`
}

type JiraWorkflowSearchResponse struct {
	IsLast bool           `json:"isLast"`
	Values []JiraWorkflow `json:"values"`
}

type JiraWorkflow struct {
	ID          WorkflowID   `json:"id"`
	Transitions []Transition `json:"transitions"`
}

type WorkflowID struct {
	Name string `json:"name"`
}

type Transition struct {
	From []string `json:"from"`
	To   string   `json:"to"`
}

// Definición de la respuesta de búsqueda de estados en Jira
type JiraStatusSearchResponse struct {
	IsLast     bool         `json:"isLast"`
	MaxResults int          `json:"maxResults"`
	NextPage   string       `json:"nextPage"`
	Self       string       `json:"self"`
	StartAt    int          `json:"startAt"`
	Total      int          `json:"total"`
	Values     []JiraStatus `json:"values"`
}

// Puedes ampliar JiraStatus si necesitas más campos; por ahora se mantienen los básicos.
type JiraStatus struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// ------------------------------------------------------------
// ------------------------------------------------------------
// ------------------------------------------------------------
// TEMPORAL
// Estructura para las credenciales
type Credentials struct {
	Domain string `json:"domain"`
	Correo string `json:"correo"`
	Token  string `json:"token"`
}

// ------------------------------------------------------------
// ------------------------------------------------------------
// ------------------------------------------------------------
