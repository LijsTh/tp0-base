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


