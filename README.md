# SimpleFlags Evaluation Engine

---

Simple flags which introduce variables, expressions and rule engine

## Overview

---
This is experimental project, do not use it in production!!!

1. No Variations
2. No Clauses
3. No configuration kind (bool, number ...)
4. No states (on, off)
5. No segments (target groups)
6. No private attributes or anonymous target

### 1. No variations

More than 80% of flags are bool flags and there is no reason to have just two values as variation.
Many companies use variation "true" and "false" but on my engine true and false are reserved keywords.
Following JSON and swagger spec engine use data types from specification. If we talk about multivariate
flags like strings, number etc. than probably there is a reason to have variation but if we look into flags 
there is no variations sharing between flags, so I don't see a reason to have variations at all.

### 2. No custom rules and clauses

Clauses are hard to read and to maintain so IMO I think there are better ways like rule engine evaluations.
instead of maintaining schema for complex clauses better use expression language.
```
dev:
    value: true
    expression: target.identifier in beta
```

### 3. No configuration kind (bool, number ...)

No reason to have flag/configuration kind and engine supports dynamic nature of flags, so if flag serves bool than the flag is boolean.
One thing why engine is different from others it can serve any values as you wish:

- Bool
- Number
- String
- Map
- Slice

```go
evaluate("some_flag", target).Bool(defaultValue)
```

if `some_flag` serves string values 'true' or 'false' you can call bool or string methods it will give same results.

when Generics are fully implemented into Golang then signature will be:
```go
type types interface {
	~bool | ~string | ~float64 | ~map[string]any | []any
}
evaluate[T types]("some_flag", target, true) T
```

### 4. No states (on, off)

Instead of "on" or "off" there is simple bool field on: bool, if flag is on then 'on' will be true.

### 5. No segments (target groups)

This is the most interesting part there are no target groups or segments. So we are introducing concept of variables
why variables? Because feature flags are remote "if" or online "if" and if is part of the language
so variables are more suitable. Variable can have any data type like bool, string, number, slice, map.
Variables can be used as serving values and values used in rule engine. Variables can be global and local.
Local variables are used in the project and globals outside the project. Imagine if you have paid customers, and you want to share among different
projects then you can specify global variable paidCustomers.

```
paidCustomers = [
    'enver', 'bisevac'
]
```

and in flag rule expr:
```
target.name in paidCustomers
```

Global variables and local variables can have the same name.

### 6. No private attributes or anonymous target

All targets are private, so they are never stored and doesn't require any required field, you can put any property.

## Draft design

Configuration structure:

* `project` is project identifier in your system
* `environment` can be one of your environments like (dev, prod, stage, stage1)
* `identifier` is unique configuration (flag) identifier
* `deprecated` if true, than this configuration should be replaced with new configuration
* `on` if true configuration is active otherwise it will serve always `off_value`
* `off_value` when `on` is false it will serve always this value
* `prerequisites` check first dependencies on flag
* `rules` array of two fields `expression` and `value`. Value can be one of the types: `string`, `number`, `object (JSON object)`, `array`, `boolean`, `null`. Expressions will be explained in next section.
* `version` configuration version

Basic sample configuration:

```json
{
  "project": "demo",
  "environment": "dev",
  "identifier": "bool-flag",
  "deprecated": false,
  "on": true,
  "off_value": false,
  "prerequisites": [],
  "rules": [
    {
      "value": true,
      "expression": ""
    }
  ],
  "version": 1
}
```

As you can see this is very simple configuration which will server just one of the values `true` or `false`. Note: true and false are reserved keywords.

Basic number flag configuration:
```json
{
  "project": "demo",
  "environment": "dev",
  "identifier": "number-flag",
  "deprecated": false,
  "on": true,
  "off_value": 1,
  "prerequisites": [],
  "rules": [
    {
      "value": 5,
      "expression": ""
    }
  ],
  "version": 1
}
```

Basic number flag configuration using variables:
```json
{
  "project": "demo",
  "environment": "dev",
  "identifier": "number-flag",
  "deprecated": false,
  "on": true,
  "off_value": "${one}",
  "prerequisites": [],
  "rules": [
    {
      "value": "${five}",
      "expression": ""
    }
  ],
  "version": 1
}
```

Basic string flag configuration:
```json
{
  "project": "demo",
  "environment": "dev",
  "identifier": "string-flag",
  "deprecated": false,
  "on": true,
  "off_value": "item two",
  "prerequisites": [],
  "rules": [
    {
      "value": "item one",
      "expression": ""
    }
  ],
  "version": 1
}
```

Basic array flag configuration:
```json
{
  "project": "demo",
  "environment": "dev",
  "identifier": "slice-flag",
  "deprecated": false,
  "on": true,
  "off_value": [],
  "prerequisites": [],
  "rules": [
    {
      "value": ["item one", "item two", "item three","${success}"],
      "expression": ""
    }
  ],
  "version": 1
}
```

Basic object flag configuration:
```json
{
  "project": "demo",
  "environment": "dev",
  "identifier": "number-flag",
  "deprecated": false,
  "on": true,
  "off_value": {},
  "prerequisites": [],
  "rules": [
    {
      "value": {
        "os": "linux",
        "distro": "arch"
      },
      "expression": ""
    }
  ],
  "version": 1
}
```

Prerequisites flag configuration:
```json
{
  "project": "demo",
  "environment": "dev",
  "identifier": "number-flag",
  "deprecated": false,
  "on": true,
  "off_value": {},
  "prerequisites": [
    {
      "identifier": "bool-flag",
      "value": true
    }
  ],
  "rules": [
    {
      "value": {
        "os": "linux",
        "distro": "arch"
      },
      "expression": ""
    }
  ],
  "version": 1
}
```

Rule based flag configuration:

it will serve true only if target identifier is equal to 'enver' otherwise it will be false
```json
{
  "project": "demo",
  "environment": "dev",
  "identifier": "bool-flag",
  "deprecated": false,
  "on": true,
  "off_value": false,
  "prerequisites": [],
  "rules": [
    {
      "value": true,
      "expression": "target.identifier == 'enver'"
    }
  ],
  "version": 1
}
```

Percentage rollout flag configuration:

```json
{
  "project": "demo",
  "environment": "dev",
  "identifier": "bool-flag",
  "deprecated": false,
  "on": true,
  "off_value": false,
  "prerequisites": [],
  "rules": [
    {
      "value": [
        {
          "__value__": true,
          "__weight__": 50
        },
        {
          "__value__": false,
          "__weight__": 50
        }
      ],
      "expression": "target.identifier in beta_users"
    }
  ],
  "version": 1
}
```

Scheduled configurations
```json
{
  "project": "demo",
  "environment": "dev",
  "identifier": "bool-flag",
  "deprecated": false,
  "on": true,
  "off_value": false,
  "prerequisites": [],
  "rules": [
    {
      "value": true,
      "expression": "target.identifier in paid_customers and now() >= date('2022-10-01')"
    }
  ],
  "version": 1
}
```

another example of scheduled flags:
```json
{
  "project": "demo",
  "environment": "dev",
  "identifier": "bool-flag",
  "deprecated": false,
  "on": true,
  "off_value": false,
  "prerequisites": [],
  "rules": [
    {
      "value": [
        {
          "__value__": true,
          "__weight__": 50
        },
        {
          "__value__": false,
          "__weight__": 50
        }
      ],
      "expression": "target.identifier in beta_users and now() >= date('2022-10-01')"
    }
  ],
  "version": 1
}
```