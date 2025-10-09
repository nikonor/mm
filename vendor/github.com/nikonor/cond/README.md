# cond

A library for creating logical conditions

## Core Concept

- All conditions are enclosed in parentheses, even the top-level one
- All conditions are built according to the rule: **<command> <first element> [<second element>]**
- All elements are strings, although for some commands the logic involves converting to numbers
- If an element consists of multiple words, it must be enclosed in quotes
- If there are quotes inside the element itself, they must be escaped

## Public Functions

The library provides one public function:

> OK - function for checking a logical condition
>     Input parameters:
>         in - logical string
>         data - variable dictionary
>     Output parameters:
>         bool - condition is met
>         err - error

**IMPORTANT**: Variables in the logical string are specified as **$$variable_name$$**, while in the dictionary they are **variable_name**.

Examples:
```go
    OK("(and (eq $$msisdn$$ 79876543210)  (gt $$age$$ 22))", 
		map[string]string{"msisdn": "79876543210", "age": "22"})
```

In this case, OK will return false. Because the phone numbers match, but age is equal to 22, not greater than 22.

## Command List

### String Commands
- eq - strings are equal
- ne - strings are not equal
- eqi - strings are case-insensitive equal
- contain - first string contains in second string
- icontain - first string case-insensitive contains in second string

### Numeric Commands

For these commands, we convert the given strings into numbers and then compare the numbers

- gt - first element is greater than the second
- lt - first element is less than the second
- gte - first element is greater than or equal to the second
- lte - first element is less than or equal to the second

### Logical Commands

- and - Logical AND
- or - Logical OR
- not - Negation  - This command accepts only 1 element
