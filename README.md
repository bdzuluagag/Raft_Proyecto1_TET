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

### Código fuente: 
Implementación del algoritmo de consenso y su integración con los procesos del sistema.

---

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
