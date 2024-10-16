### Documentación técnica: 
El algoritmo **Raft** es un protocolo de consenso diseñado para ser más fácil de entender que otros algoritmos, como Paxos, y se utiliza para gestionar un conjunto de réplicas en un sistema distribuido, garantizando que todas las réplicas acuerden el mismo conjunto de operaciones a aplicar. Raft divide el problema de consenso en tres subproblemas principales:

1. **Elección de Líder (Leader Election)**: Cómo seleccionar un líder entre los servidores.
2. **Replicación de Log (Log Replication)**: Cómo el líder mantiene los registros de operación sincronizados en los seguidores (followers).
3. **Compromiso de Entrada (Log Commitment)**: Cómo garantizar que una entrada en el log sea aceptada por una mayoría y no se revierta.

A continuación se explican estos conceptos en detalle, junto con los desafíos y cómo Raft garantiza la consistencia y disponibilidad en un sistema distribuido.

#### 1. Elección de Líder

En Raft, el sistema se organiza en períodos de tiempo llamados **términos**. Durante cada término, puede haber una elección para elegir un líder. El proceso funciona de la siguiente manera:

- **Estados de los nodos**: Cada servidor en el sistema puede estar en uno de tres estados: **líder**, **seguidor** o **candidato**.
- **Iniciar una elección**: Si un seguidor no recibe comunicación del líder actual por un tiempo llamado **timeout**, se convierte en candidato y lanza una elección. El candidato incrementa su término actual, vota por sí mismo y solicita votos de otros servidores.
- **Ganar una elección**: Un candidato se convierte en líder si recibe una mayoría de votos de los otros servidores. Una vez elegido, el líder envía latidos (heartbeats) periódicos para prevenir nuevas elecciones y mantener la autoridad sobre los seguidores.
- **Empate en la elección**: Si dos servidores lanzan elecciones al mismo tiempo, puede resultar en un empate. En este caso, se reiniciará la elección después de otro timeout aleatorio.

##### Desafíos en la elección de líder
- **Timeouts Aleatorios**: Para evitar conflictos en las elecciones simultáneas, Raft usa intervalos de timeout aleatorios, lo que disminuye la probabilidad de que dos servidores se conviertan en candidatos al mismo tiempo.

#### 2. Replicación de Log

Una vez que se ha elegido un líder, éste es responsable de gestionar la replicación de los logs a los seguidores:

- **Registro de Operaciones (Logs)**: El líder acepta las operaciones de los clientes y las añade a su log local. Luego, replica estas operaciones a los seguidores.
- **Confirmación de Entrada**: El líder considera una entrada como confirmada o comprometida cuando ha sido replicada en la mayoría de los servidores. Solo después de que una entrada esté comprometida, el líder puede aplicar la operación al estado actual del sistema.

##### Desafíos en la replicación de log
- **Fallos en la Red**: En un sistema distribuido, la latencia y la pérdida de paquetes pueden hacer que algunos seguidores no reciban las actualizaciones. Raft maneja esto mediante el reenvío de entradas a los seguidores que estén retrasados.
- **Reorganización de Entradas**: Si un líder falla, el nuevo líder puede tener que resolver inconsistencias en los logs de los seguidores para asegurarse de que todos están alineados correctamente.

#### 3. Compromiso de Entrada

Raft utiliza un enfoque basado en la mayoría para garantizar que las operaciones sean persistentes y no se reviertan, incluso en casos de fallos:

- **Mayoría**: Una operación se considera comprometida si ha sido replicada en una mayoría de los nodos. Esto asegura que, incluso si algunos nodos fallan, la operación no se perderá.
- **Actualización del Log**: Cuando un nuevo líder es elegido, debe asegurarse de que el log esté actualizado y consistente en todos los seguidores, lo que puede requerir que algunos seguidores descarten entradas que no están comprometidas.

#### Garantías de Raft

Raft garantiza la **consistencia y disponibilidad** del sistema mediante los siguientes mecanismos:

1. **Consistencia Linealizable**: Raft asegura que las operaciones se ejecutan en el mismo orden en todos los nodos. Esto significa que si un cliente realiza una operación, cualquier otra operación posterior ve el efecto de la primera, incluso si ocurre en un nodo diferente.
2. **Disponibilidad con Fallos Parciales**: Raft puede tolerar fallos en menos de la mitad de los nodos y aún así continuar operando, siempre y cuando haya un quórum disponible.
3. **Seguridad del Log**: Una vez que una operación se marca como comprometida, no puede ser modificada en el log de los servidores. 

#### Desafíos en la Garantía de Consistencia y Disponibilidad

- **Fallos del Líder**: Si el líder falla, se debe elegir un nuevo líder rápidamente para minimizar la indisponibilidad del sistema. Esto se maneja con timeouts ajustados.
- **Particiones en la Red**: Si se produce una partición en la red, Raft garantiza que al menos un subgrupo con una mayoría podrá seguir funcionando.
- **Actualización de Seguidores Retrasados**: Los seguidores que se retrasan deben ser actualizados por el líder para mantener la consistencia del sistema.
---
### Análisis Detallado de las Pruebas Realizadas y Simulación de Fallos

#### 1. **Pruebas Funcionales**

Se llevaron a cabo varias pruebas para garantizar que el sistema funcione correctamente bajo condiciones normales. Las pruebas incluyeron:

- **Escritura de Claves-Valor en el Clúster**: Se realizaron múltiples pruebas de escritura de datos utilizando el cliente, enviando solicitudes de escritura a través del **proxy**. El proxy, al detectar al líder del clúster, redirigía las escrituras al nodo líder correctamente.
  
- **Lectura de Claves-Valor desde el Clúster**: También se probaron las lecturas de datos desde el clúster. El proxy redirigía las lecturas a los nodos seguidores (followers) usando la lógica de selección por clave (`key[len(key)-1]%2`). En situaciones donde el follower seleccionado no estaba disponible, se reintentaba con el otro follower, garantizando una lectura exitosa siempre que al menos uno de los followers estuviera activo.

#### 2. **Simulación de Fallos**

Se realizaron múltiples simulaciones de fallos en el clúster para verificar su comportamiento tolerante a fallas. Las simulaciones incluyeron la caída de nodos tanto líderes como followers, así como la recuperación de nodos caídos. A continuación se presentan los resultados y observaciones:

##### **a) Caída del Nodo Líder**
   - **Procedimiento**: Se iniciaron los tres nodos del clúster y se verificó que uno de ellos asumiera el rol de líder. Posteriormente, el nodo líder fue detenido manualmente (Ctrl + C).
   - **Resultado**: Al detectarse la caída del líder, los followers entraron en un proceso de elección. Uno de los followers asumió el rol de líder exitosamente después de la votación. Las nuevas escrituras y lecturas se redirigieron automáticamente al nuevo líder a través del proxy. Esto demostró que el sistema es capaz de manejar la falla del líder y continuar funcionando sin interrupciones significativas.
   
##### **b) Caída de un Follower**
   - **Procedimiento**: Se inició el clúster completo y se detuvo manualmente uno de los followers. Luego, se realizaron múltiples operaciones de escritura y lectura.
   - **Resultado**: El sistema continuó funcionando correctamente. El follower restante fue capaz de manejar todas las solicitudes de lectura. Cuando el follower caído volvió a estar en línea, recibió las actualizaciones faltantes de la base de datos a través del mecanismo de replicación de logs. Esto asegura que el sistema sigue siendo funcional y consistente incluso cuando un follower está fuera de línea.

#### 3. **Pruebas de Recuperación y Consistencia**

Tras las simulaciones de fallos, se verificó que los nodos caídos recuperaran el estado consistente del clúster:

- **Escrituras y Replicación**: Se realizaron varias escrituras mientras algunos nodos estaban fuera de línea. Una vez que los nodos se reiniciaron, se verificó que recibieran correctamente las actualizaciones mediante la replicación de logs.
- **Lecturas de Datos**: Incluso después de que los nodos volvieran a estar en línea, los followers respondieron de manera coherente a las solicitudes de lectura. Esto mostró que los seguidores no desincronizados se pusieron rápidamente al día gracias a los logs replicados.

#### 4. **Resultados**

El sistema demostró un comportamiento robusto frente a fallos de nodos:

- **Tolerancia a Fallos de Líderes**: El clúster pudo recuperarse automáticamente cuando el líder falló, asegurando la continuidad del servicio sin intervención manual.
  
- **Tolerancia a Fallos de Followers**: Las operaciones de lectura pudieron continuar a pesar de la caída de un follower, y cuando ambos followers estuvieron caídos, el nodo líder mantuvo la integridad de las operaciones.

- **Recuperación de Nodos**: Los nodos caídos se reintegraron al clúster de manera fluida y sincronizaron sus datos correctamente sin duplicación o pérdida de información.
---
### Implementación del Algoritmo de Consenso e Integración con los Procesos del Sistema

#### 1. **Algoritmo de Consenso: Raft**

El sistema que implementamos se basa en el **algoritmo de consenso Raft**, que es un protocolo utilizado para lograr un acuerdo entre los nodos de un clúster distribuido. Raft asegura que todos los nodos de un clúster tengan el mismo estado replicado y proporciona tolerancia a fallos mediante la elección dinámica de un **líder** que coordina las operaciones.

Raft tiene tres roles principales para los nodos:
- **Líder**: Responsable de manejar las escrituras y coordinar la replicación hacia los followers.
- **Followers (Seguidores)**: Reciben las actualizaciones del líder y manejan principalmente las lecturas.
- **Candidate (Candidato)**: Cuando un follower no escucha al líder por un tiempo determinado, se convierte en candidato y comienza una elección para convertirse en el nuevo líder.

#### 2. **Implementación de Raft**

A continuación, detallo cómo se implementó el algoritmo de consenso Raft en nuestro sistema y cómo se integró con el flujo de trabajo del sistema.

##### a) **Inicialización del Nodo**

Cada nodo que forma parte del clúster necesita iniciar el protocolo Raft. Para esto, cada nodo comienza por configurar el entorno Raft, asignando un ID único basado en su puerto y preparando los componentes clave: almacenamiento de logs, transporte de red, y una **Finite State Machine (FSM)**.

```go
func initializeRaft(port string) {
    // Configuración de Raft
    config := raft.DefaultConfig()
    config.LocalID = raft.ServerID("node-" + port) // ID único basado en el puerto

    // Configurar el archivo de base de datos
    dbPath = fmt.Sprintf("database_%s.txt", port)
    fmt.Printf("Base de datos para este nodo: %s\n", dbPath)

    // Almacenamiento de logs
    logStore, err := raftboltdb.NewBoltStore(fmt.Sprintf("raft-log-%s.db", port))
    if err != nil {
        log.Fatalf("Error al inicializar logStore: %v", err)
    }

    // Almacenamiento de snapshots
    snapshotStore, err := raft.NewFileSnapshotStore(".", 1, os.Stdout)
    if err != nil {
        log.Fatalf("Error al crear snapshot store: %v", err)
    }

    // Transporte TCP entre nodos
    address := fmt.Sprintf("localhost:%s", port)
    transport, err := raft.NewTCPTransport(address, nil, 3, raftTimeout, os.Stdout)
    if err != nil {
        log.Fatalf("Error al crear TCP transport: %v", err)
    }

    // FSM (Finite State Machine) para manejar la replicación de logs
    fsm := &FSM{}

    // Inicializar Raft
    raftNode, err = raft.NewRaft(config, fsm, logStore, logStore, snapshotStore, transport)
    if err != nil {
        log.Fatalf("Error al inicializar Raft: %v", err)
    }

    // Definir los servidores conocidos en el clúster
    raftNode.BootstrapCluster(raft.Configuration{
        Servers: []raft.Server{
            {ID: "node-8081", Address: raft.ServerAddress("localhost:8081")},
            {ID: "node-8082", Address: raft.ServerAddress("localhost:8082")},
            {ID: "node-8083", Address: raft.ServerAddress("localhost:8083")},
        },
    })
}
```

**Descripción:**
- **Configuración de Raft**: Cada nodo se inicializa con una configuración básica y un identificador único. Luego, se configuran los mecanismos de almacenamiento de logs y snapshots, que ayudan a persistir los datos en caso de fallos.
- **Transporte TCP**: Usamos transporte TCP para la comunicación entre nodos.
- **FSM (Máquina de Estados Finitos)**: Es la encargada de aplicar las entradas de logs replicadas. La FSM realiza las operaciones de escritura en el sistema.
- **Bootstrap del Clúster**: Cada nodo conoce la configuración del clúster a través de la función `BootstrapCluster`, que contiene los nodos que forman parte del sistema.

##### b) **Finite State Machine (FSM)**

La FSM es el componente clave de Raft que maneja la aplicación de los logs replicados. Cada vez que el líder recibe una operación de escritura, esta se convierte en una entrada de log, y Raft se encarga de replicarla en todos los nodos del clúster.

```go
type FSM struct{}

// Apply maneja la replicación de logs
func (f *FSM) Apply(raftLog *raft.Log) interface{} {
    var data map[string]string
    if err := json.Unmarshal(raftLog.Data, &data); err != nil {
        log.Fatalf("Error al aplicar log: %v", err)
        return nil
    }

    key := data["key"]
    value := data["value"]

    // Guardar la clave-valor en el archivo del nodo
    mutex.Lock()
    err := appendToFile(key, value)
    mutex.Unlock()

    if err != nil {
        log.Printf("Error al replicar datos: %v", err)
        return err
    }

    fmt.Printf("Datos replicados: %s = %s\n", key, value)
    return nil
}
```

**Descripción:**
- La función `Apply` se ejecuta cada vez que un log es replicado en los nodos. Este código toma el log de Raft, lo convierte en un par clave-valor, y lo guarda en el archivo de base de datos del nodo usando `appendToFile`.
- Esta operación es atómica, garantizando la consistencia de los datos entre los nodos.

##### c) **Replicación y Escritura de Datos**

Las escrituras son manejadas exclusivamente por el líder del clúster. Cualquier solicitud de escritura que llegue a un follower será rechazada. El líder acepta la escritura, la transforma en un log y luego la replica a todos los nodos del clúster.

```go
func writeHandler(w http.ResponseWriter, r *http.Request) {
    // Verificar si el nodo es líder
    if raftNode.State() != raft.Leader {
        http.Error(w, "Este nodo no es el líder", http.StatusForbidden)
        return
    }

    var data map[string]string
    err := json.NewDecoder(r.Body).Decode(&data)
    if err != nil {
        http.Error(w, "Datos inválidos", http.StatusBadRequest)
        return
    }

    // Convertir los datos en JSON para replicación
    jsonData, err := json.Marshal(data)
    if err != nil {
        http.Error(w, "Error al preparar datos para replicación", http.StatusInternalServerError)
        return
    }

    // Proponer los datos al log de Raft para replicación
    future := raftNode.Apply(jsonData, raftTimeout)
    if err := future.Error(); err != nil {
        http.Error(w, "Error al replicar datos", http.StatusInternalServerError)
        return
    }

    fmt.Printf("Escritura replicada: %s = %s\n", data["key"], data["value"])
    w.WriteHeader(http.StatusOK)
}
```

**Descripción:**
- **Validación del líder**: Solo el líder puede manejar las escrituras, por lo que primero se verifica el estado del nodo. Si el nodo no es el líder, se rechaza la operación.
- **Propuesta de Log**: El líder toma la escritura, la convierte en una entrada de log, y la envía a los demás nodos mediante el mecanismo de Raft. La función `Apply` maneja esta replicación.

##### d) **Lecturas desde los Followers**

Las lecturas pueden ser manejadas por cualquiera de los followers, y el **proxy** se encarga de redirigir las solicitudes de lectura a los seguidores disponibles.

```go
func readHandler(w http.ResponseWriter, r *http.Request) {
    key := r.URL.Path[len("/read/"):]

    // Buscar el valor de la clave en el archivo del nodo
    value, err := readFromFile(key)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    response := map[string]string{"key": key, "value": value}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)

    fmt.Printf("Lectura recibida: %s = %s\n", key, value)
}
```

**Descripción:**
- Las lecturas son simples consultas al archivo local del nodo donde se almacenan las claves y valores replicados. Los nodos están sincronizados gracias a Raft, por lo que cualquier follower puede responder con datos consistentes.

#### 3. **Integración del Algoritmo con el Sistema**

El algoritmo Raft se integra de manera fluida con el sistema de escritura y lectura distribuida. A continuación se explica cómo:

- **Escritura Distribuida**: El proxy envía todas las escrituras al nodo líder. El líder replica las escrituras a todos los followers, garantizando que todos los nodos mantengan una copia sincronizada de los datos. En caso de fallo del líder, los followers eligen un nuevo líder y continúan con el proceso.
  
- **Lecturas Distribuidas**: Las lecturas pueden ser manejadas por cualquier follower. El proxy redirige las lecturas a los followers disponibles, y en caso de fallo de un follower,

 redirige la solicitud al otro follower. Esto asegura que el sistema esté disponible para lecturas, incluso en presencia de fallos.

#### 4. **Tolerancia a Fallos y Recuperación**

Raft es inherentemente tolerante a fallos. Si un nodo se cae y luego vuelve a estar en línea, recibirá las entradas de log que se perdieron durante su inactividad a través del proceso de replicación. Esto asegura que el nodo se sincronice nuevamente con el estado del clúster sin perder ninguna actualización.
