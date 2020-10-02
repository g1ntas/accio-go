# Templates
Templates are files that will be created when a generator is executed. There are two kinds of templates - static templates and blueprints. 

### Static templates
Static templates, just as the name implies, are regular static files, they have no distinctive features and are generated as they exist - with a matching relative path and content.

### Blueprints
Blueprints are powerful models that represent templates and can be used to generate files with custom filenames, content composed from user input, or can even evaluate complex logical expressions and decide whether the file should be generated at all. These are files ending with the `.accio` extension (e.g. `file.txt.accio`) and are powered by Accio markup language.

Blueprints are modeled with special tags, which can define variables, customize filename, specify the content, and more. There are two types of tags - script and template. Script tags can only accept Starlark (Python-like language) code and can be utilized to handle user input. Template tags accept Mustache templates, which may be used to compose the content of the generated file.

In the following example, a variable is created and passed to the template tag, which represents the final output that is equal to `9 * 9 = 81`:
```
variable -name="number" <<
    return 9 * 9
>>

template <<
9 * 9 = {{number}}
>>
```

#### Context and variables
Each blueprint has an independent context that can be accessed by tags. The context contains all user prompted data (global variables) plus all local variables defined in the same blueprint.

A global variable can be defined in the configuration file as a prompt:
```
[prompts.city]
type=input
message="Enter city name:"
```

A local variable can be defined in the blueprint with a `variable` tag:
```
variable -name="number" <<
    return 9 * 9
>>
```

Context variables in script tags can be accessed with the `vars` variable. In the template tags, variables can be used directly. The example below shows how variables can be shared between different tags:
```
# here local variable "cityUppercase" is defined, which
# takes a global variable "city" and makes it uppercase.
variable -name="cityUppercase" <<
    return vars['city'].upper()
>>

filename << 
    return vars['cityUppercase'] + ".txt"
>>

# In templates variables can be accessed directly
template << 
City is {{cityUppercase}}
>>
```