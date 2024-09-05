# Resolución parte 2

## Ejercicio 5

### Protocolo de apuestas
Las apuestas de parte del cliente se modelan con el siguiente struct:

```go
type Bet struct {
	agency uint8 
	firstName string
	lastName string
	document uint32
	birthDate string
	number uint16
}
```

El cliente primero le envia un paquete con la apuesta en cuestion serializada de la siguiente manera:

```
| AGENCY   [1]  | NAME_N [2]     | NAME   [N]   | SURNAME_N [2] | SURNAME[N] | 
| DOCUMENT [4]  | BIRTHDATE [10] | NUMBER [2]   |
```

Primero se manda el numero de agencia, luego se manda el largo del nombre junto con el mismo. Se repite para el appelido para finalmente enviar el documento, la fecha de nacimiento y el numero del sorteo. 

Entre corchetes se encuentra el tamaño en bytes de cada uno de los campos: ex. agencia ocupa un byte. 

Luego en respuesta el servidor le envia un 0 representando que se guardo la apuesta correctamente o un 1 en caso de que hubo un error. 

Finalmente el cliente lee la respuesta del servidor y procede a seguir mandando apuestas. 

## Ejercicio 6
### Protocolo de batch
Siguiendo el formato de como se envia una apuesta, el paquete se arma primero con el byte del tamaño del batch y luego con las bets quedando de la siguiente manera:

```
| N_BETS [2] | BET_1 | BET_2 | BET_3 | ... | BET_N |
```

El servidor entonces lee primero los dos bytes de la cantidad de bets y va leyendo la cantidad de bets recibidas. Si algun paquete es mas grande que 8kb se genera una excepción en el cliente que se estan mandando batches muy grandes. 

Una vez terminado de leer ese batch el servidor responde de igual manera que en la parte 5. 


## Ejercicio 7
### Protocolo sorteo
Cuando el cliente termina de mandar las apuestas, envia lo siguiente:
```
| 0 [2] | AGENCY[1] |
```
Manda un cero (2 bytes para q sea compatible con N_BETS del batch) indicando que no hay mas bets para mandar para luego mandar su agency. 

Luego el servidor se guarda los clientes que terminaron para realizar el sorteo. Al finalizar el mismo le manda a cada una de las agencias/clientes los ganadores de sus correspondientes agencias con el siguiente formato:

```
N_GANADORES | DOCUMENT_1[4] | DOCUMENT_2[4] | ... | DOCUMENT_N |
```

Finalmente cada cliente le manda un 1 para indicar que recibio los resultados para que servidor termine. 

# Resolución parte 3
## Ejercicio 8

Utilice una pool de procesos de parte del servidor en la que se mantiene una pool del tamaño de las agencias. Estos procesos van a ir recibiendo y almacenando las bets. 

Para evitar la RC en el archivo de las bets se utiliza un lock que se comparten entre los archivos a traves del manager de multiprocessing. 

Luego cuando el handler detecta que le llego un cliente que termino de mandar (mediante el protocolo del ejercicio 7) se queda esperando a los demas procesos utilizando una barrera (tmb compartida por el manager)

Al llegar todos los procesos a la barrera se leen las apuestas hechas, y se realiza el sorteo en cada proceso. Al sincronizar los procesos mediante la barrera no hay RC al leer el archivo de las apuestas (debido a que nadie esta escribiendo en ese momento)

Cabe aclarar que a diferencia del ejercicio 7, el servidor no termina por su cuenta sino que espera el SIGTERM/SIGINT para terminar y cerrar de manera limpia. 