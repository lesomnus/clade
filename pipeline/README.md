# Pipeline

Single line expression inspired by pipeline in `text/template`. A pipeline is a sequence of functions separated by "|". Functions can take arguments, and the result of the previous function is passed to the last argument of the next function. The first word of a pipeline element is the name of the function, and the following words become the function's arguments. Pipeline can be nested by wrapping them with '(...)' in argument position.

## Syntax

```ebnf
pipeline = function, { '|', function };
function = name, { { space }-, argument };
name     = text;
argument = text | string | '(', { space }-, pipeline, { space }-, ')';

string = '`', text, '`';
text   = ? all visible characters ? - symbols

symbols = '|' | '`' | '(' | ')'
spaces  = { ' ' }
```

## Implicit Conversions

By default, the given arguments are treated as a string type. However, if the type of the function parameter in the corresponding position is different, "implicit conversion" occurs to the corresponding type. This is also true when the result of a function is passed as an argument to the next element. If an appropriate conversion is not possible, the pipeline returns an error.

## Examples

Consider a function `mul` that takes two integers as arguments and returns the product.

Pipeline "`mul` 1 2 \| `mul` 3" is evaluated as:
1. `mul` 1 2 | `mul` 3
2. `mul` 3 2
3. 6

Pipeline "`mul` 1 (`mul` 2 3) | `mul` 4" is evaluated as:
1. `mul` 1 (`mul` 2 3) | `mul` 4
2. `mul` 1 6 | `mul` 4
3. `mul` 4 6
4. 24
