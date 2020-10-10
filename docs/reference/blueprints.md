# Blueprints

## Mustache
[Mustache](http://mustache.github.io/) is a logic-less templating engine and is used as a templating engine for blueprints. Documentation can be found at [Mustache(5)](http://mustache.github.io/mustache.5.html).

**NOTE:** lambdas are not supported.

### Context variables
In Mustache templates, context variables can be accessed directly:
```
variable -name="someContextVariable" << "test" >>

template <<
{{someContextVariable}}
>>
```

## Tags
### filename
Specifies the path of the generated file, which is relative to the generator’s root directory. The body must contain Starlark script, which should return a Unix-like path. If the tag is not specified, the relative path of the current file will be used with the `.accio` extension removed. 

Example:
```
# file: generator/abc/file.txt.accio
# result: generator/somefile.txt
filename << "somefile.txt" >>
```

* Directories are supported and will be created automatically if they don't exist yet:
```
# file: generator/file.txt.accio
# result: generator/abc/somefile.txt
filename << "abc/somefile.txt" >>
```

* If the evaluated path is an existing directory, then the file will be generated inside that directory with its original name:
```
# file: generator/abc/file.txt.accio
# existing directory: generator/foo
# result: generator/foo/file.txt
filename << "foo" >>
```

* If path evaluates outside of the generator's root directory, then it will be automatically corrected to the root directory:
```
# file: generator/file.txt.accio
# result: generator/file.txt
filename << "../file.txt" >>
```

### variable
Defines a new variable in the blueprint context. The name attribute specifies the name of the variable. The body expects the Starlark script, which should return `none`, `int`, `string`, `bool`, `float`, `dict`,`tuple`, or `list` value. It can overwrite already existing context variables.

Example:
```
# Returns 10
variable -name="number" <<
    return 5 + 5
>>

# Overwrites previous value and returns 20
variable -name="number" <<
    return vars['number'] + 10 
>>
```

### skipif
Determines if the blueprint file should be skipped - not generated. Expects Starlark script. Returning conditions evaluating to true will skip the file.

Examples:
```
# File won't be generated:
skipif << True >>
skipif << "something" >>
skipif << 1 >>

# File will be generated:
skipif << False >>
skipif << "" >>
skipif << 0 >>
```

### partial
Defines a partial output template, which can be included in other templating tags. The name attribute specifies the name of the partial template. The body expects the Mustache template.

```
partial -name="firstName" <<John>>

# Renders `My name is John`
partial -name="introduction" <<
My name is {{> firstName}}
>>
```

### template
Defines the output content of the generated file. Expects the Mustache template as a body.

```
# Generated file will contain `That's it.`
template <<
That’s it.
>>
```