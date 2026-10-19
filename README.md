# expr

Validador de cadenas para gramáticas independientes del contexto (GIC), escrito en Go.

Lo forman dos partes:

- Una **librería** (`pkg/gic`) que recibe una gramática con producciones en notación extendida (alternativa, cerradura, potencias numéricas y simbólicas) y decide si una cadena pertenece al lenguaje.
- Un **programa de consola** (`cmd`) que usa la librería para resolver la Evidencia de Aprendizaje 2 de Teoría de Autómatas (FCFM, UANL).

## Problema que resuelve el programa

Lenguaje solicitado:

```
L = { i (w)ⁿ i (wᴵ)²ⁿ j² | n ≥ 1 }
```

| Símbolo | Significado |
|---|---|
| `w` | Iniciales de los apellidos |
| `i` | Matrícula completa |
| `wᴵ` | `w` invertida |
| `j` | Primer nombre |

Ejemplo con matrícula `0123456` y nombre `Yozedh Jahday Guerrero Ceja` (`w = gc`, `wᴵ = cg`, `j = yozedh`):

```
0123456gc0123456cgcgyozedhyozedh            n = 1
0123456gcgc0123456cgcgcgcgyozedhyozedh      n = 2
```

## Requisitos

- Go 1.26 o superior.

## Uso del programa

```sh
go run ./cmd
# o
go build -o expr ./cmd && ./expr
```

Al iniciar pide, una sola vez, nombre(s), apellidos y matrícula. Con esos datos construye esta gramática:

```
VN = { S, M, W, I, J }
VT = símbolos que aparecen en matrícula, iniciales y primer nombre
S  → M Wⁿ M I²ⁿ J²
M  → <matrícula>
W  → <iniciales de apellidos>
I  → <iniciales invertidas>
J  → <primer nombre>
```

Después repite el ciclo:

```
Ingrese la cadena a validar: 0123456gc0123456cgcgyozedhyozedh
Valid Expression.
Continue? y/N: n
```

- El nombre y los apellidos se convierten a minúsculas; la cadena a validar no.
- Solo `n` o `N` terminan el programa. Cualquier otra respuesta, incluido Enter vacío, pide otra cadena.
- Mensajes de resultado:
  - `Valid Expression.`
  - `Invalid Expression: Text not accepted: "<cadena>".`
  - `Invalid Expression: Evaluation step limit exceeded (left recursion?) (max 1000).`

### Limitaciones conocidas

- Se acepta `n = 0` (por ejemplo `01234560123456yozedhyozedh`), aunque el enunciado pide `n ≥ 1`. La potencia simbólica prueba valores desde 0.
- El límite de 1000 pasos está fijo en `cmd/main.go`. Con los datos del ejemplo, las cadenas con `n ≥ 18` lo superan aunque sean válidas.

## Estructura

```
cmd/main.go        Programa de consola: captura datos, arma la GIC y ejecuta el ciclo.
pkg/lexer/         Tokeniza el lado derecho de una producción.
pkg/parser/        Construye el AST de una producción (parser de precedencia).
pkg/gic/           Definición de la gramática, compilación y evaluación.
pkg/utils/         Limpieza de pantalla (no se usa actualmente).
```

### Flujo interno

1. `gic.Compile` recorre cada producción, la tokeniza con `lexer`, la convierte en AST con `parser` y valida la gramática (símbolos declarados, producción para el símbolo inicial, variables de potencia).
2. `Program.Eval` simula un autómata de pila no determinista sobre el AST. Explora las alternativas por retroceso, memoriza las configuraciones visitadas y cuenta los pasos hasta `maxSteps`.
3. Las potencias simbólicas (`ⁿ`) se resuelven al vuelo: la primera aparición de una variable fija su valor probando `k = 0, 1, 2, …` y las apariciones siguientes reutilizan ese valor. La cadena se acepta solo si se consumió completa y toda variable fijada se volvió a verificar.

## Uso como librería

```go
import "expr/pkg/gic"

g := &gic.GIC{
    VN: []rune{'S'},
    VT: []rune{'a', 'b'},
    P:  gic.Productions{'S': "a²ⁿbⁿ"},
    S:  'S',
}

prog, err := gic.Compile(g, 10000) // maxSteps
if err != nil {
    // gramática inválida
}

if err := prog.Eval("aaaabb"); err != nil {
    // cadena rechazada
}
```

El módulo se llama `expr`. Para importarlo desde otro módulo, usa una directiva `replace` en el `go.mod` o cambia la ruta del módulo.

### Tipos

| Tipo | Descripción |
|---|---|
| `gic.GIC` | `VN` no terminales, `VT` terminales, `P` producciones, `S` símbolo inicial. |
| `gic.Productions` | `map[rune]string`: no terminal → lado derecho. |
| `gic.Program` | Gramática compilada. `Eval(text string) error` devuelve `nil` si la cadena se acepta. |

### Sintaxis de producciones

Cada símbolo es un solo carácter (rune). Los espacios se ignoran.

| Elemento | Ejemplo | Significado |
|---|---|---|
| Terminal | `a` | Cualquier rune en `VT`. |
| No terminal | `A` | Cualquier rune en `VN`. |
| Concatenación | `aB` | Yuxtaposición. |
| Alternativa | `a \| b` | Una u otra. |
| Cerradura | `a*` | Cero o más. |
| Cerradura positiva | `a+` | Una o más. |
| Vacío | `ε` | Cadena vacía. |
| Agrupación | `(ab)*` | Paréntesis. |
| Potencia numérica | `a³` | Se expande a `aaa`; `⁰` equivale a `ε`. |
| Potencia simbólica | `aⁿ`, `b²ⁿ` | Repetición con variable compartida entre potencias. |

Precedencia, de mayor a menor: operadores posfijos (`*`, `+`, potencias), concatenación, alternativa. Un posfijo afecta solo al elemento anterior: `ab²` es `a(bb)`; para repetir ambos usa `(ab)²`.

Si un carácter está en `VT` y en `VN`, se trata como terminal.

Reglas de las potencias simbólicas:

- Exponentes en superíndice: dígitos `⁰`–`⁹` y letras minúsculas (`ⁿ`, `ᵐ`, …). Los factores se multiplican: `²ⁿ` = 2·n.
- Cada variable debe aparecer en al menos dos potencias de la gramática; si no, `Compile` falla.
- Cada exponente puede tener a lo más una variable todavía sin valor.
- Las variables toman valores desde 0.

### Errores

`pkg/gic`:

| Error | Cuándo |
|---|---|
| `ErrSymbolNotDeclared` | El símbolo inicial o una clave de `P` no está en `VN` (`Compile`). |
| `ErrNoProductionForStart` | No hay producción para `S` (`Compile`). |
| `ErrPowVarSingleAppearance` | Variable de potencia usada una sola vez (`Compile`). |
| `ErrNotAccepted` | La cadena no pertenece al lenguaje (`Eval`). |
| `ErrEvalStepLimit` | Se superó `maxSteps`, por ejemplo con recursión izquierda (`Eval`). |
| `ErrPowUnverified` | Una variable de potencia se fijó y nunca se volvió a verificar (`Eval`). |

`pkg/lexer`: `ErrTokenInvalid` cuando una producción contiene un carácter que no está en `VT` ni en `VN` y no es un operador.

Todos se comparan con `errors.Is`, excepto `ErrSymbolNotDeclared`, que se devuelve formateado sin envolver.

## Pruebas

```sh
go test ./...
```

## Licencia

MIT. Ver [LICENSE](LICENSE).
