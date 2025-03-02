// Función para obtener las credenciales almacenadas desde /connection_status
async function getStoredCredentials() {
  try {
    const res = await fetch("/connection_status");
    if (!res.ok) {
      console.log("No se pudo obtener el estado de la conexión");
      return null;
    }
    const data = await res.json();
    if (!data.active || !data.connections || data.connections.length === 0) {
      return null;
    }
    const current = data.current;
    if (current === -1 || current >= data.connections.length) {
      return null;
    }
    // Retorna la conexión activa extraída del arreglo.
    return data.connections[current];
  } catch (error) {
    console.error("Error al obtener credenciales almacenadas:", error);
    return null;
  }
}

// Función para renderizar las tablas con los datos obtenidos
function renderDataTables(data) {
  const resultado = document.getElementById("resultado");
  resultado.innerHTML = ""; // Limpiar contenido previo

  // Tabla para "Estados"
  if (data.estados) {
    const headingEstados = document.createElement("h2");
    headingEstados.textContent = "Estados";
    resultado.appendChild(headingEstados);

    const estadosTable = document.createElement("table");
    estadosTable.classList.add("table", "table-striped");
    const theadEstados = document.createElement("thead");
    theadEstados.innerHTML = `<tr>
      <th>ID</th>
      <th>Nombre</th>
      <th>Descripción</th>
    </tr>`;
    estadosTable.appendChild(theadEstados);

    const tbodyEstados = document.createElement("tbody");
    data.estados.forEach(estado => {
      const tr = document.createElement("tr");
      tr.innerHTML = `<td>${estado.id}</td>
                      <td>${estado.name}</td>
                      <td>${estado.description || ""}</td>`;
      tbodyEstados.appendChild(tr);
    });
    estadosTable.appendChild(tbodyEstados);
    resultado.appendChild(estadosTable);
  }

  // Tabla para "Proyectos"
  if (data.proyectos) {
    const headingProyectos = document.createElement("h2");
    headingProyectos.textContent = "Proyectos";
    resultado.appendChild(headingProyectos);

    const proyectosTable = document.createElement("table");
    proyectosTable.classList.add("table", "table-striped");
    const theadProyectos = document.createElement("thead");
    theadProyectos.innerHTML = `<tr>
      <th>Clave</th>
      <th>Nombre</th>
    </tr>`;
    proyectosTable.appendChild(theadProyectos);

    const tbodyProyectos = document.createElement("tbody");
    data.proyectos.forEach(proyecto => {
      const tr = document.createElement("tr");
      tr.innerHTML = `<td>${proyecto.key}</td>
                      <td>${proyecto.name}</td>`;
      tbodyProyectos.appendChild(tr);
    });
    proyectosTable.appendChild(tbodyProyectos);
    resultado.appendChild(proyectosTable);
  }

  // Tabla para "Workflows"
  if (data.workflows) {
    const headingWorkflows = document.createElement("h2");
    headingWorkflows.textContent = "Workflows";
    resultado.appendChild(headingWorkflows);

    const workflowsTable = document.createElement("table");
    workflowsTable.classList.add("table", "table-striped");
    const theadWorkflows = document.createElement("thead");
    theadWorkflows.innerHTML = `<tr>
      <th>Workflow</th>
      <th>Transiciones</th>
    </tr>`;
    workflowsTable.appendChild(theadWorkflows);

    const tbodyWorkflows = document.createElement("tbody");
    data.workflows.forEach(workflow => {
      const tr = document.createElement("tr");
      // Se asume que workflow.id es un objeto con el nombre del workflow en "name"
      const workflowName = workflow.id.name || "";
      // Procesar cada transición para mostrar sus estados de origen y destino
      let transitionsHtml = "";
      if (workflow.transitions && Array.isArray(workflow.transitions)) {
        transitionsHtml = '<ul>' + workflow.transitions.map(t => {
          const fromStates = (t.from && t.from.length > 0) ? t.from.join(", ") : "Inicio";
          return `<li>Desde: ${fromStates} → Hasta: ${t.to}</li>`;
        }).join("") + '</ul>';
      }
      tr.innerHTML = `<td>${workflowName}</td>
                      <td>${transitionsHtml}</td>`;
      tbodyWorkflows.appendChild(tr);
    });
    workflowsTable.appendChild(tbodyWorkflows);
    resultado.appendChild(workflowsTable);
  }
}

// Función para asignar el listener al formulario y ejecutar la consulta a Jira
export async function submitJiraFormData() {
  const form = document.getElementById("jiraForm");
  if (!form) return;

  form.addEventListener("submit", async (e) => {
    e.preventDefault();

    // Obtener las credenciales almacenadas
    const creds = await getStoredCredentials();
    if (!creds || !creds.domain || !creds.correo || !creds.token) {
      alert("No hay conexión configurada. Por favor, configura la conexión en la página de Connection Settings.");
      return;
    }

    // Obtener el estado de los checkboxes
    const proyectos = document.getElementById("proyectos").checked;
    const workflows = document.getElementById("workflows").checked;
    const estados = document.getElementById("estados").checked;

    const bodyData = {
      domain: creds.domain,
      correo: creds.correo,
      token: creds.token,
      Proyectos: proyectos,
      Workflows: workflows,
      Estados: estados
    };

    try {
      const res = await fetch("/execute", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(bodyData)
      });
      const data = await res.json();
      // Renderizar la información en tablas en lugar de un JSON crudo
      renderDataTables(data);
    } catch (error) {
      console.error("Error al ejecutar consulta a Jira:", error);
      alert("Error al ejecutar consulta: " + error);
    }
  });
}

// Función de inicialización para data.html
export async function initGetting() {
  await submitJiraFormData();
}
