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

## Código fuente: 
Implementación del algoritmo de consenso y su integración con los procesos del sistema.

## Documentación técnica: 
Explicación detallada del algoritmo elegido (Raft, Paxos o solución propia), incluyendo los desafíos enfrentados y cómo se garantiza la consistencia y la disponibilidad en el sistema

