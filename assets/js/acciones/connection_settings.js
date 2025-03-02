// Actualiza el navbar basándose en el estado almacenado en el JSON.
export async function updateNavbar() {
  try {
    const res = await fetch("/connection_status");
    if (!res.ok) {
      console.log("No se pudo obtener el estado de la conexión");
      return;
    }
    const data = await res.json();
    const connectionIndicator = document.getElementById("connectionStatus");
    const instanceUrlEl = document.getElementById("instanceUrl");
    // Si no hay conexiones o el flag active es false, se marca desconectado
    if (!data.connections || data.connections.length === 0 || data.active === false) {
      if (connectionIndicator) connectionIndicator.style.backgroundColor = "red";
      if (instanceUrlEl) instanceUrlEl.textContent = "Desconectado";
    } else {
      // Si hay conexión activa, se toma la conexión actual según "current"
      const currentIndex = data.current;
      if (currentIndex >= 0 && currentIndex < data.connections.length) {
        const activeConn = data.connections[currentIndex];
        if (connectionIndicator) connectionIndicator.style.backgroundColor = "#28a745";
        if (instanceUrlEl) instanceUrlEl.textContent = activeConn.domain;
      } else {
        if (connectionIndicator) connectionIndicator.style.backgroundColor = "red";
        if (instanceUrlEl) instanceUrlEl.textContent = "Desconectado";
      }
    }
  } catch (error) {
    console.error("Error en updateNavbar:", error);
  }
}

// Lista todas las conexiones y actualiza el contenedor correspondiente
export async function listConnections() {
  try {
    const res = await fetch("/getconnections");
    if (!res.ok) {
      console.log("No se encontraron conexiones.");
      return;
    }
    // Se espera que el JSON tenga la estructura { connections: [...], current: número }
    const data = await res.json();
    const conns = data.connections;
    const currentIndex = data.current;
    let html = "<ul class='list-group'>";
    conns.forEach((conn, index) => {
      const isActive = index === currentIndex;
      html += `<li class="list-group-item d-flex justify-content-between align-items-center">
                ${conn.domain} - ${conn.correo}
                <div>
                  <button class="btn btn-sm ${isActive ? "btn-success" : "btn-outline-primary"}" onclick="setCurrent(${index})" ${isActive ? "disabled" : ""}>
                    ${isActive ? "Conectado" : "Conectar"}
                  </button>
                  <button class="btn btn-sm btn-outline-danger ms-2" onclick="deleteConnection(${index})">
                    Eliminar
                  </button>
                </div>
              </li>`;
    });
    html += "</ul>";
    document.getElementById("connectionsList").innerHTML = html;
  } catch (error) {
    console.error("Error al listar conexiones:", error);
  }
}

// Función global para eliminar una conexión
window.deleteConnection = async function(index) {
  try {
    const res = await fetch(`/deleteconnection?index=${index}`);
    if (res.ok) {
      alert("Conexión eliminada");
      await listConnections();
      await checkConnectionStatus();
      await updateNavbar();
    } else {
      alert("Error al eliminar la conexión");
    }
  } catch (error) {
    console.error("Error al eliminar conexión:", error);
    alert("Error al eliminar la conexión: " + error);
  }
};

// Función global para cambiar la conexión actual (para usar en onclick)
window.setCurrent = async function(index) {
  try {
    const res = await fetch(`/setcurrent?index=${index}`);
    if (res.ok) {
      alert("Conexión actualizada");
      await checkConnectionStatus();
      await listConnections();
      await updateNavbar();
    } else {
      alert("Error al cambiar la conexión");
    }
  } catch (error) {
    console.error(error);
    alert("Error al cambiar la conexión: " + error);
  }
};

// Comprueba el estado de conexión y actualiza el indicador en el navbar.
// Se basa en los valores que el usuario haya ingresado en el formulario.
export async function checkConnectionStatus() {
  const domainEl = document.getElementById("domain");
  const correoEl = document.getElementById("correo");
  const tokenEl = document.getElementById("token");
  if (!domainEl || !correoEl || !tokenEl) return false;

  const domain = domainEl.value;
  const correo = correoEl.value;
  const token = tokenEl.value;

  const connectionIndicator = document.getElementById("connectionStatus");
  const instanceUrlEl = document.getElementById("instanceUrl");

  if (!domain || !correo || !token) {
    console.log("No hay conexión configurada");
    if (connectionIndicator) connectionIndicator.style.backgroundColor = "red";
    if (instanceUrlEl) instanceUrlEl.textContent = "Desconectado";
    return false;
  }

  const bodyData = { domain, correo, token };
  try {
    const res = await fetch("/testjira", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(bodyData)
    });
    if (res.ok) {
      if (connectionIndicator) connectionIndicator.style.backgroundColor = "#28a745"; // verde
      if (instanceUrlEl) instanceUrlEl.textContent = domain;
      return true;
    } else {
      if (connectionIndicator) connectionIndicator.style.backgroundColor = "red";
      if (instanceUrlEl) instanceUrlEl.textContent = "Desconectado";
      return false;
    }
  } catch (error) {
    console.error("Error en checkConnectionStatus:", error);
    if (connectionIndicator) connectionIndicator.style.backgroundColor = "red";
    if (instanceUrlEl) instanceUrlEl.textContent = "Desconectado";
    return false;
  }
}

// Asigna listener al formulario para comprobar la conexión y mostrar el mensaje resultante.
// Al hacer "Check", se enviarán los datos; si la conexión es exitosa, se guarda la conexión, se marca active = true, se vacían los campos y se actualiza la lista y el navbar.
// Si falla, se muestra un mensaje de error.
export async function submitJiraForm() {
  const form = document.getElementById("jiraForm");
  if (!form) return;
  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    const domain = document.getElementById("domain").value;
    const correo = document.getElementById("correo").value;
    const token = document.getElementById("token").value;
    const bodyData = { domain, correo, token };

    try {
      const res = await fetch("/testjira", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(bodyData)
      });
      const data = await res.json();
      if (res.ok) {
        alert(data.message); // Muestra el mensaje devuelto por el endpoint (p. ej.: "¡Conexión exitosa y guardada!" o "Ya existe la conexión!")
        clearConnectionForm(); // Vacía los campos tras la comprobación exitosa
        await listConnections();
        await checkConnectionStatus();
        await updateNavbar();
      } else {
        alert(data.message || "Error al conectar. Verifica tus datos.");
      }
    } catch (error) {
      console.error("Error al probar conexión:", error);
      alert("Error al probar conexión: " + error);
    }
  });
}

// Función para limpiar los campos del formulario (si existen)
function clearConnectionForm() {
  const domainEl = document.getElementById("domain");
  const correoEl = document.getElementById("correo");
  const tokenEl = document.getElementById("token");
  if (domainEl) domainEl.value = "";
  if (correoEl) correoEl.value = "";
  if (tokenEl) tokenEl.value = "";
}

// Inicialización: en la página de settings se espera que el usuario ingrese los datos manualmente,
// pero se actualiza la lista de conexiones y el navbar según el JSON.
export async function initConnectionSettings() {
  await checkConnectionStatus();
  await listConnections();
  await submitJiraForm();
  await updateNavbar();
}
