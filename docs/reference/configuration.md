# Configuration
The configuration uses the TOML markup language, which is compatible with [v0.4.0 specification](https://github.com/toml-lang/toml/blob/master/versions/en/toml-v0.4.0.md).

## help
A documentation of the generator, which will appear when the help command/flag is invoked for the specific generator.

```
help="""
Documentation
"""
```

## ignore
List of paths that should be ignored and not generated. The given paths should be relative to the root directory of 
the generator. If the path is a directory, then all files inside that directory will be ignored as well. Only Unix 
paths are recognized, regardless of the operating system, Accio is running on.

```
ignore=[
  "README.md", # Ignores file `~/generator/README.md`
  "/directory/file.txt", # Ignores file `~/generator/directory/file.txt`
  "some-directory", # Ignores directory `~/generator/some-directory/`
]
```

## prompts
Defines what data will be prompted when a generator is executed. Prompts are defined as nested [tables/maps](https://github.com/toml-lang/toml#user-content-table). The key within the table represents the name of the data entry, which will be used in generator templates. The value should contain a collection of options describing the behavior of the prompt. The order of prompts is not taken into account and may appear in a different order than in the configuration file.

**NOTE:** Prompt key should be no longer than 64 characters, must start with a letter or underscore and consist only of digits, letters, or underscores.

### Prompt options

#### type (required)
Specifies a prompt type that describes how prompt will behave and what kind of input will be accepted.

#### message (required)
A message that will appear when prompted. The message must be no longer than 128 characters.

#### help (optional)
A help text, which can include additional information about the usage of the prompt. This text will be shown next to the prompt. Besides, it will be included in the help reference of the generator.

```
[prompts.age]
type="integer"
message="Enter your age:"
help="""
Some long explanation
on how to use this prompt
"""
```

### Prompt types

#### input
Prompts for any text input.

```
[prompts.name]
type="input"
message="Enter your name:"
```

#### integer
Prompts for integer input.  All signed 64-bit integers are supported (-9223372036854775808 to 9223372036854775807), larger integers will return an error.

```
[prompts.age]
type="integer"
message="Enter your age:"
```

#### confirm
Prompts for confirmation, either yes or no.

```
[prompts.drivingLicense]
type="confirm"
message="Do you have a driving license?"
```

#### choice
Prompts to choose a single option from a pre-defined options list. 

**Additional prompt options:**  
##### options (required)
Defines a list of available choices. Only string values are supported, using any other data type will return an error.  Internally, returns selected string value.

```
[prompts.city]
type="choice"
message="Select city:"
options=["Amsterdam", "Vilnius"]
```

#### multi-choice
Prompts to choose multiple options from a pre-defined options list.

**Additional prompt options:**  
##### options (required)
Defines a list of available choices. Only string values are supported, using any other data type will return an error.  Internally, returns a list of selected string values.

```
[prompts.colors]
type="multi-choice"
message="Select your favorite colors:"
options=["Blue", "Red", "Green", "Yellow", "Black", "White"]
```