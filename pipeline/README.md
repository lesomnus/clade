# Pipeline

Single line expression inspired by pipeline in `text/template`. It is chained sequence of commands that can accepts arguments. Each command is may be chained by character '|', and the result of the previous command is passed to the next command. Pipeline can be nested by wrapping them with '(...)' in parameter possiotns.

## Examples

Consider that the `mul` command accepts two integers and returns the product.

Pipeline "`mul` 1 2 \| `mul` 3" is evaluated as:
1. `mul` 1 2 | `mul` 3
1. `mul` 3 2
2. 6

Pipeline "`mul` 1 (`mul` 2 3) | `mul` 4" is evaluated as:
1. `mul` 1 (`mul` 2 3) | `mul` 4
2. `mul` 1 6 | `mul` 4
3. `mul` 4 6
4. 24
