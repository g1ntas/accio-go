# Introduction to writing generators

## Blueprints and interactive prompts
To make generators more flexible, we can configure interactive data prompts, which 
allows us to pass custom data to the generator. Then we can use that data to render our 
templates.
 
Let's say we want to create a generator, which will prompt the user to enter his first and 
last names and then creates a text file with the full name in it. 

We can start by creating a config file named `~/generator/.accio.toml` and adding prompts configuration:
```
[prompts.firstName]
type="input"
message="Enter your first name:"

[prompts.lastName]
type="input"
message="Enter your last name:"
```


Here `firstName` and `lastName` represent variables, which we will be able to use in 
templates. For text input, we're using the `input` prompt type, but you can find more types 
in the [configuration reference](reference/configuration.md#prompts).


Once our data is ready, we can create a [blueprint](concepts/templates.md#blueprints)  file 
`~/example/name.txt.accio` that will render our output file:
```
template <<
Hello, {{firstName}} {{lastName}}
>>
```

> Blueprints are like templates, but more abstract, allowing to control 
> more aspects around file generation, like setting custom file names 
> or processing data. They must follow special Accio markup syntax, 
> consisting of tags. To learn more about blueprints,
> [read here](concepts/templates.md#blueprints).

In this file, we added a `template` tag, which renders the content of the output file.
For rendering, Accio uses the Mustache templating engine. We also pass two placeholder 
variables, `firstName` and `lastName`, which will be replaced with actual data when 
the generator is running.
 
And that's it - now we can invoke the generator with `accio run ~/example`:
```
> accio run ~/example
$ Running...
$ Enter your first name:
> John
$ Enter your last name:
> Doe
$ Done.
> cat ~/example/name.txt
$ Hello, John Doe
```

## Data processing

Now when we know how to work with prompts already, what if we want to manipulate 
data and, for example, make entered text uppercase? For that, Accio blueprints 
support `variable` tags that allow us to write scripts with Starlark language.

For a practical example, let's create a blueprint that takes comma-separated 
cities as input and renders them as an HTML list.

For the sake of simplicity, let's say that we already have a configured prompt for 
cities, then what's left is to create a blueprint variable and split cities into 
the list:
```
variable -name="citiesList" <<
  return vars["cities"].split(",")
>>
```

> NOTE: the code within scriptable tags must be indented.

Then in the `template` tag, we can render it as a regular variable:
```
variable -name="citiesList" << ... >>

template <<
<ul>
    {#citiesList}
    <li>{{.}}</li>
    {{/citiesList}}
</ul>
>>
```

Running generator renders nice and clean HTML:
```
> accio run ~/example
$ Running...
> Enter cities:
> Amsterdam,London
$ Done.
> cat ~/example/cities.html
$ <ul>
$     <li>Amsterdam</li>
$     <li>London</li>
$ </ul>
```

You can read more about the `variable` tag in [blueprint reference](reference/blueprints.md#variable).

## Custom filenames

You can customize blueprint filenames with the tag `filename` and just like 
the `variable` tag - it is scriptable:
```
filename << 
  return "directory/custom-file.txt" 
>>
``` 

> NOTE: the filename path should be relative to the generator's root directory.

You can read more about the `filename` tag in [blueprint reference](reference/blueprints.md#filename).


## Skipping files 

Sometimes, if certain conditions apply, you may want not to render the file at all. 
You can achieve this with a scriptable `skipif` tag:
```
skipif <<
  return True
>>
```

Now running the generator will never render the file.

You can read more about the `skipif` tag in [blueprint reference](reference/blueprints.md#skipif).

## Read more

**About core concepts:**
* [Generators](concepts/generators.md)
* [Templates and blueprints](concepts/templates.md)   

**About domain-specific languages:**
* [Accio markup](reference/accio-ml.md)
* [Starlark](reference/starlark.md)
* [Mustache](reference/blueprints.md#mustache)

**Reference:**
* [Configuration](reference/configuration.md)
* [Blueprint tags](reference/blueprints.md#tags)


