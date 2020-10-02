# Generators
A generator is fundamentally a collection of files represented by directory structure. 
For a directory, to be considered a generator, it only has to contain a configuration 
file `.accio.toml`. Every other file within a generator is a template.

File tree example:
```
generator/
|── some-subdirectory/
│   ├── static-template.txt
├── .accio.toml
├── file.txt
├── blueprint.txt.accio

```

The configuration defines metadata for a generator, like, for example, a help text. 
Moreover, it can define data input that the user has to enter when the generator 
is executed. 

Example:
```
help="This is an example"

[prompts.city]
type=input
message="Enter city name:"
```